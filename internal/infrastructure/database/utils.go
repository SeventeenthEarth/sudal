package database

import (
	"context"
	"fmt"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/redis"
	"time"

	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

// VerifyDatabaseConnectivity is a standalone utility function to verify database connectivity
// This can be called during application startup or on-demand for health checks
func VerifyDatabaseConnectivity(ctx context.Context, cfg *config.Config) error {
	logger := log.GetLogger().With(zap.String("component", "database_verification"))

	logger.Info("Starting database connectivity verification",
		zap.String("host", cfg.DB.Host),
		zap.String("port", cfg.DB.Port),
		zap.String("database", cfg.DB.Name),
		zap.String("ssl_mode", cfg.DB.SSLMode),
	)

	// Create a temporary PostgreSQL manager for verification
	pgManager, err := postgres.NewPostgresManager(cfg)
	if err != nil {
		logger.Error("Failed to create PostgreSQL manager for verification",
			log.FormatError(err),
		)
		return fmt.Errorf("database connectivity verification failed: %w", err)
	}
	defer func() {
		if closeErr := pgManager.Close(); closeErr != nil {
			logger.Warn("Failed to close PostgreSQL manager after verification",
				log.FormatError(closeErr),
			)
		}
	}()

	// Perform health check with timeout
	verificationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	healthStatus, err := pgManager.HealthCheck(verificationCtx)
	if err != nil {
		logger.Error("Database connectivity verification failed",
			log.FormatError(err),
		)
		return fmt.Errorf("database connectivity verification failed: %w", err)
	}

	logger.Info("Database connectivity verification successful",
		zap.String("status", healthStatus.Status),
		zap.String("message", healthStatus.Message),
		zap.Any("connection_stats", healthStatus.Stats),
	)

	return nil
}

// GetConnectionPoolStats returns the current connection pool statistics
// This is useful for monitoring and debugging connection pool behavior
func GetConnectionPoolStats(pgManager postgres.PostgresManager) *postgres.ConnectionStats {
	if pgManager == nil {
		return nil
	}

	// Use HealthCheck to get stats since we can't access internal fields
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus, err := pgManager.HealthCheck(ctx)
	if err != nil || healthStatus.Stats == nil {
		return nil
	}

	return healthStatus.Stats
}

// LogConnectionPoolStats logs the current connection pool statistics
// This can be called periodically to monitor pool health
func LogConnectionPoolStats(pgManager postgres.PostgresManager) {
	if pgManager == nil {
		return
	}

	stats := GetConnectionPoolStats(pgManager)
	if stats == nil {
		return
	}

	logger := log.GetLogger().With(zap.String("component", "postgres_manager"))
	logger.Info("Connection pool statistics",
		zap.Int("max_open_connections", stats.MaxOpenConnections),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)
}

// VerifyRedisConnectivity is a standalone utility function to verify Redis connectivity
// This can be called during application startup or on-demand for health checks
func VerifyRedisConnectivity(ctx context.Context, cfg *config.Config) error {
	logger := log.GetLogger().With(zap.String("component", "redis_verification"))

	logger.Info("Starting Redis connectivity verification",
		zap.String("addr", cfg.Redis.Addr),
		zap.Bool("password_set", cfg.Redis.Password != ""),
		zap.Int("db", cfg.Redis.DB),
		zap.Int("pool_size", cfg.Redis.PoolSize),
		zap.Int("max_retries", cfg.Redis.MaxRetries),
	)

	// Create a temporary Redis manager for verification
	redisManager, err := redis.NewRedisManager(cfg)
	if err != nil {
		logger.Error("Failed to create Redis manager for verification",
			log.FormatError(err),
		)
		return fmt.Errorf("redis connectivity verification failed: %w", err)
	}
	defer func() {
		if closeErr := redisManager.Close(); closeErr != nil {
			logger.Warn("Failed to close Redis manager after verification",
				log.FormatError(closeErr),
			)
		}
	}()

	// Perform health check with timeout
	verificationCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := redisManager.Ping(verificationCtx); err != nil {
		logger.Error("Redis connectivity verification failed",
			log.FormatError(err),
		)
		return fmt.Errorf("redis connectivity verification failed: %w", err)
	}

	// Log connection pool statistics
	redisManager.LogConnectionPoolStats()

	logger.Info("Redis connectivity verification successful")

	return nil
}
