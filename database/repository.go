package database

import (
	"context"
	"strings"
	"errors"
	"titles-mcp/database/models"
	"titles-mcp/database/sentinel"

	"gorm.io/gorm"
)

type Repository interface {
	AddTitles(ctx context.Context, titles []models.Title) ([]models.TitleSummary, error)
	GetAllTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error)
	GetWatchedTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error)
	GetWishedTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error)
	UpdateTitle(ctx context.Context, id uint, updates map[string]interface{}) error
	SearchTitles(ctx context.Context, searchParams SearchParams, page, pageSize int) ([]models.TitleSummary, int64, error)
	GetTitlesByIds(ctx context.Context, IDs []uint) ([]models.Title, error)
	DeleteTitle(ctx context.Context, id uint) error
	GetExistingTmdbIds(ctx context.Context, tmdbIds []string) ([]string, error)
}

type repository struct {
	db *gorm.DB
}

type SearchParams struct {
	TitleNames       *[]string         `json:"title_names"`
	ReleaseYearRange *ReleaseYearRange `json:"release_year_range"`
	Watched          *bool             `json:"watched"`
	Wished           *bool             `json:"wished"`
}

type ReleaseYearRange struct {
	From uint16 `json:"from"`
	To   uint16 `json:"to"`
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func paginate(query *gorm.DB, page, pageSize int, total *int64) *gorm.DB {
	query.Count(total)
	if page > 0 && pageSize > 0 {
		offset := (page - 1) * pageSize
		return query.Offset(offset).Limit(pageSize)
	}
	return query
}

func (r *repository) GetExistingTmdbIds(ctx context.Context, tmdbIds []string) ([]string, error) {
	var existingIds []string
	err := r.db.WithContext(ctx).
		Model(&models.Title{}).
		Where("tmdb_id IN ?", tmdbIds).
		Pluck("tmdb_id", &existingIds).Error
	return existingIds, err
}

func (r *repository) AddTitles(ctx context.Context, titles []models.Title) ([]models.TitleSummary, error) {
	if err := r.db.WithContext(ctx).Create(&titles).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || 
		   strings.Contains(strings.ToLower(err.Error()), "unique constraint") || 
		   strings.Contains(strings.ToLower(err.Error()), "duplicate key") {
			return nil, sentinel.ErrTitleAlreadyExists
		}
		return nil, err
	}
	titlesAdded := make([]models.TitleSummary, len(titles))
	for i, title := range titles {
		titlesAdded[i] = models.TitleSummary{
			ID:          title.ID,
			Name:        title.Name,
			ReleaseYear: title.ReleaseYear,
			Watched:     title.Watched,
			Wished:      title.Wished,
			PosterPath:  title.PosterPath,
		}
	}
	return titlesAdded, nil
}

func (r *repository) SearchTitles(ctx context.Context, searchParams SearchParams, page, pageSize int) ([]models.TitleSummary, int64, error) {
	var summaries []models.TitleSummary
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Title{})

	if searchParams.TitleNames != nil && len(*searchParams.TitleNames) > 0 {
		likeQuery := r.db
		for i, name := range *searchParams.TitleNames {
			if i == 0 {
				likeQuery = likeQuery.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
			} else {
				likeQuery = likeQuery.Or("LOWER(name) LIKE LOWER(?)", "%"+name+"%")
			}
		}
		query = query.Where(likeQuery)
	}

	if searchParams.ReleaseYearRange != nil {
		if searchParams.ReleaseYearRange.From > 0 {
			query = query.Where("release_year >= ?", searchParams.ReleaseYearRange.From)
		}
		if searchParams.ReleaseYearRange.To > 0 {
			query = query.Where("release_year <= ?", searchParams.ReleaseYearRange.To)
		}
	}

	if searchParams.Watched != nil {
		query = query.Where("watched = ?", *searchParams.Watched)
	}

	if searchParams.Wished != nil {
		query = query.Where("wished = ?", *searchParams.Wished)
	}

	query = paginate(query, page, pageSize, &total)

	err := query.Select("id", "name", "release_year", "watched", "wished", "poster_path").Find(&summaries).Error
	if err != nil {
		return nil, 0, err
	}

	return summaries, total, nil
}

func (r *repository) GetTitlesByIds(ctx context.Context, IDs []uint) ([]models.Title, error) {
	var titles []models.Title
	if err := r.db.WithContext(ctx).Where("id IN ?", IDs).Find(&titles).Error; err != nil {
		return nil, err
	}
	return titles, nil
}

func (r *repository) GetAllTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error) {
	var titles []models.TitleSummary
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Title{})
	query = paginate(query, page, pageSize, &total)
	err := query.Select("id", "name", "release_year", "watched", "wished", "poster_path").Find(&titles).Error
	return titles, total, err
}

func (r *repository) GetWatchedTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error) {
	var titles []models.TitleSummary
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Title{}).Where("watched = ?", true)
	query = paginate(query, page, pageSize, &total)
	err := query.Find(&titles).Error
	return titles, total, err
}

func (r *repository) GetWishedTitles(ctx context.Context, page, pageSize int) ([]models.TitleSummary, int64, error) {
	var titles []models.TitleSummary
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Title{}).Where("wished = ?", true)
	query = paginate(query, page, pageSize, &total)
	err := query.Find(&titles).Error
	return titles, total, err
}

func (r *repository) UpdateTitle(ctx context.Context, id uint, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&models.Title{}).Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return sentinel.ErrTitleNotFound
	}
	return nil
}

func (r *repository) DeleteTitle(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Title{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return sentinel.ErrTitleNotFound
	}
	return nil
}
