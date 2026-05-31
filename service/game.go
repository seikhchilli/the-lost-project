package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
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

	const maxRetries = 5
	const maxPages = 500
	triedPages := make(map[int]bool)

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Pick a random page we haven't tried yet
		page := rand.Intn(maxPages) + 1
		for triedPages[page] {
			page = rand.Intn(maxPages) + 1
		}
		triedPages[page] = true

		popularURL := fmt.Sprintf("%s/movie/popular?api_key=%s&page=%d", tmdbBaseURL, apiKey, page)

		resp, err := http.Get(popularURL)
		if err != nil {
			return GetNextGameMovieOutput{}, fmt.Errorf("failed to fetch popular movies: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			// If a high page number 404s, just try the next attempt
			continue
		}

		var popResp tmdbPopularResponse
		if err := json.NewDecoder(resp.Body).Decode(&popResp); err != nil {
			resp.Body.Close()
			return GetNextGameMovieOutput{}, fmt.Errorf("failed to decode TMDB response: %w", err)
		}
		resp.Body.Close()

		if len(popResp.Results) == 0 {
			continue
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

		// Shuffle results to make it more random even within the page
		rand.Shuffle(len(popResp.Results), func(i, j int) {
			popResp.Results[i], popResp.Results[j] = popResp.Results[j], popResp.Results[i]
		})

		for _, result := range popResp.Results {
			idStr := fmt.Sprintf("%d", result.ID)
			if !existingMap[idStr] {
				// Found an untracked movie — fetch its full metadata
				metadata, err := t.fetchTMDBDetails(apiKey, result.ID, "movie")
				if err != nil {
					return GetNextGameMovieOutput{}, err
				}

				metadata.NormalizeIMDbID()

				return GetNextGameMovieOutput{
					Status:   "success",
					Metadata: metadata,
				}, nil
			}
		}

		// All movies on this page were tracked — loop to try another page
	}

	return GetNextGameMovieOutput{
		Status:  "error",
		Message: "Could not find new movies after several attempts. You may have tracked most popular titles!",
	}, nil
}
