package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_postgres_manager.go -package=mocks github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres PostgresManager

// PostgresManager defines the interface for PostgreSQL database connection management
type PostgresManager interface {
	// GetDB returns the underlying database connection
	GetDB() *sql.DB
	// Ping performs a health check on the database connection
	Ping(ctx context.Context) error
	// HealthCheck performs a comprehensive health check including connection stats
	HealthCheck(ctx context.Context) (*HealthStatus, error)
	// Close closes the database connection pool
	Close() error
}

// PostgresManagerImpl manages PostgreSQL database connections and connection pooling
type PostgresManagerImpl struct {
	db     *sql.DB
	config *config.Config
	logger *zap.Logger
}

// NewPostgresManager creates a new PostgreSQL connection manager with connection pooling
func NewPostgresManager(cfg *config.Config) (PostgresManager, error) {
	logger := log.GetLogger().With(zap.String("component", "postgres_manager"))

	if cfg.DB.DSN == "" {
		return nil, fmt.Errorf("database DSN is required")
	}

	logger.Info("Initializing PostgreSQL connection pool",
		zap.String("host", cfg.DB.Host),
		zap.String("port", cfg.DB.Port),
		zap.String("database", cfg.DB.Name),
		zap.String("ssl_mode", cfg.DB.SSLMode),
		zap.Int("max_open_conns", cfg.DB.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.DB.MaxIdleConns),
		zap.Int("conn_max_lifetime_seconds", cfg.DB.ConnMaxLifetimeSeconds),
		zap.Int("conn_max_idle_time_seconds", cfg.DB.ConnMaxIdleTimeSeconds),
		zap.Int("connect_timeout_seconds", cfg.DB.ConnectTimeoutSeconds),
	)

	// Create database connection with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.DB.ConnectTimeoutSeconds)*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		logger.Error("Failed to open database connection",
			log.FormatError(err),
		)
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool settings
	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnMaxLifetimeSeconds) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(cfg.DB.ConnMaxIdleTimeSeconds) * time.Second)

	// Test the connection
	if err := db.PingContext(ctx); err != nil {
		logger.Error("Failed to ping database",
			log.FormatError(err),
		)
		db.Close() // nolint:errcheck
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection pool initialized successfully")

	return &PostgresManagerImpl{
		db:     db,
		config: cfg,
		logger: logger,
	}, nil
}

// GetDB returns the underlying database connection
// This should be used sparingly and only when direct access to *sql.DB is needed
func (pm *PostgresManagerImpl) GetDB() *sql.DB {
	return pm.db
}

// Ping performs a health check on the database connection
func (pm *PostgresManagerImpl) Ping(ctx context.Context) error {
	pm.logger.Debug("Performing database health check")

	if err := pm.db.PingContext(ctx); err != nil {
		pm.logger.Error("Database health check failed",
			log.FormatError(err),
		)
		return fmt.Errorf("database health check failed: %w", err)
	}

	pm.logger.Debug("Database health check successful")
	return nil
}

// HealthCheck performs a comprehensive health check including connection stats
func (pm *PostgresManagerImpl) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	pm.logger.Debug("Performing comprehensive database health check")

	// Perform basic ping
	if err := pm.Ping(ctx); err != nil {
		return &HealthStatus{
			Status:  "unhealthy",
			Message: err.Error(),
		}, err
	}

	// Get connection pool statistics
	stats := pm.db.Stats()

	healthStatus := &HealthStatus{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Stats: &ConnectionStats{
			MaxOpenConnections: stats.MaxOpenConnections,
			OpenConnections:    stats.OpenConnections,
			InUse:              stats.InUse,
			Idle:               stats.Idle,
			WaitCount:          stats.WaitCount,
			WaitDuration:       stats.WaitDuration,
			MaxIdleClosed:      stats.MaxIdleClosed,
			MaxLifetimeClosed:  stats.MaxLifetimeClosed,
		},
	}

	pm.logger.Debug("Comprehensive database health check successful",
		zap.Int("max_open_connections", stats.MaxOpenConnections),
		zap.Int("open_connections", stats.OpenConnections),
		zap.Int("in_use", stats.InUse),
		zap.Int("idle", stats.Idle),
		zap.Int64("wait_count", stats.WaitCount),
		zap.Duration("wait_duration", stats.WaitDuration),
		zap.Int64("max_idle_closed", stats.MaxIdleClosed),
		zap.Int64("max_lifetime_closed", stats.MaxLifetimeClosed),
	)

	return healthStatus, nil
}

// Close closes the database connection pool
func (pm *PostgresManagerImpl) Close() error {
	pm.logger.Info("Closing PostgreSQL connection pool")

	if pm.db != nil {
		if err := pm.db.Close(); err != nil {
			pm.logger.Error("Failed to close database connection pool",
				log.FormatError(err),
			)
			return fmt.Errorf("failed to close database connection pool: %w", err)
		}
	}

	pm.logger.Info("PostgreSQL connection pool closed successfully")
	return nil
}

// HealthStatus represents the health status of the database connection
type HealthStatus struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Stats   *ConnectionStats `json:"stats,omitempty"`
}

// ConnectionStats represents database connection pool statistics
type ConnectionStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}
