package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_health_check_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application HealthCheckUseCase

// HealthCheckUseCase defines the interface for the health check functionality
type HealthCheckUseCase interface {
	Execute(ctx context.Context) (*entity.HealthStatus, error)
}

// healthCheckUseCase implements the HealthCheckUseCase interface
type healthCheckUseCase struct {
	repo repo.HealthRepository
}

// NewHealthCheckUseCase creates a new health check use case
func NewHealthCheckUseCase(repository repo.HealthRepository) HealthCheckUseCase {
	return &healthCheckUseCase{
		repo: repository,
	}
}

// Execute performs a health check on the service
func (uc *healthCheckUseCase) Execute(ctx context.Context) (*entity.HealthStatus, error) {
	// Get health status from repository
	return uc.repo.GetStatus(ctx)
}
