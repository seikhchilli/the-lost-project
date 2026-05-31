package service

import (
	"context"
	"titles-mcp/database"
	"titles-mcp/database/models"
)

type SearchTitlesInput struct {
	TitleNames       *[]string                  `json:"title_names,omitempty" jsonschema:"List of title names to filter by"`
	ReleaseYearRange *database.ReleaseYearRange `json:"release_year_range,omitempty" jsonschema:"Range of release years"`
	Watched          *bool                      `json:"watched,omitempty" jsonschema:"Filter by watched status"`
	Wished           *bool                      `json:"wished,omitempty" jsonschema:"Filter by wished status"`
	Page             int                        `json:"page,omitempty" jsonschema:"Page number"`
	PageSize         int                        `json:"page_size,omitempty" jsonschema:"Number of items per page"`
}

type SearchTitlesOutput struct {
	Status string                `json:"status"`
	Titles []models.TitleSummary `json:"titles"`
	Total  int64                 `json:"total"`
}

func (t *titleService) SearchTitles(ctx context.Context, input SearchTitlesInput) (SearchTitlesOutput, error) {
	summaries, total, err := t.repository.SearchTitles(ctx, database.SearchParams{
		TitleNames:       input.TitleNames,
		ReleaseYearRange: input.ReleaseYearRange,
		Watched:          input.Watched,
		Wished:           input.Wished,
	}, input.Page, input.PageSize)
	if err != nil {
		return SearchTitlesOutput{Status: "error"}, err
	}

	return SearchTitlesOutput{Status: "success", Titles: summaries, Total: total}, nil
}
