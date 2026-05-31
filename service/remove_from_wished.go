package service

import (
	"context"
)

type RemoveTitleFromWishedInput struct {
	TitleId uint `json:"title_id" jsonschema:"The ID of the title to remove from wished list"`
}

type RemoveTitleFromWishedOutput struct {
	Status string `json:"status"`
}

func (t *titleService) RemoveTitleFromWished(ctx context.Context, input RemoveTitleFromWishedInput) (RemoveTitleFromWishedOutput, error) {
	updates := map[string]interface{}{"wished": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return RemoveTitleFromWishedOutput{Status: "error"}, err
	}
	return RemoveTitleFromWishedOutput{Status: "success"}, nil
}
