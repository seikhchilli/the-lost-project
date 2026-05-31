package service

import (
	"context"
)

type MarkTitleAsWishedInput struct {
	TitleId uint `json:"title_id"`
}

type MarkTitleAsWishedOutput struct {
	Status string `json:"status"`
}

func (t *titleService) MarkTitleAsWished(ctx context.Context, input MarkTitleAsWishedInput) (MarkTitleAsWishedOutput, error) {
	updates := map[string]interface{}{"wished": true, "watched": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return MarkTitleAsWishedOutput{Status: "error"}, err
	}
	return MarkTitleAsWishedOutput{Status: "success"}, nil
}
