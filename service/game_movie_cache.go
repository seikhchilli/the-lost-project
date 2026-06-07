package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"titles-mcp/database"
)

const cacheKey = "NEXT_TITLES_POOL"

type GameNextTitlesPool interface {
	AddTitles(ctx context.Context, titles []TitleNameAndReleaseYear) error
	GetNextTitle(ctx context.Context) (TitleNameAndReleaseYear, error)
	GetPoolSize(ctx context.Context) (int, error)
}

type gameNextTitlesPool struct {
	cache    database.Cache
	cacheKey string
	cacheTtl time.Duration
}

func NewGameNextTitlesPool() GameNextTitlesPool {
	return &gameNextTitlesPool{
		cache:    database.NewCache(),
		cacheKey: cacheKey,
		cacheTtl: 5 * time.Hour,
	}
}

// TODO: use set for next titles
func (g *gameNextTitlesPool) AddTitles(ctx context.Context, titles []TitleNameAndReleaseYear) error {
	var nextTitles []TitleNameAndReleaseYear
	if g.cache.IsKeyInCache(ctx, g.cacheKey) {
		cacheValue, err := g.cache.Get(ctx, cacheKey)
		if err != nil {
			log.Print(err)
			return fmt.Errorf("Failed to add titles to NextTitlesPool")
		}

		err = json.Unmarshal([]byte(cacheValue), &nextTitles)
		if err != nil {
			log.Print(err, cacheValue)
			return fmt.Errorf("Failed to add titles to NextTitlesPool")
		}
	}

	nextTitles = append(nextTitles, titles...)
	bytes, marshalErr := json.Marshal(nextTitles)
	if marshalErr != nil {
		log.Print(marshalErr)
		return fmt.Errorf("Failed to marshal titles to JSON")
	}
	err := g.cache.Set(ctx, g.cacheKey, string(bytes), g.cacheTtl)
	if err != nil {
		log.Print(err, nextTitles)
		return fmt.Errorf("Failed to add titles to NextTitlesPool")
	}
	return nil
}

func (g *gameNextTitlesPool) GetNextTitle(ctx context.Context) (TitleNameAndReleaseYear, error) {
	var nextTitles []TitleNameAndReleaseYear
	var nextTitle TitleNameAndReleaseYear
	cacheValue, err := g.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Print(err)
		return TitleNameAndReleaseYear{}, fmt.Errorf("Failed to get titles from NextTitlesPool")
	}

	err = json.Unmarshal([]byte(cacheValue), &nextTitles)
	if err != nil {
		log.Print(err, cacheValue)
		return TitleNameAndReleaseYear{}, fmt.Errorf("Failed to get titles from NextTitlesPool")
	}
	if len(nextTitles) == 0 {
		return TitleNameAndReleaseYear{}, fmt.Errorf("NextTitlesPool is empty")
	}
	nextTitle = nextTitles[0]
	nextTitles = nextTitles[1:]
	bytes, marshalErr := json.Marshal(nextTitles)
	if marshalErr != nil {
		log.Print(marshalErr)
		return TitleNameAndReleaseYear{}, fmt.Errorf("Failed to marshal titles to JSON")
	}
	err = g.cache.Set(ctx, g.cacheKey, string(bytes), g.cacheTtl)
	if err != nil {
		log.Print(err, cacheValue)
		return TitleNameAndReleaseYear{}, fmt.Errorf("Failed to get titles from NextTitlesPool")
	}
	return nextTitle, nil
}

func (g *gameNextTitlesPool) GetPoolSize(ctx context.Context) (int, error) {
	if !g.cache.IsKeyInCache(ctx, g.cacheKey) {
		return 0, nil
	}
	var nextTitles []TitleNameAndReleaseYear
	cacheValue, err := g.cache.Get(ctx, cacheKey)
	if err != nil {
		log.Print(err)
		return -1, fmt.Errorf("Failed to get titles from NextTitlesPool")
	}

	err = json.Unmarshal([]byte(cacheValue), &nextTitles)
	if err != nil {
		log.Print(err, cacheValue)
		return -1, fmt.Errorf("Failed to get titles from NextTitlesPool")
	}
	return len(nextTitles), nil
}
