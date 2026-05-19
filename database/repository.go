package database

import (
	"context"
	"titles-mcp/database/models"
	"titles-mcp/database/sentinel"

	"gorm.io/gorm"
)

type Repository interface {
	AddTitles(ctx context.Context, titles []models.Title) ([]models.TitleSummary, error)
	GetAllTitles(ctx context.Context) ([]models.TitleSummary, error)
	GetWatchedTitles(ctx context.Context) ([]models.TitleSummary, error)
	GetWishedTitles(ctx context.Context) ([]models.TitleSummary, error)
	UpdateTitle(ctx context.Context, id uint, updates map[string]interface{}) error
	SearchTitles(ctx context.Context, searchParams SearchParams) ([]models.TitleSummary, error)
	GetTitlesByIds(ctx context.Context, IDs []uint) ([]models.Title, error)
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

func (r *repository) AddTitles(ctx context.Context, titles []models.Title) ([]models.TitleSummary, error) {
	if err := r.db.WithContext(ctx).Create(&titles).Error; err != nil {
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
		}
	}
	return titlesAdded, nil
}

func (r *repository) SearchTitles(ctx context.Context, searchParams SearchParams) ([]models.TitleSummary, error) {
	var summaries []models.TitleSummary
	query := r.db.WithContext(ctx).Model(&models.Title{})

	if searchParams.TitleNames != nil && len(*searchParams.TitleNames) > 0 {
		query = query.Where("name IN ?", *searchParams.TitleNames)
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

	err := query.Select("id", "name", "release_year", "watched", "wished").Find(&summaries).Error
	if err != nil {
		return nil, err
	}

	return summaries, nil
}

func (r *repository) GetTitlesByIds(ctx context.Context, IDs []uint) ([]models.Title, error) {
	var titles []models.Title
	if err := r.db.WithContext(ctx).Where("id IN ?", IDs).Find(&titles).Error; err != nil {
		return nil, err
	}
	return titles, nil
}

func (r *repository) GetAllTitles(ctx context.Context) ([]models.TitleSummary, error) {
	var titles []models.TitleSummary
	err := r.db.WithContext(ctx).Model(&models.Title{}).Select("id", "name", "release_year", "watched", "wished").Find(&titles).Error
	return titles, err
}

func (r *repository) GetWatchedTitles(ctx context.Context) ([]models.TitleSummary, error) {
	var titles []models.TitleSummary
	err := r.db.WithContext(ctx).Model(&models.Title{}).Where("watched = ?", true).Find(&titles).Error
	return titles, err
}

func (r *repository) GetWishedTitles(ctx context.Context) ([]models.TitleSummary, error) {
	var titles []models.TitleSummary
	err := r.db.WithContext(ctx).Model(&models.Title{}).Where("wished = ?", true).Find(&titles).Error
	return titles, err
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
