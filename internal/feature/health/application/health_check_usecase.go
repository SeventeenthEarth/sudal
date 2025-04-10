package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_health_check_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application HealthCheckUseCase

// HealthCheckUseCase defines the interface for the health check functionality
type HealthCheckUseCase interface {
	Execute(ctx context.Context) (*domain.Status, error)
}

// healthCheckUseCase implements the HealthCheckUseCase interface
type healthCheckUseCase struct {
	repo domain.Repository
}

// NewHealthCheckUseCase creates a new health check use case
func NewHealthCheckUseCase(repo domain.Repository) HealthCheckUseCase {
	return &healthCheckUseCase{
		repo: repo,
	}
}

// Execute performs a health check on the service
func (uc *healthCheckUseCase) Execute(ctx context.Context) (*domain.Status, error) {
	// Get health status from repository
	return uc.repo.GetStatus(ctx)
}
