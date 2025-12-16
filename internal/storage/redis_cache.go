package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache implements caching using Redis
type RedisCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(addr, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
		ttl:    ttl,
	}, nil
}

// Get retrieves a value from cache
func (c *RedisCache) Get(key string) ([]byte, error) {
	val, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}
	return val, nil
}

// Set stores a value in cache
func (c *RedisCache) Set(key string, value []byte) error {
	return c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL stores a value in cache with custom TTL
func (c *RedisCache) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	if err := c.client.Set(c.ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}
	return nil
}

// Delete removes a value from cache
func (c *RedisCache) Delete(key string) error {
	if err := c.client.Del(c.ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern
func (c *RedisCache) DeletePattern(pattern string) error {
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	for iter.Next(c.ctx) {
		if err := c.client.Del(c.ctx, iter.Val()).Err(); err != nil {
			return fmt.Errorf("failed to delete pattern from cache: %w", err)
		}
	}
	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (c *RedisCache) Close() error {
	return c.client.Close()
}

// Cache key generators
func cacheKeyUserFavorites(userID, listReference string) string {
	if listReference == "" {
		listReference = "default"
	}
	return fmt.Sprintf("favorites:user:%s:list:%s", userID, listReference)
}

func cacheKeyAsset(assetID string) string {
	return fmt.Sprintf("asset:%s", assetID)
}

// ErrCacheMiss indicates a cache miss
var ErrCacheMiss = fmt.Errorf("cache miss")

