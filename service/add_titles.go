package service

import (
	"context"
	"titles-mcp/database/models"
)

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

func (t *titleService) AddTitles(ctx context.Context, input AddTitlesInput) (AddTitlesOutput, error) {
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
		return AddTitlesOutput{Status: "error"}, err
	}
	return AddTitlesOutput{Status: "success", TitlesAdded: titlesAdded}, nil
}
