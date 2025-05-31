package cacheutil

import (
	"context"
	"errors"
	"fmt"
	redis2 "github.com/seventeenthearth/sudal/internal/infrastructure/database/redis"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_cache_util.go -package=mocks -mock_names=CacheUtil=MockCacheUtil github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil CacheUtil

// CacheUtil defines the interface for cache operations
// This interface abstracts cache operations for better testability
type CacheUtil interface {
	// Set stores a key-value pair with an optional TTL
	Set(key string, value string, ttl time.Duration) error

	// Get retrieves the value for a given key
	Get(key string) (string, error)

	// Delete removes a key-value pair from the cache
	Delete(key string) error

	// DeleteByPattern deletes all keys matching a pattern
	DeleteByPattern(pattern string) error
}

// CacheUtilImpl provides simple key-value caching operations using Redis
type CacheUtilImpl struct {
	redisManager redis2.RedisManager
	logger       *zap.Logger
}

// NewCacheUtil creates a new cache utility instance
func NewCacheUtil(redisManager redis2.RedisManager) CacheUtil {
	return &CacheUtilImpl{
		redisManager: redisManager,
		logger:       log.GetLogger().With(zap.String("component", "cache_util")),
	}
}

// Set stores a key-value pair with an optional TTL
// If ttl is zero or negative, the key will persist indefinitely
func (c *CacheUtilImpl) Set(key string, value string, ttl time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()
	if c.redisManager == nil {
		return fmt.Errorf("redis client is not available")
	}
	client := c.redisManager.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Setting cache key",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)

	var err error
	if ttl > 0 {
		err = client.Set(ctx, key, value, ttl).Err()
	} else {
		err = client.Set(ctx, key, value, 0).Err()
	}

	if err != nil {
		c.logger.Error("Failed to set cache key",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set cache key '%s': %w", key, err)
	}

	c.logger.Debug("Successfully set cache key",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)

	return nil
}

// Get retrieves the value for a given key
// Returns ErrCacheMiss if the key is not found or has expired
func (c *CacheUtilImpl) Get(key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()
	if c.redisManager == nil {
		return "", fmt.Errorf("redis client is not available")
	}
	client := c.redisManager.GetClient()
	if client == nil {
		return "", fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Getting cache key",
		zap.String("key", key),
	)

	value, err := client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Debug("Cache miss for key",
				zap.String("key", key),
			)
			return "", ErrCacheMiss
		}

		c.logger.Error("Failed to get cache key",
			zap.String("key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get cache key '%s': %w", key, err)
	}

	c.logger.Debug("Successfully retrieved cache key",
		zap.String("key", key),
		zap.Int("value_length", len(value)),
	)

	return value, nil
}

// Delete removes a key-value pair from the cache
func (c *CacheUtilImpl) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()
	if c.redisManager == nil {
		return fmt.Errorf("redis client is not available")
	}
	client := c.redisManager.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Deleting cache key",
		zap.String("key", key),
	)

	deletedCount, err := client.Del(ctx, key).Result()
	if err != nil {
		c.logger.Error("Failed to delete cache key",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete cache key '%s': %w", key, err)
	}

	c.logger.Debug("Successfully deleted cache key",
		zap.String("key", key),
		zap.Int64("deleted_count", deletedCount),
	)

	return nil
}

// DeleteByPattern deletes all keys matching a pattern
// This is useful for test cleanup
func (c *CacheUtilImpl) DeleteByPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	ctx := context.Background()
	if c.redisManager == nil {
		return fmt.Errorf("redis client is not available")
	}
	client := c.redisManager.GetClient()
	if client == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Deleting cache keys by pattern",
		zap.String("pattern", pattern),
	)

	// Get all keys matching the pattern
	keys, err := client.Keys(ctx, pattern).Result()
	if err != nil {
		c.logger.Error("Failed to get keys by pattern",
			zap.String("pattern", pattern),
			zap.Error(err),
		)
		return fmt.Errorf("failed to get keys by pattern '%s': %w", pattern, err)
	}

	if len(keys) == 0 {
		c.logger.Debug("No keys found matching pattern",
			zap.String("pattern", pattern),
		)
		return nil
	}

	// Delete all matching keys
	deletedCount, err := client.Del(ctx, keys...).Result()
	if err != nil {
		c.logger.Error("Failed to delete keys by pattern",
			zap.String("pattern", pattern),
			zap.Strings("keys", keys),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete keys by pattern '%s': %w", pattern, err)
	}

	c.logger.Debug("Successfully deleted cache keys by pattern",
		zap.String("pattern", pattern),
		zap.Int("keys_found", len(keys)),
		zap.Int64("deleted_count", deletedCount),
	)

	return nil
}
