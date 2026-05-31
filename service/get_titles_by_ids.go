package service

import (
	"context"
	"titles-mcp/database/models"
)

type GetTitlesByIdsInput struct {
	Ids []uint `json:"ids"`
}

type GetTitlesByIdsOutput struct {
	Status string         `json:"status"`
	Titles []models.Title `json:"titles"`
}

func (t *titleService) GetTitlesByIds(ctx context.Context, input GetTitlesByIdsInput) (GetTitlesByIdsOutput, error) {
	titles, err := t.repository.GetTitlesByIds(ctx, input.Ids)
	if err != nil {
		return GetTitlesByIdsOutput{Status: "error", Titles: titles}, err
	}
	return GetTitlesByIdsOutput{Status: "success", Titles: titles}, nil
}
