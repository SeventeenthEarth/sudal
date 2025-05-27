package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

// RedisManager manages Redis client connections
type RedisManager struct {
	client *redis.Client
	config *config.Config
	logger *zap.Logger
}

// NewRedisManager creates a new Redis connection manager
func NewRedisManager(cfg *config.Config) (*RedisManager, error) {
	logger := log.GetLogger().With(zap.String("component", "redis_manager"))

	if cfg.RedisAddr == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	logger.Info("Initializing Redis client",
		zap.String("addr", cfg.RedisAddr),
		zap.Bool("password_set", cfg.RedisPassword != ""),
	)

	// Create Redis client options
	options := &redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0, // Use default DB
	}

	// Create Redis client
	client := redis.NewClient(options)

	// Test the connection with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to ping Redis server",
			log.FormatError(err),
		)
		client.Close()
		return nil, fmt.Errorf("failed to ping Redis server: %w", err)
	}

	logger.Info("Redis client initialized successfully")

	return &RedisManager{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

// GetClient returns the underlying Redis client
// This should be used sparingly and only when direct access to *redis.Client is needed
func (rm *RedisManager) GetClient() *redis.Client {
	return rm.client
}

// Ping performs a health check on the Redis connection
func (rm *RedisManager) Ping(ctx context.Context) error {
	rm.logger.Debug("Performing Redis health check")

	if err := rm.client.Ping(ctx).Err(); err != nil {
		rm.logger.Error("Redis health check failed",
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
