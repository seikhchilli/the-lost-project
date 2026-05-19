package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"titles-mcp/config"
	"titles-mcp/database"
	"titles-mcp/database/models"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

type TitleTool interface {
	Register(server *mcp.Server)
	AddTitles(ctx context.Context, req *mcp.CallToolRequest, input AddTitlesInput) (
		*mcp.CallToolResult,
		AddTitlesOutput,
		error,
	)
	MarkTitleAsWatched(ctx context.Context, req *mcp.CallToolRequest, input MarkTitleAsWatchedInput) (
		*mcp.CallToolResult,
		MarkTitleAsWatchedOutput,
		error,
	)
	ListAllTitles(ctx context.Context, req *mcp.CallToolRequest, input GetAllTitlesInput) (
		*mcp.CallToolResult,
		GetAllTitlesOutput,
		error,
	)
	GetTitlesByIds(ctx context.Context, req *mcp.CallToolRequest, input GetTitlesByIdsInput) (
		*mcp.CallToolResult,
		GetTitlesByIdsOutput,
		error,
	)
	MarkTitleAsWished(ctx context.Context, req *mcp.CallToolRequest, input MarkTitleAsWishedInput) (
		*mcp.CallToolResult,
		MarkTitleAsWishedOutput,
		error,
	)
	ListWatchedTitles(ctx context.Context, req *mcp.CallToolRequest, input ListWatchedTitlesInput) (
		*mcp.CallToolResult,
		ListWatchedTitlesOutput,
		error,
	)
	SearchTitles(ctx context.Context, req *mcp.CallToolRequest, input SearchTitlesInput) (
		*mcp.CallToolResult,
		SearchTitlesOutput,
		error,
	)
	RemoveTitleFromWatched(ctx context.Context, req *mcp.CallToolRequest, input RemoveTitleFromWatchedInput) (
		*mcp.CallToolResult,
		RemoveTitleFromWatchedOutput,
		error,
	)
	RemoveTitleFromWished(ctx context.Context, req *mcp.CallToolRequest, input RemoveTitleFromWishedInput) (
		*mcp.CallToolResult,
		RemoveTitleFromWishedOutput,
		error,
	)
	GetTMDBMetadata(ctx context.Context, req *mcp.CallToolRequest, input GetTMDBMetadataInput) (
		*mcp.CallToolResult,
		GetTMDBMetadataOutput,
		error,
	)
}

type titleTool struct {
	repository database.Repository
}

func NewTitleTool(db *gorm.DB) TitleTool {
	return &titleTool{
		repository: database.NewRepository(db),
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
	mcp.AddTool(server, &mcp.Tool{Name: "get_tmdb_metadata", Description: "Search TMDB for movie or TV show metadata"}, t.GetTMDBMetadata)
}

type GetTMDBMetadataInput struct {
	MovieName   string  `json:"movie_name" jsonschema:"The movie title to search for on TMDB"`
	ReleaseYear *string `json:"release_year,omitempty" jsonschema:"Optional release year to narrow search"`
}

type TMDBMetadata struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`        // Movie
	Name         string   `json:"name"`         // TV
	ReleaseDate  string   `json:"release_date"` // Movie
	FirstAirDate string   `json:"first_air_date"` // TV
	Overview     string   `json:"overview"`
	PosterPath   string   `json:"poster_path"`
	BackdropPath string   `json:"backdrop_path"`
	VoteAverage  float32  `json:"vote_average"`
	Genres       []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Runtime        int   `json:"runtime"`          // Movie
	EpisodeRunTime []int `json:"episode_run_time"` // TV
	Tagline        string `json:"tagline"`
}

func (m *TMDBMetadata) GetDisplayName() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

func (m *TMDBMetadata) GetDisplayDate() string {
	if m.ReleaseDate != "" {
		return m.ReleaseDate
	}
	return m.FirstAirDate
}

type GetTMDBMetadataOutput struct {
	Status   string        `json:"status"`
	Message  string        `json:"message,omitempty"`
	Metadata *TMDBMetadata `json:"metadata,omitempty"`
}

func (t *titleTool) GetTMDBMetadata(ctx context.Context, req *mcp.CallToolRequest, input GetTMDBMetadataInput) (
	*mcp.CallToolResult,
	GetTMDBMetadataOutput,
	error,
) {
	apiKey := config.AppConfig.TMDBConfig.APIKey
	if apiKey == "" {
		return nil, GetTMDBMetadataOutput{
			Status:  "error",
			Message: "TMDB_API_KEY not found in environment. Please add it to your .env file to use this tool.",
		}, nil
	}

	tmdbID, mediaType, score, err := t.searchTMDB(apiKey, input.MovieName, input.ReleaseYear)
	if err != nil {
		return nil, GetTMDBMetadataOutput{Status: "error", Message: err.Error()}, nil
	}

	metadata, err := t.fetchTMDBDetails(apiKey, tmdbID, mediaType)
	if err != nil {
		return nil, GetTMDBMetadataOutput{Status: "error", Message: err.Error()}, nil
	}

	return nil, GetTMDBMetadataOutput{
		Status:   "success",
		Message:  fmt.Sprintf("Metadata retrieved successfully for %s (Score: %d)", metadata.GetDisplayName(), score),
		Metadata: metadata,
	}, nil
}

func (t *titleTool) searchTMDB(apiKey, query string, releaseYear *string) (int, string, int, error) {
	searchURL := fmt.Sprintf("https://api.themoviedb.org/3/search/multi?api_key=%s&query=%s", apiKey, url.QueryEscape(query))

	var searchResponse struct {
		Results []struct {
			ID           int     `json:"id"`
			MediaType    string  `json:"media_type"`
			Title        string  `json:"title"`
			Name         string  `json:"name"`
			ReleaseDate  string  `json:"release_date"`
			FirstAirDate string  `json:"first_air_date"`
			Popularity   float64 `json:"popularity"`
		} `json:"results"`
	}

	if err := t.getAndDecodeJSON(searchURL, &searchResponse); err != nil {
		return 0, "", 0, fmt.Errorf("failed to search: %w", err)
	}

	if len(searchResponse.Results) == 0 {
		return 0, "", 0, fmt.Errorf("no results found")
	}

	var bestID int
	var bestType string
	maxScore := -1

	for _, res := range searchResponse.Results {
		if res.MediaType != "movie" && res.MediaType != "tv" {
			continue
		}

		currentTitle := res.Title
		if res.MediaType == "tv" {
			currentTitle = res.Name
		}

		currentDate := res.ReleaseDate
		if res.MediaType == "tv" {
			currentDate = res.FirstAirDate
		}

		score := 0
		if strings.EqualFold(currentTitle, query) {
			score += 1000
		} else if strings.Contains(strings.ToLower(currentTitle), strings.ToLower(query)) {
			score += 100
		}

		if releaseYear != nil && *releaseYear != "" && strings.HasPrefix(currentDate, *releaseYear) {
			score += 500
		}

		// Add a bit of popularity to break ties, but cap its influence
		popScore := int(res.Popularity)
		if popScore > 100 {
			popScore = 100
		}
		score += popScore

		if score > maxScore {
			maxScore = score
			bestID = res.ID
			bestType = res.MediaType
		}
	}

	if maxScore == -1 {
		// Fallback to first movie or tv result
		for _, res := range searchResponse.Results {
			if res.MediaType == "movie" || res.MediaType == "tv" {
				return res.ID, res.MediaType, 0, nil
			}
		}
		return 0, "", 0, fmt.Errorf("no movie or tv show found")
	}

	return bestID, bestType, maxScore, nil
}

func (t *titleTool) fetchTMDBDetails(apiKey string, tmdbID int, mediaType string) (*TMDBMetadata, error) {
	detailsURL := fmt.Sprintf("https://api.themoviedb.org/3/%s/%d?api_key=%s", mediaType, tmdbID, apiKey)
	var metadata TMDBMetadata

	if err := t.getAndDecodeJSON(detailsURL, &metadata); err != nil {
		return nil, fmt.Errorf("failed to get details: %w", err)
	}

	return &metadata, nil
}

func (t *titleTool) getAndDecodeJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}


type TitleInput struct {
	Name         string   `json:"name" jsonschema:"The name of the movie or TV show"`
	ReleaseYear  uint16   `json:"release_year,omitempty" jsonschema:"The year the title was released"`
	Genres       []string `json:"genres,omitempty" jsonschema:"List of genres"`
	ImdbRating   *float32 `json:"imdb_rating,omitempty" jsonschema:"The IMDb rating (e.g. 8.5)"`
	ImdbId       *string  `json:"imdb_id,omitempty" jsonschema:"The IMDb ID (e.g. tt1234567)"`
	TmdbId       *string  `json:"tmdb_id,omitempty" jsonschema:"The TMDB ID"`
	PosterPath   *string  `json:"poster_path,omitempty" jsonschema:"URL path to the poster image"`
	BackdropPath *string  `json:"backdrop_path,omitempty" jsonschema:"URL path to the backdrop image"`
	Overview     *string  `json:"overview,omitempty" jsonschema:"A brief plot summary"`
	Watched      bool     `json:"watched,omitempty" jsonschema:"Whether you have already watched this"`
	Wished       bool     `json:"wished,omitempty" jsonschema:"Whether you want to watch this"`
}

type AddTitlesInput struct {
	Titles []TitleInput `json:"titles"`
}

type AddTitlesOutput struct {
	Status      string                `json:"status" jsonschema:"status of operation"`
	TitlesAdded []models.TitleSummary `json:"titles_added" jsonschema:"list of summarised titles added"`
}

func (t *titleTool) AddTitles(ctx context.Context, req *mcp.CallToolRequest, input AddTitlesInput) (
	*mcp.CallToolResult,
	AddTitlesOutput,
	error,
) {
	titles := make([]models.Title, len(input.Titles))
	for i, title := range input.Titles {
		titles[i] = models.Title{
			Name:         title.Name,
			ReleaseYear:  title.ReleaseYear,
			Genres:       title.Genres,
			ImdbRating:   title.ImdbRating,
			ImdbId:       title.ImdbId,
			TmdbId:       title.TmdbId,
			PosterPath:   title.PosterPath,
			BackdropPath: title.BackdropPath,
			Overview:     title.Overview,
			Watched:      title.Watched,
			Wished:       title.Wished,
		}
	}

	titlesAdded, err := t.repository.AddTitles(ctx, titles)
	if err != nil {
		return nil, AddTitlesOutput{Status: "error"}, err
	}
	return nil, AddTitlesOutput{Status: "success", TitlesAdded: titlesAdded}, nil
}

type MarkTitleAsWatchedInput struct {
	TitleId uint `json:"title_id"`
}

type MarkTitleAsWatchedOutput struct {
	Status string `json:"status"`
}

func (t *titleTool) MarkTitleAsWatched(ctx context.Context, req *mcp.CallToolRequest, input MarkTitleAsWatchedInput) (
	*mcp.CallToolResult,
	MarkTitleAsWatchedOutput,
	error,
) {
	updates := map[string]interface{}{"watched": true, "wished": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return nil, MarkTitleAsWatchedOutput{Status: "error"}, err
	}
	return nil, MarkTitleAsWatchedOutput{Status: "success"}, nil
}

type GetAllTitlesInput struct {
}

type GetAllTitlesOutput struct {
	Status string                `json:"status"`
	Titles []models.TitleSummary `json:"titles"`
}

func (t *titleTool) ListAllTitles(ctx context.Context, req *mcp.CallToolRequest, input GetAllTitlesInput) (
	*mcp.CallToolResult,
	GetAllTitlesOutput,
	error,
) {
	titles, err := t.repository.GetAllTitles(ctx)
	if err != nil {
		return nil, GetAllTitlesOutput{Status: "error", Titles: titles}, err
	}
	return nil, GetAllTitlesOutput{Status: "success", Titles: titles}, nil
}

type GetTitlesByIdsInput struct {
	Ids []uint `json:"ids"`
}

type GetTitlesByIdsOutput struct {
	Status string         `json:"status"`
	Titles []models.Title `json:"titles"`
}

func (t *titleTool) GetTitlesByIds(ctx context.Context, req *mcp.CallToolRequest, input GetTitlesByIdsInput) (
	*mcp.CallToolResult,
	GetTitlesByIdsOutput,
	error,
) {
	titles, err := t.repository.GetTitlesByIds(ctx, input.Ids)
	if err != nil {
		return nil, GetTitlesByIdsOutput{Status: "error", Titles: titles}, err
	}
	return nil, GetTitlesByIdsOutput{Status: "success", Titles: titles}, nil
}

type MarkTitleAsWishedInput struct {
	TitleId uint `json:"title_id"`
}

type MarkTitleAsWishedOutput struct {
	Status string `json:"status"`
}

func (t *titleTool) MarkTitleAsWished(ctx context.Context, req *mcp.CallToolRequest, input MarkTitleAsWishedInput) (
	*mcp.CallToolResult,
	MarkTitleAsWishedOutput,
	error,
) {
	updates := map[string]interface{}{"wished": true, "watched": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return nil, MarkTitleAsWishedOutput{Status: "error"}, err
	}
	return nil, MarkTitleAsWishedOutput{Status: "success"}, nil
}

type ListWatchedTitlesInput struct {
}

type ListWatchedTitlesOutput struct {
	Status        string                `json:"status"`
	WatchedTitles []models.TitleSummary `json:"watched_titles"`
}

func (t *titleTool) ListWatchedTitles(ctx context.Context, req *mcp.CallToolRequest, input ListWatchedTitlesInput) (
	*mcp.CallToolResult,
	ListWatchedTitlesOutput,
	error,
) {
	watched_titles, err := t.repository.GetWatchedTitles(ctx)
	if err != nil {
		return nil, ListWatchedTitlesOutput{Status: "error", WatchedTitles: watched_titles}, err
	}
	return nil, ListWatchedTitlesOutput{Status: "success", WatchedTitles: watched_titles}, nil
}

type SearchTitlesInput struct {
	TitleNames       *[]string                  `json:"title_names,omitempty" jsonschema:"List of title names to filter by"`
	ReleaseYearRange *database.ReleaseYearRange `json:"release_year_range,omitempty" jsonschema:"Range of release years"`
	Watched          *bool                      `json:"watched,omitempty" jsonschema:"Filter by watched status"`
	Wished           *bool                      `json:"wished,omitempty" jsonschema:"Filter by wished status"`
}

type SearchTitlesOutput struct {
	Status string         `json:"status"`
	Titles []models.Title `json:"titles"`
}

func (t *titleTool) SearchTitles(ctx context.Context, req *mcp.CallToolRequest, input SearchTitlesInput) (
	*mcp.CallToolResult,
	SearchTitlesOutput,
	error,
) {
	summaries, err := t.repository.SearchTitles(ctx, database.SearchParams{
		TitleNames:       input.TitleNames,
		ReleaseYearRange: input.ReleaseYearRange,
		Watched:          input.Watched,
		Wished:           input.Wished,
	})
	if err != nil {
		return nil, SearchTitlesOutput{Status: "error"}, err
	}

	ids := make([]uint, len(summaries))
	for i, s := range summaries {
		ids[i] = s.ID
	}

	titles, err := t.repository.GetTitlesByIds(ctx, ids)
	if err != nil {
		return nil, SearchTitlesOutput{Status: "error"}, err
	}

	return nil, SearchTitlesOutput{Status: "success", Titles: titles}, nil
}

type RemoveTitleFromWatchedInput struct {
	TitleId uint `json:"title_id" jsonschema:"The ID of the title to remove from watched list"`
}

type RemoveTitleFromWatchedOutput struct {
	Status string `json:"status"`
}

func (t *titleTool) RemoveTitleFromWatched(ctx context.Context, req *mcp.CallToolRequest, input RemoveTitleFromWatchedInput) (
	*mcp.CallToolResult,
	RemoveTitleFromWatchedOutput,
	error,
) {
	updates := map[string]interface{}{"watched": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return nil, RemoveTitleFromWatchedOutput{Status: "error"}, err
	}
	return nil, RemoveTitleFromWatchedOutput{Status: "success"}, nil
}

type RemoveTitleFromWishedInput struct {
	TitleId uint `json:"title_id" jsonschema:"The ID of the title to remove from wished list"`
}

type RemoveTitleFromWishedOutput struct {
	Status string `json:"status"`
}

func (t *titleTool) RemoveTitleFromWished(ctx context.Context, req *mcp.CallToolRequest, input RemoveTitleFromWishedInput) (
	*mcp.CallToolResult,
	RemoveTitleFromWishedOutput,
	error,
) {
	updates := map[string]interface{}{"wished": false}
	if err := t.repository.UpdateTitle(ctx, input.TitleId, updates); err != nil {
		return nil, RemoveTitleFromWishedOutput{Status: "error"}, err
	}
	return nil, RemoveTitleFromWishedOutput{Status: "success"}, nil
}
