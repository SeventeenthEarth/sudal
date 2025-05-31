package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_database_health_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application DatabaseHealthUseCase

// DatabaseHealthUseCase defines the protocol for the database health check functionality
type DatabaseHealthUseCase interface {
	Execute(ctx context.Context) (*entity.DatabaseStatus, error)
}

// databaseHealthUseCase implements the DatabaseHealthUseCase protocol
type databaseHealthUseCase struct {
	repo repo.HealthRepository
}

// NewDatabaseHealthUseCase creates a new database health check use case
func NewDatabaseHealthUseCase(repository repo.HealthRepository) DatabaseHealthUseCase {
	return &databaseHealthUseCase{
		repo: repository,
	}
}

// Execute performs a database health check
func (uc *databaseHealthUseCase) Execute(ctx context.Context) (*entity.DatabaseStatus, error) {
	// Get database health status from repository
	return uc.repo.GetDatabaseStatus(ctx)
}
