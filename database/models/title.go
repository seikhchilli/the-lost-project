package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Title struct {
	ID           uint           `gorm:"primarykey" json:"id,omitempty"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	Name         string         `json:"name" gorm:"index;not null"`
	ReleaseYear  uint16         `json:"release_year"`
	Genres       pq.StringArray `json:"genres" gorm:"type:text[]"`
	ImdbRating   *float32       `json:"imdb_rating"`
	ImdbId       *string        `json:"imdb_id" gorm:"uniqueIndex"`
	TmdbId       *string        `json:"tmdb_id" gorm:"uniqueIndex"`
	PosterPath   *string        `json:"poster_path"`
	BackdropPath *string        `json:"backdrop_path"`
	Overview     *string        `json:"overview"`
	Watched      bool           `json:"watched" gorm:"default:false"`
	Wished       bool           `json:"wished" gorm:"default:false"`
}

type TitleSummary struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	ReleaseYear uint16 `json:"release_year"`
	Watched     bool   `json:"watched"`
	Wished      bool   `json:"wished"`
}
