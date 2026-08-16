package clients

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	ytsBaseURL     = "https://www13.yts-official.to/"
	ytsSearchDelay = 2 * time.Second
)

// TorrentLink represents a single torrent/magnet link with its quality label.
type TorrentLink struct {
	Title string `json:"title"`
	Href  string `json:"href"`
}

// YTSResult holds the scraping output for a movie lookup.
type YTSResult struct {
	MovieTitle      string        `json:"movie_title"`
	PageURL         string        `json:"page_url"`
	BestQualityLink string        `json:"best_quality_link"`
	AllLinks        []TorrentLink `json:"all_links"`
}

// YTS defines the interface for the YTS scraper client.
type YTS interface {
	GetDownloadLink(movieName string, releaseYear string) (*YTSResult, error)
}

type yts struct{}

// NewYTS creates a new YTS scraper client.
func NewYTS() YTS {
	return &yts{}
}

func (y *yts) GetDownloadLink(movieName string, releaseYear string) (*YTSResult, error) {
	slog.Info("YTS: Starting download link scrape", "movie", movieName, "year", releaseYear)

	// Launch a headless browser
	u, err := launcher.New().Headless(true).Launch()
	if err != nil {
		return nil, fmt.Errorf("failed to launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to browser: %w", err)
	}
	defer browser.MustClose()

	page, err := browser.Page(proto.TargetCreateTarget{URL: ytsBaseURL})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to YTS: %w", err)
	}

	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for page load: %w", err)
	}

	// Build search query: append release year for better accuracy
	searchQuery := movieName
	if releaseYear != "" {
		searchQuery = fmt.Sprintf("%s %s", movieName, releaseYear)
	}

	// Find the search input and type the query
	slog.Info("YTS: Searching", "query", searchQuery)
	searchBox, err := page.Element(`input[type="search"], input[name="keyword"], #quick-search-input`)
	if err != nil {
		return nil, fmt.Errorf("failed to find search box: %w", err)
	}

	if err := searchBox.Input(searchQuery); err != nil {
		return nil, fmt.Errorf("failed to type search query: %w", err)
	}

	// Press Enter to search
	ka, err := searchBox.KeyActions()
	if err != nil {
		return nil, fmt.Errorf("failed to create key actions: %w", err)
	}
	if err := ka.Type(input.Enter).Do(); err != nil {
		return nil, fmt.Errorf("failed to press Enter: %w", err)
	}

	// Wait for search results to load
	time.Sleep(ytsSearchDelay)
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for search results: %w", err)
	}

	// Find and click the exact movie title match
	slog.Info("YTS: Looking for exact title match", "movie", movieName)
	clicked, err := y.clickExactMatch(page, movieName)
	if err != nil {
		return nil, fmt.Errorf("failed to find movie links: %w", err)
	}
	if !clicked {
		return nil, fmt.Errorf("could not find exact match for %q on YTS", movieName)
	}

	// Wait for movie detail page
	time.Sleep(ytsSearchDelay)
	if err := page.WaitLoad(); err != nil {
		return nil, fmt.Errorf("failed to wait for movie page: %w", err)
	}

	pageURL := page.MustInfo().URL
	slog.Info("YTS: On movie page", "url", pageURL)

	// Extract torrent/magnet links
	torrentLinks, err := y.extractTorrentLinks(page)
	if err != nil {
		return nil, fmt.Errorf("failed to extract torrent links: %w", err)
	}

	bestLink := selectBestQualityLink(torrentLinks)

	result := &YTSResult{
		MovieTitle:      movieName,
		PageURL:         pageURL,
		BestQualityLink: bestLink,
		AllLinks:        torrentLinks,
	}

	slog.Info("YTS: Scrape complete", "movie", movieName, "links_found", len(torrentLinks), "best_link", bestLink != "")
	return result, nil
}

// clickExactMatch iterates through links/headings on the search results page
// and clicks the one that exactly matches the movie name (case-insensitive).
func (y *yts) clickExactMatch(page *rod.Page, movieName string) (bool, error) {
	elements, err := page.Elements("a, h2, h3, .movie-title, .title")
	if err != nil {
		return false, err
	}

	targetLower := strings.ToLower(strings.TrimSpace(movieName))

	for _, el := range elements {
		text, err := el.Text()
		if err != nil {
			continue
		}
		text = strings.TrimSpace(text)
		if strings.ToLower(text) == targetLower {
			slog.Info("YTS: Found exact match, clicking", "text", text)
			if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
				slog.Warn("YTS: Failed to click element", "error", err)
				continue
			}
			return true, nil
		}
	}

	return false, nil
}

// extractTorrentLinks finds all torrent and magnet links on the movie detail page.
func (y *yts) extractTorrentLinks(page *rod.Page) ([]TorrentLink, error) {
	allAnchors, err := page.Elements("a")
	if err != nil {
		return nil, err
	}

	var links []TorrentLink
	for _, a := range allAnchors {
		href, err := a.Attribute("href")
		if err != nil || href == nil || *href == "" {
			continue
		}

		hrefVal := *href
		if !strings.Contains(hrefVal, "torrent") && !strings.Contains(hrefVal, "magnet") {
			continue
		}

		title, _ := a.Attribute("title")
		var label string
		if title != nil && *title != "" {
			label = strings.TrimSpace(*title)
		} else {
			text, _ := a.Text()
			label = strings.TrimSpace(text)
		}

		links = append(links, TorrentLink{
			Title: label,
			Href:  hrefVal,
		})
	}

	return links, nil
}

// selectBestQualityLink picks the highest quality link from the list,
// preferring 2160p > 4K > 1080p > 720p, falling back to the first link.
func selectBestQualityLink(links []TorrentLink) string {
	qualityOrder := []string{"2160p", "4K", "1080p", "720p"}

	for _, q := range qualityOrder {
		for _, link := range links {
			if strings.Contains(strings.ToLower(link.Title), strings.ToLower(q)) {
				return link.Href
			}
		}
	}

	// Fallback to first available
	if len(links) > 0 {
		return links[0].Href
	}

	return ""
}
