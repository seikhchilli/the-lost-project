package service

import (
	"context"
)

type MarkTitleAsWatchedInput struct {
	TitleId uint `json:"title_id"`
}

type MarkTitleAsWatchedOutput struct {
	Status string `json:"status"`
}

func (t *titleService) MarkTitleAsWatched(ctx context.Context, input MarkTitleAsWatchedInput) (MarkTitleAsWatchedOutput, error) {
	updates := map[string]interface{}{"watched": true, "wished": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return MarkTitleAsWatchedOutput{Status: "error"}, err
	}
	return MarkTitleAsWatchedOutput{Status: "success"}, nil
}
