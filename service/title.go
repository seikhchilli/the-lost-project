package service

import (
	"context"
	"titles-mcp/clients"
	"titles-mcp/database"
)

type TitleService interface {
	AddTitles(ctx context.Context, input AddTitlesInput) (AddTitlesOutput, error)
	MarkTitleAsWatched(ctx context.Context, input MarkTitleAsWatchedInput) (MarkTitleAsWatchedOutput, error)
	ListAllTitles(ctx context.Context, input GetAllTitlesInput) (GetAllTitlesOutput, error)
	GetTitlesByIds(ctx context.Context, input GetTitlesByIdsInput) (GetTitlesByIdsOutput, error)
	MarkTitleAsWished(ctx context.Context, input MarkTitleAsWishedInput) (MarkTitleAsWishedOutput, error)
	ListWatchedTitles(ctx context.Context, input ListWatchedTitlesInput) (ListWatchedTitlesOutput, error)
	SearchTitles(ctx context.Context, input SearchTitlesInput) (SearchTitlesOutput, error)
	RemoveTitleFromWatched(ctx context.Context, input RemoveTitleFromWatchedInput) (RemoveTitleFromWatchedOutput, error)
	RemoveTitleFromWished(ctx context.Context, input RemoveTitleFromWishedInput) (RemoveTitleFromWishedOutput, error)
	GetTitleDetails(ctx context.Context, input GetTitleDetailsInput) (GetTitleDetailsOutput, error)
	DeleteTitle(ctx context.Context, input DeleteTitleInput) (DeleteTitleOutput, error)
	GetNextGameMovie(ctx context.Context) (NextGameMovie, error)
}

type titleService struct {
	repository database.Repository
	llmClient  clients.LLM
}

func NewTitleService(repository database.Repository) TitleService {
	return &titleService{
		repository: repository,
		llmClient:  clients.NewLLM(context.Background()),
	}
}
