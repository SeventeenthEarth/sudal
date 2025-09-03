package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	sredis "github.com/seventeenthearth/sudal/internal/service/redis"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_infra_cache.go -package=mocks -mock_names=CacheUtil=MockCacheUtil github.com/seventeenthearth/sudal/internal/infrastructure/cache CacheUtil

// CacheUtil defines the protocol for cache operations
// This protocol abstracts cache operations for better testability
type CacheUtil interface {
	// Set stores a key-value pair with an optional TTL
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Get retrieves the value for a given key
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a key-value pair from the cache
	Delete(ctx context.Context, key string) error

	// DeleteByPattern deletes all keys matching a pattern
	DeleteByPattern(ctx context.Context, pattern string) error
}

// CacheUtilImpl provides simple key-value caching operations using Redis
type CacheUtilImpl struct {
	kv     sredis.KV
	logger *zap.Logger
}

// NewCacheUtil creates a new cache utility instance
func NewCacheUtil(kv sredis.KV) CacheUtil {
	return &CacheUtilImpl{
		kv:     kv,
		logger: log.GetLogger().With(zap.String("component", "cache")),
	}
}

// Set stores a key-value pair with an optional TTL
// If ttl is zero or negative, the key will persist indefinitely
func (c *CacheUtilImpl) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if c.kv == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Setting cache key",
		zap.String("key", key),
		zap.Duration("ttl", ttl),
	)

	var err error
	if ttl > 0 {
		err = c.kv.Set(ctx, key, value, ttl)
	} else {
		err = c.kv.Set(ctx, key, value, 0)
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
func (c *CacheUtilImpl) Get(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("key cannot be empty")
	}

	if c.kv == nil {
		return "", fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Getting cache key",
		zap.String("key", key),
	)

	value, err := c.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, sredis.ErrNotFound) {
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
func (c *CacheUtilImpl) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if c.kv == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Deleting cache key",
		zap.String("key", key),
	)

	deletedCount, err := c.kv.Del(ctx, key)
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
func (c *CacheUtilImpl) DeleteByPattern(ctx context.Context, pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern cannot be empty")
	}

	if c.kv == nil {
		return fmt.Errorf("redis client is not available")
	}

	c.logger.Debug("Deleting cache keys by pattern",
		zap.String("pattern", pattern),
	)

	// Get all keys matching the pattern
	keys, err := c.kv.Keys(ctx, pattern)
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

	// Delete all matching keys in batches to avoid sending overly large commands
	const batchSize = 1000
	var totalDeleted int64
	for i := 0; i < len(keys); i += batchSize {
		end := i + batchSize
		if end > len(keys) {
			end = len(keys)
		}
		deletedCount, err := c.kv.Del(ctx, keys[i:end]...)
		if err != nil {
			c.logger.Error("Failed to delete keys by pattern (batch)",
				zap.String("pattern", pattern),
				zap.Int("batch_start", i),
				zap.Int("batch_end", end),
				zap.Error(err),
			)
			return fmt.Errorf("failed to delete keys by pattern '%s': %w", pattern, err)
		}
		totalDeleted += deletedCount
		if err := ctx.Err(); err != nil {
			return err
		}
	}

	c.logger.Debug("Successfully deleted cache keys by pattern",
		zap.String("pattern", pattern),
		zap.Int("keys_found", len(keys)),
		zap.Int64("deleted_count", totalDeleted),
	)

	return nil
}
