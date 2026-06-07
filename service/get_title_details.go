package service

import (
	"context"
	"fmt"
	"titles-mcp/clients"
	"titles-mcp/config"
)

// --- Input / Output DTOs ---

type GetTitleDetailsInput struct {
	MovieName   string  `json:"movie_name" jsonschema:"The movie title to search for on TMDB"`
	ReleaseYear *string `json:"release_year,omitempty" jsonschema:"Optional release year to narrow search"`
}

type GetTitleDetailsOutput struct {
	Status   string                `json:"status"`
	Message  string                `json:"message,omitempty"`
	Metadata *clients.TMDBMetadata `json:"metadata,omitempty"`
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

	match, err := t.tmdbClient.SearchTMDB(input.MovieName, input.ReleaseYear)
	if err != nil {
		return GetTitleDetailsOutput{Status: "error", Message: err.Error()}, nil
	}

	metadata, err := t.tmdbClient.FetchTMDBDetails(match.ID, match.MediaType)
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
