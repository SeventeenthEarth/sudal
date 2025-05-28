package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_database_health_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application DatabaseHealthUseCase

// DatabaseHealthUseCase defines the interface for the database health check functionality
type DatabaseHealthUseCase interface {
	Execute(ctx context.Context) (*domain.DatabaseStatus, error)
}

// databaseHealthUseCase implements the DatabaseHealthUseCase interface
type databaseHealthUseCase struct {
	repo domain.Repository
}

// NewDatabaseHealthUseCase creates a new database health check use case
func NewDatabaseHealthUseCase(repo domain.Repository) DatabaseHealthUseCase {
	return &databaseHealthUseCase{
		repo: repo,
	}
}

// Execute performs a database health check
func (uc *databaseHealthUseCase) Execute(ctx context.Context) (*domain.DatabaseStatus, error) {
	// Get database health status from repository
	return uc.repo.GetDatabaseStatus(ctx)
}
