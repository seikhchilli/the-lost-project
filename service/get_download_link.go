package service

import (
	"context"
	"titles-mcp/clients"
)

// --- Input / Output DTOs ---

type GetDownloadLinkInput struct {
	MovieName   string `json:"movie_name" jsonschema:"The name of the movie to search for on YTS"`
	ReleaseYear string `json:"release_year,omitempty" jsonschema:"The release year of the movie to narrow the search"`
}

type GetDownloadLinkOutput struct {
	Status          string               `json:"status"`
	Message         string               `json:"message,omitempty"`
	MovieTitle      string               `json:"movie_title,omitempty"`
	PageURL         string               `json:"page_url,omitempty"`
	BestQualityLink string               `json:"best_quality_link,omitempty"`
	AllLinks        []clients.TorrentLink `json:"all_links,omitempty"`
}

// --- Service handler ---

func (t *titleService) GetDownloadLink(ctx context.Context, input GetDownloadLinkInput) (GetDownloadLinkOutput, error) {
	if input.MovieName == "" {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: "movie_name is required",
		}, nil
	}

	result, err := t.ytsClient.GetDownloadLink(input.MovieName, input.ReleaseYear)
	if err != nil {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: err.Error(),
		}, nil
	}

	if result.BestQualityLink == "" && len(result.AllLinks) == 0 {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: "No download links found for " + input.MovieName,
		}, nil
	}

	return GetDownloadLinkOutput{
		Status:          "success",
		Message:         "Download links found for " + result.MovieTitle,
		MovieTitle:      result.MovieTitle,
		PageURL:         result.PageURL,
		BestQualityLink: result.BestQualityLink,
		AllLinks:        result.AllLinks,
	}, nil
}
