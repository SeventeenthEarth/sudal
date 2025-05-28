package data

import (
	"context"
	"fmt"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
)

// Implementation of the domain.Repository interface

// Repository is the implementation of the domain.Repository interface
type Repository struct {
	// Database manager for performing health checks
	dbManager *database.PostgresManager
}

// NewRepository creates a new health repository
func NewRepository(dbManager *database.PostgresManager) *Repository {
	return &Repository{
		dbManager: dbManager,
	}
}

// GetStatus retrieves the current health status
// In a real application, this would check database connections, cache, etc.
func (r *Repository) GetStatus(ctx context.Context) (*domain.Status, error) {
	// In a real application, this would perform actual health checks
	// For example:
	// - Check database connection
	// - Check cache connection
	// - Check external API dependencies

	// For now, we just return a healthy status
	return domain.HealthyStatus(), nil
}

// GetDatabaseStatus retrieves the current database health status
func (r *Repository) GetDatabaseStatus(ctx context.Context) (*domain.DatabaseStatus, error) {
	// If no database manager is available (e.g., in tests), return a mock status
	if r.dbManager == nil {
		stats := &domain.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    1,
			InUse:              0,
			Idle:               1,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		}
		return domain.HealthyDatabaseStatus("Mock database connection is healthy", stats), nil
	}

	// Perform actual database health check
	infraHealthStatus, err := r.dbManager.HealthCheck(ctx)
	if err != nil {
		return domain.UnhealthyDatabaseStatus(fmt.Sprintf("Database health check failed: %v", err)), err
	}

	// Convert infrastructure health status to domain model
	var domainStats *domain.ConnectionStats
	if infraHealthStatus.Stats != nil {
		domainStats = &domain.ConnectionStats{
			MaxOpenConnections: infraHealthStatus.Stats.MaxOpenConnections,
			OpenConnections:    infraHealthStatus.Stats.OpenConnections,
			InUse:              infraHealthStatus.Stats.InUse,
			Idle:               infraHealthStatus.Stats.Idle,
			WaitCount:          infraHealthStatus.Stats.WaitCount,
			WaitDuration:       infraHealthStatus.Stats.WaitDuration,
			MaxIdleClosed:      infraHealthStatus.Stats.MaxIdleClosed,
			MaxLifetimeClosed:  infraHealthStatus.Stats.MaxLifetimeClosed,
		}
	}

	return domain.HealthyDatabaseStatus(infraHealthStatus.Message, domainStats), nil
}
