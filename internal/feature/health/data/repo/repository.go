package repo

import (
	"context"
	"fmt"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

// Implementation of the repo.HealthRepository interface

// HealthRepository is the implementation of the repo.HealthRepository interface
type HealthRepository struct {
	// Database manager for performing health checks
	dbManager postgres.PostgresManager
}

// NewHealthRepository creates a new health repository
func NewHealthRepository(dbManager postgres.PostgresManager) *HealthRepository {
	return &HealthRepository{
		dbManager: dbManager,
	}
}

// GetStatus retrieves the current health status
// In a real application, this would check database connections, cache, etc.
func (r *HealthRepository) GetStatus(ctx context.Context) (*entity.HealthStatus, error) {
	// In a real application, this would perform actual health checks
	// For example:
	// - Check database connection
	// - Check cache connection
	// - Check external API dependencies

	// For now, we just return a healthy status
	return entity.HealthyStatus(), nil
}

// GetDatabaseStatus retrieves the current database health status
func (r *HealthRepository) GetDatabaseStatus(ctx context.Context) (*entity.DatabaseStatus, error) {
	// If no database manager is available (e.g., in tests), return a mock status
	if r.dbManager == nil {
		stats := &entity.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    1,
			InUse:              0,
			Idle:               1,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		}
		return entity.HealthyDatabaseStatus("Mock database connection is healthy", stats), nil
	}

	// Perform actual database health check
	infraHealthStatus, err := r.dbManager.HealthCheck(ctx)
	if err != nil {
		return entity.UnhealthyDatabaseStatus(fmt.Sprintf("Database health check failed: %v", err)), err
	}

	// Convert infrastructure health status to domain model
	var domainStats *entity.ConnectionStats
	if infraHealthStatus.Stats != nil {
		domainStats = &entity.ConnectionStats{
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

	return entity.NewDatabaseStatus(infraHealthStatus.Status, infraHealthStatus.Message, domainStats), nil
}
