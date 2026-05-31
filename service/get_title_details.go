package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"titles-mcp/config"
)

const (
	tmdbBaseURL = "https://api.themoviedb.org/3"

	scoreExactMatch    = 1000
	scorePartialMatch  = 100
	scoreYearMatch     = 500
	maxPopularityBonus = 100
)

// --- Input / Output DTOs ---

type GetTitleDetailsInput struct {
	MovieName   string  `json:"movie_name" jsonschema:"The movie title to search for on TMDB"`
	ReleaseYear *string `json:"release_year,omitempty" jsonschema:"Optional release year to narrow search"`
}

type GetTitleDetailsOutput struct {
	Status   string        `json:"status"`
	Message  string        `json:"message,omitempty"`
	Metadata *TMDBMetadata `json:"metadata,omitempty"`
}

// --- TMDB Domain Types ---

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExternalIDs struct {
	IMDbID string `json:"imdb_id"`
}

type TMDBMetadata struct {
	ID             int          `json:"id"`
	Title          string       `json:"title"`          // Movie
	Name           string       `json:"name"`           // TV
	IMDbID         string       `json:"imdb_id"`        // Movie (direct), TV (via external_ids)
	ReleaseDate    string       `json:"release_date"`   // Movie
	FirstAirDate   string       `json:"first_air_date"` // TV
	Overview       string       `json:"overview"`
	PosterPath     string       `json:"poster_path"`
	BackdropPath   string       `json:"backdrop_path"`
	VoteAverage    float32      `json:"vote_average"`
	Genres         []Genre      `json:"genres"`
	Runtime        int          `json:"runtime"`          // Movie
	EpisodeRunTime []int        `json:"episode_run_time"` // TV
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

// GetIMDbID returns the IMDb ID, checking both the direct field (movies)
// and the external_ids sub-resource (TV shows).
func (m *TMDBMetadata) GetIMDbID() string {
	if m.IMDbID != "" {
		return m.IMDbID
	}
	if m.ExternalIDs != nil {
		return m.ExternalIDs.IMDbID
	}
	return ""
}

// NormalizeIMDbID promotes the IMDb ID from ExternalIDs into the top-level
// field so callers always find it in one place.
func (m *TMDBMetadata) NormalizeIMDbID() {
	if m.IMDbID == "" {
		m.IMDbID = m.GetIMDbID()
	}
}

// --- Search internals ---

type tmdbSearchResult struct {
	ID           int     `json:"id"`
	MediaType    string  `json:"media_type"`
	Title        string  `json:"title"`
	Name         string  `json:"name"`
	ReleaseDate  string  `json:"release_date"`
	FirstAirDate string  `json:"first_air_date"`
	Popularity   float64 `json:"popularity"`
}

func (r *tmdbSearchResult) displayName() string {
	if r.Title != "" {
		return r.Title
	}
	return r.Name
}

func (r *tmdbSearchResult) displayDate() string {
	if r.ReleaseDate != "" {
		return r.ReleaseDate
	}
	return r.FirstAirDate
}

func (r *tmdbSearchResult) isMediaTitle() bool {
	return r.MediaType == "movie" || r.MediaType == "tv"
}

type bestMatch struct {
	ID        int
	MediaType string
	Score     int
}

// --- Tool handler ---

func (t *titleService) GetTitleDetails(ctx context.Context, input GetTitleDetailsInput) (GetTitleDetailsOutput, error) {
	apiKey := config.AppConfig.TMDBConfig.APIKey
	if apiKey == "" {
		return GetTitleDetailsOutput{
			Status:  "error",
			Message: "TMDB_API_KEY not found in environment. Please add it to your .env file to use this tool.",
		}, nil
	}

	match, err := t.searchTMDB(apiKey, input.MovieName, input.ReleaseYear)
	if err != nil {
		return GetTitleDetailsOutput{Status: "error", Message: err.Error()}, nil
	}

	metadata, err := t.fetchTMDBDetails(apiKey, match.ID, match.MediaType)
	if err != nil {
		return GetTitleDetailsOutput{Status: "error", Message: err.Error()}, nil
	}

	metadata.NormalizeIMDbID()

	return GetTitleDetailsOutput{
		Status:   "success",
		Message:  fmt.Sprintf("Metadata retrieved successfully for %s (Score: %d)", metadata.GetDisplayName(), match.Score),
		Metadata: metadata,
	}, nil
}

// --- TMDB API helpers ---

func (t *titleService) searchTMDB(apiKey, query string, releaseYear *string) (bestMatch, error) {
	searchURL := fmt.Sprintf("%s/search/multi?api_key=%s&query=%s", tmdbBaseURL, apiKey, url.QueryEscape(query))

	var response struct {
		Results []tmdbSearchResult `json:"results"`
	}

	if err := t.getAndDecodeJSON(searchURL, &response); err != nil {
		return bestMatch{}, fmt.Errorf("failed to search TMDB: %w", err)
	}

	candidates := filterMediaTitles(response.Results)
	if len(candidates) == 0 {
		return bestMatch{}, fmt.Errorf("no movie or TV show found for query %q", query)
	}

	return selectBestMatch(candidates, query, releaseYear), nil
}

func filterMediaTitles(results []tmdbSearchResult) []tmdbSearchResult {
	filtered := make([]tmdbSearchResult, 0, len(results))
	for _, r := range results {
		if r.isMediaTitle() {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func selectBestMatch(candidates []tmdbSearchResult, query string, releaseYear *string) bestMatch {
	best := bestMatch{
		ID:        candidates[0].ID,
		MediaType: candidates[0].MediaType,
		Score:     0,
	}

	for _, candidate := range candidates {
		score := scoreCandidate(candidate, query, releaseYear)
		if score > best.Score {
			best = bestMatch{ID: candidate.ID, MediaType: candidate.MediaType, Score: score}
		}
	}

	return best
}

func scoreCandidate(candidate tmdbSearchResult, query string, releaseYear *string) int {
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

func (t *titleService) fetchTMDBDetails(apiKey string, tmdbID int, mediaType string) (*TMDBMetadata, error) {
	detailsURL := fmt.Sprintf("%s/%s/%d?api_key=%s&append_to_response=external_ids", tmdbBaseURL, mediaType, tmdbID, apiKey)

	var metadata TMDBMetadata
	if err := t.getAndDecodeJSON(detailsURL, &metadata); err != nil {
		return nil, fmt.Errorf("failed to fetch details for %s/%d: %w", mediaType, tmdbID, err)
	}

	return &metadata, nil
}

func (t *titleService) getAndDecodeJSON(requestURL string, target any) error {
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
