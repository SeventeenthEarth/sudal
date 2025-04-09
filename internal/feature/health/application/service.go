package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

//go:generate mockgen -destination=../../../mocks/mock_health_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application Service

// Service defines the health check service interface
type Service interface {
	Ping(ctx context.Context) (*domain.Status, error)
	Check(ctx context.Context) (*domain.Status, error)
}

// service is the implementation of the health check service
// It acts as a facade for the individual use cases
type service struct {
	pingUseCase        PingUseCase
	healthCheckUseCase HealthCheckUseCase
}

// NewService creates a new health check service
func NewService(repo domain.Repository) Service {
	return &service{
		pingUseCase:        NewPingUseCase(),
		healthCheckUseCase: NewHealthCheckUseCase(repo),
	}
}

// Ping returns a simple status to indicate the service is alive
func (s *service) Ping(ctx context.Context) (*domain.Status, error) {
	return s.pingUseCase.Execute(ctx)
}

// Check performs a health check on the service
// Uses the repository to check the health of infrastructure components
func (s *service) Check(ctx context.Context) (*domain.Status, error) {
	return s.healthCheckUseCase.Execute(ctx)
}
