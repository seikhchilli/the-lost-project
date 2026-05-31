package service

import (
	"context"
	"titles-mcp/database/models"
)

type ListWatchedTitlesInput struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

type ListWatchedTitlesOutput struct {
	Status        string                `json:"status"`
	WatchedTitles []models.TitleSummary `json:"watched_titles"`
	Total         int64                 `json:"total"`
}

func (t *titleService) ListWatchedTitles(ctx context.Context, input ListWatchedTitlesInput) (ListWatchedTitlesOutput, error) {
	watched_titles, total, err := t.repository.GetWatchedTitles(ctx, input.Page, input.PageSize)
	if err != nil {
		return ListWatchedTitlesOutput{Status: "error", WatchedTitles: watched_titles}, err
	}
	return ListWatchedTitlesOutput{Status: "success", WatchedTitles: watched_titles, Total: total}, nil
}
