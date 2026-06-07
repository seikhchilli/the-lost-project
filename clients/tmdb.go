package clients

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"titles-mcp/config"
)

const (
	tmdbBaseURL        = "https://api.themoviedb.org/3"
	scoreExactMatch    = 1000
	scorePartialMatch  = 100
	scoreYearMatch     = 500
	maxPopularityBonus = 100
)

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExternalIDs struct {
	IMDbID string `json:"imdb_id"`
}

type TMDBMetadata struct {
	ID             int          `json:"id"`
	Title          string       `json:"title"`
	Name           string       `json:"name"`
	IMDbID         string       `json:"imdb_id"`
	ReleaseDate    string       `json:"release_date"`
	FirstAirDate   string       `json:"first_air_date"`
	Overview       string       `json:"overview"`
	PosterPath     string       `json:"poster_path"`
	BackdropPath   string       `json:"backdrop_path"`
	VoteAverage    float32      `json:"vote_average"`
	Genres         []Genre      `json:"genres"`
	Runtime        int          `json:"runtime"`
	EpisodeRunTime []int        `json:"episode_run_time"`
	Tagline        string       `json:"tagline"`
	ExternalIDs    *ExternalIDs `json:"external_ids,omitempty"`
}

func (m *TMDBMetadata) GetDisplayName() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

func (m *TMDBMetadata) GetDisplayDate() string {
	if m.ReleaseDate != "" {
		return m.ReleaseDate
	}
	return m.FirstAirDate
}

func (m *TMDBMetadata) GetIMDbID() string {
	if m.IMDbID != "" {
		return m.IMDbID
	}
	if m.ExternalIDs != nil {
		return m.ExternalIDs.IMDbID
	}
	return ""
}

func (m *TMDBMetadata) NormalizeIMDbID() {
	if m.IMDbID == "" {
		m.IMDbID = m.GetIMDbID()
	}
}

type TMDBSearchResult struct {
	ID           int     `json:"id"`
	MediaType    string  `json:"media_type"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	Popularity   float64 `json:"popularity"`
}

func (r *TMDBSearchResult) displayName() string {
	if r.Title != "" {
		return r.Title
	}
	return r.Name
}

func (r *TMDBSearchResult) displayDate() string {
	if r.ReleaseDate != "" {
		return r.ReleaseDate
	}
	return r.FirstAirDate
}

func (r *TMDBSearchResult) isMediaTitle() bool {
	return r.MediaType == "movie" || r.MediaType == "tv"
}

type BestMatch struct {
	ID        int
	MediaType string
	Score     int
}

type TMDB interface {
	SearchTMDB(query string, releaseYear *string) (BestMatch, error)
	FetchTMDBDetails(tmdbID int, mediaType string) (*TMDBMetadata, error)
	DiscoverGameMovies(genre int) ([]TMDBSearchResult, error)
}

type tmdb struct {
	apiKey string
}

func NewTMDB() TMDB {
	return &tmdb{
		apiKey: config.AppConfig.TMDBConfig.APIKey,
	}
}

func (t *tmdb) SearchTMDB(query string, releaseYear *string) (BestMatch, error) {
	searchURL := fmt.Sprintf("%s/search/multi?api_key=%s&query=%s", tmdbBaseURL, t.apiKey, url.QueryEscape(query))

	var response struct {
		Results []TMDBSearchResult `json:"results"`
	}

	if err := t.getAndDecodeJSON(searchURL, &response); err != nil {
		return BestMatch{}, fmt.Errorf("failed to search TMDB: %w", err)
	}

	candidates := filterMediaTitles(response.Results)
	if len(candidates) == 0 {
		return BestMatch{}, fmt.Errorf("no movie or TV show found for query %q", query)
	}

	return selectBestMatch(candidates, query, releaseYear), nil
}

func filterMediaTitles(results []TMDBSearchResult) []TMDBSearchResult {
	filtered := make([]TMDBSearchResult, 0, len(results))
	for _, r := range results {
		if r.isMediaTitle() {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func selectBestMatch(candidates []TMDBSearchResult, query string, releaseYear *string) BestMatch {
	best := BestMatch{
		ID:        candidates[0].ID,
		MediaType: candidates[0].MediaType,
		Score:     0,
	}

	for _, candidate := range candidates {
		score := scoreCandidate(candidate, query, releaseYear)
		if score > best.Score {
			best = BestMatch{ID: candidate.ID, MediaType: candidate.MediaType, Score: score}
		}
	}

	return best
}

func scoreCandidate(candidate TMDBSearchResult, query string, releaseYear *string) int {
	score := scoreTitleMatch(candidate.displayName(), query)
	score += scoreYearMatchBonus(candidate.displayDate(), releaseYear)
	score += scorePopularity(candidate.Popularity)
	return score
}

func scoreTitleMatch(title, query string) int {
	if strings.EqualFold(title, query) {
		return scoreExactMatch
	}
	if strings.Contains(strings.ToLower(title), strings.ToLower(query)) {
		return scorePartialMatch
	}
	return 0
}

func scoreYearMatchBonus(date string, releaseYear *string) int {
	if releaseYear != nil && *releaseYear != "" && strings.HasPrefix(date, *releaseYear) {
		return scoreYearMatch
	}
	return 0
}

func scorePopularity(popularity float64) int {
	bonus := int(popularity)
	if bonus > maxPopularityBonus {
		return maxPopularityBonus
	}
	return bonus
}

func (t *tmdb) FetchTMDBDetails(tmdbID int, mediaType string) (*TMDBMetadata, error) {
	detailsURL := fmt.Sprintf("%s/%s/%d?api_key=%s&append_to_response=external_ids", tmdbBaseURL, mediaType, tmdbID, t.apiKey)

	var metadata TMDBMetadata
	if err := t.getAndDecodeJSON(detailsURL, &metadata); err != nil {
		return nil, fmt.Errorf("failed to fetch details for %s/%d: %w", mediaType, tmdbID, err)
	}

	return &metadata, nil
}

func (t *tmdb) getAndDecodeJSON(requestURL string, target any) error {
	resp, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func (t *tmdb) DiscoverGameMovies(genre int) ([]TMDBSearchResult, error) {
	startYear := 1980 + (rand.Intn(5) * 10)
	dateGte := fmt.Sprintf("%d-01-01", startYear)
	dateLte := fmt.Sprintf("%d-12-31", startYear+9)

	page := rand.Intn(2) + 1

	discoverURL := fmt.Sprintf("%s/discover/movie?api_key=%s&page=%d&with_genres=%d&primary_release_date.gte=%s&primary_release_date.lte=%s&vote_count.gte=3000&vote_average.gte=7.0&sort_by=popularity.desc",
		tmdbBaseURL, t.apiKey, page, genre, dateGte, dateLte)

	var resp struct {
		Results []TMDBSearchResult `json:"results"`
	}

	if err := t.getAndDecodeJSON(discoverURL, &resp); err != nil {
		return nil, err
	}

	return resp.Results, nil
}
