package internal

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
}

type CacheConfig struct {
	Addr     string
	Password string
	DB       int
}

func NewCacheService(redisURL string) *CacheService {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		opt = &redis.Options{
			Addr:     "localhost:6379",
			Password: "",
			DB:       0,
		}
	}

	rdb := redis.NewClient(opt)
	return &CacheService{client: rdb}
}

func NewCacheServiceWithConfig(config CacheConfig) *CacheService {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})
	return &CacheService{client: rdb}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, jsonData, expiration).Err()
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	result, err := c.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(result), dest)
}

func (c *CacheService) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *CacheService) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *CacheService) Close() error {
	return c.client.Close()
}

// GetClient returns the underlying Redis client for advanced operations
func (c *CacheService) GetClient() *redis.Client {
	return c.client
}
