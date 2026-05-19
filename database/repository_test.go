package database

import (
	"context"
	"errors"
	"testing"
	"titles-mcp/database/models"
	"titles-mcp/database/sentinel"
	"titles-mcp/testdata"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	// 1. Remove ?cache=shared so each test gets a unique DB
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 2. Restrict the connection pool to 1 so GORM doesn't lose the in-memory state
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)

	// AutoMigrate the schema for testing
	err = db.AutoMigrate(&models.Title{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func Test_AddTitles(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewRepository(db)

	titles := []models.Title{testdata.TitleTestData}

	ctx := context.Background()
	summaries, err := repo.AddTitles(ctx, titles)
	if err != nil {
		t.Error("Add titles test failed", err)
	}
	if len(summaries) != 1 {
		t.Errorf("Expected 1 summary, got %d", len(summaries))
	}
}

func Test_UpdateTitle(t *testing.T) {
	t.Parallel()

	db := setupTestDB(t)
	repo := NewRepository(db)

	title := testdata.TitleTestData
	title.Watched = false

	ctx := context.Background()
	summaries, err := repo.AddTitles(ctx, []models.Title{title})
	if err != nil {
		t.Fatal("Add title operation failed", err)
	}
	id := summaries[0].ID

	updates := map[string]interface{}{"watched": true}
	if err := repo.UpdateTitle(ctx, id, updates); err != nil {
		t.Error("UpdateTitle test failed", err)
	}

	var updatedTitle models.Title
	if err := db.First(&updatedTitle, id).Error; err != nil {
		t.Error("Failed to fetch updated title", err)
	}

	if updatedTitle.Watched != true {
		t.Error("Expected title.watched to be true, found: ", updatedTitle.Watched)
	}
}

func Test_UpdateNonExistingTitle(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	updates := map[string]interface{}{"watched": true}
	if err := repo.UpdateTitle(ctx, 999, updates); !errors.Is(err, sentinel.ErrTitleNotFound) {
		t.Error("ErrTitleNotFound was expected. Got: ", err)
	}
}

func Test_GetWatchedTitles(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()
	title1 := testdata.TitleTestData
	title1.Watched = true
	title2 := testdata.TitleTestData2
	title2.Watched = true
	title3 := testdata.TitleTestData3
	title3.Watched = false
	_, _ = repo.AddTitles(ctx, []models.Title{title1, title2, title3})

	watched_titles, err := repo.GetWatchedTitles(ctx)
	if err != nil {
		t.Error(err)
	}
	if len(watched_titles) != 2 {
		t.Errorf("Expected len of watched titles to be 2, found %d", len(watched_titles))
	}

	for _, title := range watched_titles {
		if !title.Watched {
			t.Error("Expected watched as true, found false")
		}
	}
}

func Test_SearchTitles(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	titles := []models.Title{
		testdata.TitleTestData,
		testdata.TitleTestData2,
		testdata.TitleTestData3,
	}
	_, err := repo.AddTitles(ctx, titles)
	if err != nil {
		t.Fatalf("Failed to add titles: %v", err)
	}

	// Test case 1: Search by name
	summaries, err := repo.SearchTitles(ctx, SearchParams{
		TitleNames: &[]string{testdata.TitleTestData.Name},
	})
	if err != nil {
		t.Errorf("Search by name failed: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("Expected 1 result, got %d", len(summaries))
	}

	// Test case 2: Search by watched status
	watched := true
	summaries, err = repo.SearchTitles(ctx, SearchParams{
		Watched: &watched,
	})
	if err != nil {
		t.Errorf("Search by watched failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("Expected 2 results, got %d", len(summaries))
	}

	// Test case 3: Search by year range
	summaries, err = repo.SearchTitles(ctx, SearchParams{
		ReleaseYearRange: &ReleaseYearRange{From: 1970, To: 1980},
	})
	if err != nil {
		t.Errorf("Search by year range failed: %v", err)
	}
	if len(summaries) != 2 {
		t.Errorf("Expected 2 results, got %d", len(summaries))
	}
}

func Test_GetTitlesByIds(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	repo := NewRepository(db)
	ctx := context.Background()

	titles := []models.Title{
		testdata.TitleTestData,
		testdata.TitleTestData2,
	}
	addedSummaries, err := repo.AddTitles(ctx, titles)
	if err != nil {
		t.Fatalf("Failed to add titles: %v", err)
	}

	ids := []uint{addedSummaries[0].ID, addedSummaries[1].ID}
	fetchedTitles, err := repo.GetTitlesByIds(ctx, ids)
	if err != nil {
		t.Errorf("GetTitlesByIds failed: %v", err)
	}

	if len(fetchedTitles) != 2 {
		t.Errorf("Expected 2 titles, got %d", len(fetchedTitles))
	}
}
