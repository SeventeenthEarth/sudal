package database

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_redis_client.go -package=mocks github.com/seventeenthearth/sudal/internal/infrastructure/database RedisClient

// RedisClient defines the interface for Redis client operations
// This interface abstracts Redis operations for better testability
type RedisClient interface {
	// Ping checks the Redis connection
	Ping(ctx context.Context) *redis.StatusCmd

	// Close closes the Redis connection
	Close() error

	// PoolStats returns connection pool statistics
	PoolStats() *redis.PoolStats
}

// RedisManager manages Redis client connections
type RedisManager struct {
	client RedisClient
	config *config.Config
	logger *zap.Logger
}

// NewRedisManager creates a new Redis connection manager with comprehensive configuration
func NewRedisManager(cfg *config.Config) (*RedisManager, error) {
	logger := log.GetLogger().With(zap.String("component", "redis_manager"))

	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	logger.Info("Initializing Redis client",
		zap.String("addr", cfg.Redis.Addr),
		zap.Bool("password_set", cfg.Redis.Password != ""),
		zap.Int("db", cfg.Redis.DB),
		zap.Int("pool_size", cfg.Redis.PoolSize),
		zap.Int("min_idle_conns", cfg.Redis.MinIdleConns),
		zap.Int("pool_timeout_seconds", cfg.Redis.PoolTimeout),
		zap.Int("idle_timeout_seconds", cfg.Redis.IdleTimeout),
		zap.Int("dial_timeout_seconds", cfg.Redis.DialTimeout),
		zap.Int("read_timeout_seconds", cfg.Redis.ReadTimeout),
		zap.Int("write_timeout_seconds", cfg.Redis.WriteTimeout),
		zap.Int("max_retries", cfg.Redis.MaxRetries),
		zap.Int("min_retry_backoff_ms", cfg.Redis.MinRetryBackoff),
		zap.Int("max_retry_backoff_ms", cfg.Redis.MaxRetryBackoff),
	)

	// Create Redis client options with comprehensive configuration
	options := &redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,

		// Connection Pool Configuration
		PoolSize:        cfg.Redis.PoolSize,
		MinIdleConns:    cfg.Redis.MinIdleConns,
		PoolTimeout:     time.Duration(cfg.Redis.PoolTimeout) * time.Second,
		ConnMaxIdleTime: time.Duration(cfg.Redis.IdleTimeout) * time.Second,

		// Timeout Configuration
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,

		// Retry Configuration
		MaxRetries:      cfg.Redis.MaxRetries,
		MinRetryBackoff: time.Duration(cfg.Redis.MinRetryBackoff) * time.Millisecond,
		MaxRetryBackoff: time.Duration(cfg.Redis.MaxRetryBackoff) * time.Millisecond,
	}

	// Create Redis client
	client := redis.NewClient(options)

	// Test the connection with retry logic
	manager := &RedisManager{
		client: client,
		config: cfg,
		logger: logger,
	}

	// Test initial connection with retry
	if err := manager.testConnectionWithRetry(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to establish Redis connection: %w", err)
	}

	logger.Info("Redis client initialized successfully")

	return manager, nil
}

// GetClient returns the underlying Redis client
// This should be used sparingly and only when direct access to *redis.Client is needed
func (rm *RedisManager) GetClient() *redis.Client {
	if client, ok := rm.client.(*redis.Client); ok {
		return client
	}
	return nil
}

// Ping performs a health check on the Redis connection with retry logic
func (rm *RedisManager) Ping(ctx context.Context) error {
	rm.logger.Debug("Performing Redis health check")

	err := rm.executeWithRetry(ctx, "health_check", func() error {
		return rm.client.Ping(ctx).Err()
	})

	if err != nil {
		rm.logger.Error("Redis health check failed after retries",
			log.FormatError(err),
		)
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	rm.logger.Debug("Redis health check successful")
	return nil
}

// Close closes the Redis connection
func (rm *RedisManager) Close() error {
	rm.logger.Info("Closing Redis connection")
	return rm.client.Close()
}

// testConnectionWithRetry tests the Redis connection with retry logic
func (rm *RedisManager) testConnectionWithRetry() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return rm.executeWithRetry(ctx, "connection_test", func() error {
		return rm.client.Ping(ctx).Err()
	})
}

// executeWithRetry executes a Redis operation with retry logic
func (rm *RedisManager) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error
	maxRetries := rm.config.Redis.MaxRetries
	minBackoff := time.Duration(rm.config.Redis.MinRetryBackoff) * time.Millisecond
	maxBackoff := time.Duration(rm.config.Redis.MaxRetryBackoff) * time.Millisecond

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := fn()
		if err == nil {
			if attempt > 0 {
				rm.logger.Info("Redis operation succeeded after retry",
					zap.String("operation", operation),
					zap.Int("attempt", attempt+1),
					zap.Int("max_retries", maxRetries+1),
				)
			}
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !rm.isRetryableError(err) {
			rm.logger.Error("Redis operation failed with non-retryable error",
				zap.String("operation", operation),
				zap.Int("attempt", attempt+1),
				log.FormatError(err),
			)
			return fmt.Errorf("redis operation '%s' failed: %w", operation, err)
		}

		// Don't retry if we've reached max attempts
		if attempt >= maxRetries {
			rm.logger.Error("Redis operation failed after all retries",
				zap.String("operation", operation),
				zap.Int("total_attempts", attempt+1),
				zap.Int("max_retries", maxRetries+1),
				log.FormatError(lastErr),
			)
			break
		}

		// Calculate backoff duration with exponential backoff
		backoff := minBackoff
		if attempt > 0 {
			multiplier := 1 << uint(attempt-1)
			backoff = time.Duration(int64(minBackoff) * int64(multiplier))
			backoff = min(backoff, maxBackoff)
		}

		rm.logger.Warn("Redis operation failed, retrying",
			zap.String("operation", operation),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", maxRetries+1),
			zap.Duration("backoff", backoff),
			log.FormatError(err),
		)

		// Wait before retry, but respect context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("redis operation '%s' cancelled: %w", operation, ctx.Err())
		case <-time.After(backoff):
			// Continue to next retry
		}
	}

	return fmt.Errorf("redis operation '%s' failed after %d attempts: %w", operation, maxRetries+1, lastErr)
}

// isRetryableError determines if an error is retryable
func (rm *RedisManager) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Import the missing packages at the top if needed
	errorStr := err.Error()
	errorStrLower := strings.ToLower(errorStr)

	// Network-related errors that are typically retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Network timeout errors are retryable
		if netErr.Timeout() {
			rm.logger.Debug("Detected network timeout error", log.FormatError(err))
			return true
		}
	}

	// Redis-specific retryable errors
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"broken pipe",
		"no route to host",
		"network is unreachable",
		"i/o timeout",
		"connection timed out",
		"dial tcp",
		"eof",
		"connection lost",
		"server closed the connection",
		"pool timeout",
		"redis: connection pool exhausted",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errorStrLower, pattern) {
			rm.logger.Debug("Detected retryable Redis error pattern",
				zap.String("pattern", pattern),
				log.FormatError(err),
			)
			return true
		}
	}

	// Redis server errors that are NOT retryable (authentication, syntax, etc.)
	nonRetryablePatterns := []string{
		"auth",
		"noauth",
		"wrongpass",
		"syntax error",
		"unknown command",
		"wrong number of arguments",
		"operation not permitted",
	}

	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errorStrLower, pattern) {
			rm.logger.Debug("Detected non-retryable Redis error pattern",
				zap.String("pattern", pattern),
				log.FormatError(err),
			)
			return false
		}
	}

	// Default to retryable for unknown errors (conservative approach)
	rm.logger.Debug("Unknown error type, treating as retryable", log.FormatError(err))
	return true
}

// GetConnectionPoolStats returns Redis connection pool statistics
func (rm *RedisManager) GetConnectionPoolStats() *redis.PoolStats {
	if rm.client == nil {
		return nil
	}
	stats := rm.client.PoolStats()
	return stats
}

// LogConnectionPoolStats logs the current Redis connection pool statistics
func (rm *RedisManager) LogConnectionPoolStats() {
	stats := rm.GetConnectionPoolStats()
	if stats == nil {
		rm.logger.Warn("Unable to retrieve Redis connection pool statistics")
		return
	}

	rm.logger.Info("Redis connection pool statistics",
		zap.Uint32("hits", stats.Hits),
		zap.Uint32("misses", stats.Misses),
		zap.Uint32("timeouts", stats.Timeouts),
		zap.Uint32("total_conns", stats.TotalConns),
		zap.Uint32("idle_conns", stats.IdleConns),
		zap.Uint32("stale_conns", stats.StaleConns),
	)
}

// ExecuteWithRetry provides a public interface for executing Redis operations with retry logic
// This can be used by other parts of the application that need retry functionality
func (rm *RedisManager) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	return rm.executeWithRetry(ctx, operation, fn)
}
