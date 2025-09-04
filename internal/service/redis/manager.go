package redis

//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_redis_manager.go -package=mocks -mock_names=RedisManager=MockRedisManager github.com/seventeenthearth/sudal/internal/service/redis RedisManager
//go:generate go run go.uber.org/mock/mockgen -destination=../../mocks/mock_redis_client.go -package=mocks github.com/seventeenthearth/sudal/internal/service/redis RedisClient

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	sconfig "github.com/seventeenthearth/sudal/internal/service/config"
	slogger "github.com/seventeenthearth/sudal/internal/service/logger"
)

// RedisClient defines the protocol for Redis client operations
// This protocol abstracts Redis operations for better testability
type RedisClient interface {
	// Ping checks the Redis connection
	Ping(ctx context.Context) *goredis.StatusCmd

	// Close closes the Redis connection
	Close() error

	// PoolStats returns connection pool statistics
	PoolStats() *goredis.PoolStats
}

// RedisManager defines the protocol for Redis connection management
// This protocol abstracts Redis manager operations for better testability
type RedisManager interface {
	// GetClient returns the underlying Redis client
	GetClient() *goredis.Client

	// Ping performs a health check on the Redis connection
	Ping(ctx context.Context) error

	// Close closes the Redis connection
	Close() error

	// GetConnectionPoolStats returns Redis connection pool statistics
	GetConnectionPoolStats() *goredis.PoolStats

	// LogConnectionPoolStats logs the current Redis connection pool statistics
	LogConnectionPoolStats()

	// ExecuteWithRetry provides a public protocol for executing Redis operations with retry logic
	ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error
}

// RedisManagerImpl manages Redis client connections
type RedisManagerImpl struct {
	client RedisClient
	config *sconfig.Config
	logger *zap.Logger
}

// NewRedisManager creates a new Redis connection manager with comprehensive configuration
func NewRedisManager(cfg *sconfig.Config) (RedisManager, error) {
	logger := slogger.GetLogger().With(zap.String("component", "redis_manager"))

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
	options := &goredis.Options{
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
	client := goredis.NewClient(options)

	// Test the connection with retry logic
	manager := &RedisManagerImpl{
		client: client,
		config: cfg,
		logger: logger,
	}

	// Test initial connection with retry
	if err := manager.checkConnectionWithRetry(); err != nil {
		client.Close() // nolint:errcheck
		return nil, fmt.Errorf("failed to establish Redis connection: %w", err)
	}

	logger.Info("Redis client initialized successfully")

	return manager, nil
}

// NewRedisManagerWithClient creates a new Redis manager with a provided client (for testing)
func NewRedisManagerWithClient(client RedisClient, cfg *sconfig.Config) RedisManager {
	logger := slogger.GetLogger().With(zap.String("component", "redis_manager"))

	return &RedisManagerImpl{
		client: client,
		config: cfg,
		logger: logger,
	}
}

// GetClient returns the underlying Redis client
// This should be used sparingly and only when direct access to *redis.Client is needed
func (rm *RedisManagerImpl) GetClient() *goredis.Client {
	if client, ok := rm.client.(*goredis.Client); ok {
		return client
	}
	return nil
}

// Ping performs a health check on the Redis connection with retry logic
func (rm *RedisManagerImpl) Ping(ctx context.Context) error {
	rm.logger.Debug("Performing Redis health check")

	err := rm.executeWithRetry(ctx, "health_check", func() error {
		return rm.client.Ping(ctx).Err()
	})

	if err != nil {
		rm.logger.Error("Redis health check failed after retries",
			slogger.FormatError(err),
		)
		return fmt.Errorf("redis health check failed: %w", err)
	}

	rm.logger.Debug("Redis health check successful")
	return nil
}

// Close closes the Redis connection
func (rm *RedisManagerImpl) Close() error {
	rm.logger.Info("Closing Redis connection")
	return rm.client.Close()
}

// checkConnectionWithRetry tests the Redis connection with retry logic
func (rm *RedisManagerImpl) checkConnectionWithRetry() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return rm.executeWithRetry(ctx, "connection_test", func() error {
		return rm.client.Ping(ctx).Err()
	})
}

// executeWithRetry executes a Redis operation with retry logic
func (rm *RedisManagerImpl) executeWithRetry(ctx context.Context, operation string, fn func() error) error {
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
				slogger.FormatError(err),
			)
			return fmt.Errorf("redis operation '%s' failed: %w", operation, err)
		}

		// Don't retry if we've reached max attempts
		if attempt >= maxRetries {
			rm.logger.Error("Redis operation failed after all retries",
				zap.String("operation", operation),
				zap.Int("total_attempts", attempt+1),
				zap.Int("max_retries", maxRetries+1),
				slogger.FormatError(lastErr),
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
			slogger.FormatError(err),
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
func (rm *RedisManagerImpl) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()
	errorStrLower := strings.ToLower(errorStr)

	// Network-related errors that are typically retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Network timeout errors are retryable
		if netErr.Timeout() {
			rm.logger.Debug("Detected network timeout error", slogger.FormatError(err))
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
				slogger.FormatError(err),
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
				slogger.FormatError(err),
			)
			return false
		}
	}

	// Default to retryable for unknown errors (conservative approach)
	rm.logger.Debug("Unknown error type, treating as retryable", slogger.FormatError(err))
	return true
}

// GetConnectionPoolStats returns Redis connection pool statistics
func (rm *RedisManagerImpl) GetConnectionPoolStats() *goredis.PoolStats {
	if rm.client == nil {
		return nil
	}
	stats := rm.client.PoolStats()
	return stats
}

// LogConnectionPoolStats logs the current Redis connection pool statistics
func (rm *RedisManagerImpl) LogConnectionPoolStats() {
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

// ExecuteWithRetry provides a public protocol for executing Redis operations with retry logic
// This can be used by other parts of the application that need retry functionality
func (rm *RedisManagerImpl) ExecuteWithRetry(ctx context.Context, operation string, fn func() error) error {
	return rm.executeWithRetry(ctx, operation, fn)
}
