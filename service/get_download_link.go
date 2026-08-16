package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// --- Input / Output DTOs ---

type TorrentLink struct {
	Title string `json:"title"`
	Href  string `json:"href"`
}

type GetDownloadLinkInput struct {
	MovieName   string `json:"movie_name" jsonschema:"The name of the movie to search for on YTS"`
	ReleaseYear string `json:"release_year,omitempty" jsonschema:"The release year of the movie to narrow the search"`
}

type GetDownloadLinkOutput struct {
	Status          string        `json:"status"`
	Message         string        `json:"message,omitempty"`
	MovieTitle      string        `json:"movie_title,omitempty"`
	PageURL         string        `json:"page_url,omitempty"`
	BestQualityLink string        `json:"best_quality_link,omitempty"`
	AllLinks        []TorrentLink `json:"all_links,omitempty"`
}

type YTSResult struct {
	Title           string        `json:"title"`
	URL             string        `json:"url"`
	BestQualityLink string        `json:"best_quality_link"`
	AllLinks        []TorrentLink `json:"all_links"`
}

// --- Service handler ---

func (t *titleService) GetDownloadLink(ctx context.Context, input GetDownloadLinkInput) (GetDownloadLinkOutput, error) {
	if input.MovieName == "" {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: "movie_name is required",
		}, nil
	}

	searchQuery := input.MovieName
	if input.ReleaseYear != "" {
		searchQuery = fmt.Sprintf("%s %s", input.MovieName, input.ReleaseYear)
	}

	cmd := exec.CommandContext(ctx, "python", "yts_scraper.py", searchQuery, "--match", input.MovieName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to run python script: %v, output: %s", err, string(out)),
		}, nil
	}

	data, err := os.ReadFile("movie_links.json")
	if err != nil {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to read movie_links.json: %v", err),
		}, nil
	}

	var result YTSResult
	if err := json.Unmarshal(data, &result); err != nil {
		return GetDownloadLinkOutput{
			Status:  "error",
			Message: fmt.Sprintf("failed to parse json: %v", err),
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
		Message:         "Download links found for " + result.Title,
		MovieTitle:      result.Title,
		PageURL:         result.URL,
		BestQualityLink: result.BestQualityLink,
		AllLinks:        result.AllLinks,
	}, nil
}
