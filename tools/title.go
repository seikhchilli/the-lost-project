package tools

import (
	"context"
	"titles-mcp/service"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type TitleTool interface {
	Register(server *mcp.Server)
}

type titleTool struct {
	titleService service.TitleService
}

func NewTitleTool(titleService service.TitleService) TitleTool {
	return &titleTool{
		titleService: titleService,
	}
}

func (t *titleTool) Register(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{Name: "add_titles", Description: "Add new movies or TV shows title to the database"}, t.AddTitles)
	mcp.AddTool(server, &mcp.Tool{Name: "mark_title_as_watched", Description: "Mark a title as watched by its ID"}, t.MarkTitleAsWatched)
	mcp.AddTool(server, &mcp.Tool{Name: "list_all_titles", Description: "List all titles with their ID, name, and release year"}, t.ListAllTitles)
	mcp.AddTool(server, &mcp.Tool{Name: "get_titles_by_ids", Description: "Retrieve full title details by their IDs"}, t.GetTitlesByIds)
	mcp.AddTool(server, &mcp.Tool{Name: "mark_title_as_wished", Description: "Mark a title as wished (to-watch) by its ID"}, t.MarkTitleAsWished)
	mcp.AddTool(server, &mcp.Tool{Name: "list_watched_titles", Description: "List all watched titles"}, t.ListWatchedTitles)
	mcp.AddTool(server, &mcp.Tool{Name: "search_titles", Description: "Search for titles with advanced filters (IDs, names, year range, watched/wished status)"}, t.SearchTitles)
	mcp.AddTool(server, &mcp.Tool{Name: "remove_from_watched", Description: "Remove a title from the watched list by its ID"}, t.RemoveTitleFromWatched)
	mcp.AddTool(server, &mcp.Tool{Name: "remove_from_wished", Description: "Remove a title from the wished list by its ID"}, t.RemoveTitleFromWished)
	mcp.AddTool(server, &mcp.Tool{Name: "get_title_details", Description: "Search for movie or TV show details including IMDB ID"}, t.GetTitleDetails)
}

func (t *titleTool) AddTitles(ctx context.Context, req *mcp.CallToolRequest, input service.AddTitlesInput) (
	*mcp.CallToolResult,
	service.AddTitlesOutput,
	error,
) {
	output, err := t.titleService.AddTitles(ctx, input)
	return nil, output, err
}

func (t *titleTool) MarkTitleAsWatched(ctx context.Context, req *mcp.CallToolRequest, input service.MarkTitleAsWatchedInput) (
	*mcp.CallToolResult,
	service.MarkTitleAsWatchedOutput,
	error,
) {
	output, err := t.titleService.MarkTitleAsWatched(ctx, input)
	return nil, output, err
}

func (t *titleTool) ListAllTitles(ctx context.Context, req *mcp.CallToolRequest, input service.GetAllTitlesInput) (
	*mcp.CallToolResult,
	service.GetAllTitlesOutput,
	error,
) {
	output, err := t.titleService.ListAllTitles(ctx, input)
	return nil, output, err
}

func (t *titleTool) GetTitlesByIds(ctx context.Context, req *mcp.CallToolRequest, input service.GetTitlesByIdsInput) (
	*mcp.CallToolResult,
	service.GetTitlesByIdsOutput,
	error,
) {
	output, err := t.titleService.GetTitlesByIds(ctx, input)
	return nil, output, err
}

func (t *titleTool) MarkTitleAsWished(ctx context.Context, req *mcp.CallToolRequest, input service.MarkTitleAsWishedInput) (
	*mcp.CallToolResult,
	service.MarkTitleAsWishedOutput,
	error,
) {
	output, err := t.titleService.MarkTitleAsWished(ctx, input)
	return nil, output, err
}

func (t *titleTool) ListWatchedTitles(ctx context.Context, req *mcp.CallToolRequest, input service.ListWatchedTitlesInput) (
	*mcp.CallToolResult,
	service.ListWatchedTitlesOutput,
	error,
) {
	output, err := t.titleService.ListWatchedTitles(ctx, input)
	return nil, output, err
}

func (t *titleTool) SearchTitles(ctx context.Context, req *mcp.CallToolRequest, input service.SearchTitlesInput) (
	*mcp.CallToolResult,
	service.SearchTitlesOutput,
	error,
) {
	output, err := t.titleService.SearchTitles(ctx, input)
	return nil, output, err
}

func (t *titleTool) RemoveTitleFromWatched(ctx context.Context, req *mcp.CallToolRequest, input service.RemoveTitleFromWatchedInput) (
	*mcp.CallToolResult,
	service.RemoveTitleFromWatchedOutput,
	error,
) {
	output, err := t.titleService.RemoveTitleFromWatched(ctx, input)
	return nil, output, err
}

func (t *titleTool) RemoveTitleFromWished(ctx context.Context, req *mcp.CallToolRequest, input service.RemoveTitleFromWishedInput) (
	*mcp.CallToolResult,
	service.RemoveTitleFromWishedOutput,
	error,
) {
	output, err := t.titleService.RemoveTitleFromWished(ctx, input)
	return nil, output, err
}

func (t *titleTool) GetTitleDetails(ctx context.Context, req *mcp.CallToolRequest, input service.GetTitleDetailsInput) (
	*mcp.CallToolResult,
	service.GetTitleDetailsOutput,
	error,
) {
	output, err := t.titleService.GetTitleDetails(ctx, input)
	return nil, output, err
}
