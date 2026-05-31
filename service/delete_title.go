package service

import (
	"context"
)

type DeleteTitleInput struct {
	TitleId uint `json:"title_id"`
}

type DeleteTitleOutput struct {
	Status  string `json:"status" jsonschema:"status of operation"`
	Message string `json:"message" jsonschema:"message string"`
}

func (t *titleService) DeleteTitle(ctx context.Context, input DeleteTitleInput) (DeleteTitleOutput, error) {
	err := t.repository.DeleteTitle(ctx, input.TitleId)
	if err != nil {
		return DeleteTitleOutput{Status: "error", Message: "Failed to delete title"}, err
	}
	return DeleteTitleOutput{Status: "success", Message: "Title deleted successfully"}, nil
}
