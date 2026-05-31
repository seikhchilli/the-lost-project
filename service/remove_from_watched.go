package service

import (
	"context"
)

type RemoveTitleFromWatchedInput struct {
	TitleId uint `json:"title_id" jsonschema:"The ID of the title to remove from watched list"`
}

type RemoveTitleFromWatchedOutput struct {
	Status string `json:"status"`
}

func (t *titleService) RemoveTitleFromWatched(ctx context.Context, input RemoveTitleFromWatchedInput) (RemoveTitleFromWatchedOutput, error) {
	updates := map[string]interface{}{"watched": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return RemoveTitleFromWatchedOutput{Status: "error"}, err
	}
	return RemoveTitleFromWatchedOutput{Status: "success"}, nil
}
