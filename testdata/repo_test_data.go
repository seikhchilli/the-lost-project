package testdata

import (
	"titles-mcp/database/models"
)

var (
	imdbRating = float32(9.1)
	imdbID     = "tt2860845951"
	tmdbID     = "10646613"

	TitleTestData = models.Title{
		Name:        "title name",
		ReleaseYear: 2001,
		Genres:      []string{"romance", "comedy"},
		ImdbRating:  &imdbRating,
		ImdbId:      &imdbID,
		TmdbId:      &tmdbID,
		Watched:     true,
	}

	imdbRating2 = float32(5.1)
	imdbID2     = "tt2860845950"
	tmdbID2     = "10646610"

	TitleTestData2 = models.Title{
		Name:        "title name 2",
		ReleaseYear: 1971,
		Genres:      []string{"action", "thriller"},
		ImdbRating:  &imdbRating2,
		ImdbId:      &imdbID2,
		TmdbId:      &tmdbID2,
		Watched:     true,
	}

	imdbRating3 = float32(5.1)
	imdbID3     = "tt2870845950"
	tmdbID3     = "10645610"

	TitleTestData3 = models.Title{
		Name:        "title name 3",
		ReleaseYear: 1971,
		Genres:      []string{"action", "thriller"},
		ImdbRating:  &imdbRating3,
		ImdbId:      &imdbID3,
		TmdbId:      &tmdbID3,
		Watched:     false,
	}
)
