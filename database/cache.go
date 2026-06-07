package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	IsKeyInCache(ctx context.Context, key string) bool
}

type redisCache struct {
	client *redis.Client
}

func NewCache() Cache {
	redisOptions := &redis.Options{
		Addr:     "192.168.1.50:6379",
		Password: "",
		DB:       0,
	}
	return &redisCache{
		client: redis.NewClient(redisOptions),
	}
}

func (r *redisCache) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		log.Printf("Redis client threw error for GET key: %s, error: %v", key, err)
		return "", fmt.Errorf("Failed to retrive key value from cache.")
	}
	return result, nil
}

func (r *redisCache) IsKeyInCache(ctx context.Context, key string) bool {
	_, err := r.client.Get(ctx, key).Result()
	return err != redis.Nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		log.Printf("Redis client threw error for SET key: %s, value: %v error: %v", key, value, err)
		return fmt.Errorf("Failed to retrive key value from cache.")
	}
	return nil
}
