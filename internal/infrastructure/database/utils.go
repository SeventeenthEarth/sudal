package database

import (
	"context"
	"fmt"
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
	pgManager, err := NewPostgresManager(cfg)
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
func GetConnectionPoolStats(pgManager *PostgresManager) *ConnectionStats {
	if pgManager == nil || pgManager.db == nil {
		return nil
	}

	stats := pgManager.db.Stats()
	return &ConnectionStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		WaitDuration:       stats.WaitDuration,
		MaxIdleClosed:      stats.MaxIdleClosed,
		MaxLifetimeClosed:  stats.MaxLifetimeClosed,
	}
}

// LogConnectionPoolStats logs the current connection pool statistics
// This can be called periodically to monitor pool health
func LogConnectionPoolStats(pgManager *PostgresManager) {
	if pgManager == nil {
		return
	}

	stats := GetConnectionPoolStats(pgManager)
	if stats == nil {
		return
	}

	pgManager.logger.Info("Connection pool statistics",
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
