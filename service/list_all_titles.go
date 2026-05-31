package service

import (
	"context"
	"titles-mcp/database/models"
)

type GetAllTitlesInput struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
}

type GetAllTitlesOutput struct {
	Status string                `json:"status"`
	Titles []models.TitleSummary `json:"titles"`
	Total  int64                 `json:"total"`
}

func (t *titleService) ListAllTitles(ctx context.Context, input GetAllTitlesInput) (GetAllTitlesOutput, error) {
	titles, total, err := t.repository.GetAllTitles(ctx, input.Page, input.PageSize)
	if err != nil {
		return GetAllTitlesOutput{Status: "error", Titles: titles}, err
	}
	return GetAllTitlesOutput{Status: "success", Titles: titles, Total: total}, nil
}
