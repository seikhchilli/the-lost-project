package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"titles-mcp/config"
)

type GetNextGameMovieOutput struct {
	Status   string        `json:"status"`
	Metadata *TMDBMetadata `json:"metadata,omitempty"`
	Message  string        `json:"message,omitempty"`
}

type tmdbPopularResponse struct {
	Results []tmdbSearchResult `json:"results"`
}

func (t *titleService) GetNextGameMovie(ctx context.Context) (GetNextGameMovieOutput, error) {
	apiKey := config.AppConfig.TMDBConfig.APIKey
	if apiKey == "" {
		return GetNextGameMovieOutput{
			Status:  "error",
			Message: "TMDB_API_KEY not found in environment.",
		}, nil
	}

	// Fetch a random page from 1 to 50 of popular movies to get variety
	rand.Seed(time.Now().UnixNano())
	page := rand.Intn(50) + 1

	popularURL := fmt.Sprintf("%s/movie/popular?api_key=%s&page=%d", tmdbBaseURL, apiKey, page)

	resp, err := http.Get(popularURL)
	if err != nil {
		return GetNextGameMovieOutput{}, fmt.Errorf("failed to fetch popular movies: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return GetNextGameMovieOutput{}, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	var popResp tmdbPopularResponse
	if err := json.NewDecoder(resp.Body).Decode(&popResp); err != nil {
		return GetNextGameMovieOutput{}, fmt.Errorf("failed to decode TMDB response: %w", err)
	}

	if len(popResp.Results) == 0 {
		return GetNextGameMovieOutput{Status: "error", Message: "No movies found"}, nil
	}

	// Extract TMDB IDs as strings
	var tmdbIds []string
	for _, result := range popResp.Results {
		tmdbIds = append(tmdbIds, fmt.Sprintf("%d", result.ID))
	}

	// Filter out movies already in our DB
	existingIds, err := t.repository.GetExistingTmdbIds(ctx, tmdbIds)
	if err != nil {
		return GetNextGameMovieOutput{}, fmt.Errorf("failed to check existing titles: %w", err)
	}

	existingMap := make(map[string]bool)
	for _, id := range existingIds {
		existingMap[id] = true
	}

	var selectedMovieID int
	// Shuffle results to make it more random even within the page
	rand.Shuffle(len(popResp.Results), func(i, j int) {
		popResp.Results[i], popResp.Results[j] = popResp.Results[j], popResp.Results[i]
	})

	for _, result := range popResp.Results {
		idStr := fmt.Sprintf("%d", result.ID)
		if !existingMap[idStr] {
			selectedMovieID = result.ID
			break
		}
	}

	if selectedMovieID == 0 {
		return GetNextGameMovieOutput{Status: "error", Message: "All movies on this page are already tracked. Please try again."}, nil
	}

	// Fetch full metadata for the selected movie
	metadata, err := t.fetchTMDBDetails(apiKey, selectedMovieID, "movie")
	if err != nil {
		return GetNextGameMovieOutput{}, err
	}

	metadata.NormalizeIMDbID()

	return GetNextGameMovieOutput{
		Status:   "success",
		Metadata: metadata,
	}, nil
}
