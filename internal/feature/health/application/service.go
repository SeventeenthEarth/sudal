package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_health_service.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application Service

// Service defines the health check service interface
type Service interface {
	Ping(ctx context.Context) (*entity.HealthStatus, error)
	Check(ctx context.Context) (*entity.HealthStatus, error)
	CheckDatabase(ctx context.Context) (*entity.DatabaseStatus, error)
}

// service is the implementation of the health check service
// It acts as a facade for the individual use cases
type service struct {
	pingUseCase           PingUseCase
	healthCheckUseCase    HealthCheckUseCase
	databaseHealthUseCase DatabaseHealthUseCase
}

// NewService creates a new health check service
func NewService(repository repo.HealthRepository) Service {
	return &service{
		pingUseCase:           NewPingUseCase(),
		healthCheckUseCase:    NewHealthCheckUseCase(repository),
		databaseHealthUseCase: NewDatabaseHealthUseCase(repository),
	}
}

// Ping returns a simple status to indicate the service is alive
func (s *service) Ping(ctx context.Context) (*entity.HealthStatus, error) {
	return s.pingUseCase.Execute(ctx)
}

// Check performs a health check on the service
// Uses the repository to check the health of infrastructure components
func (s *service) Check(ctx context.Context) (*entity.HealthStatus, error) {
	return s.healthCheckUseCase.Execute(ctx)
}

// CheckDatabase performs a database health check
// Uses the repository to check the health of database connections
func (s *service) CheckDatabase(ctx context.Context) (*entity.DatabaseStatus, error) {
	return s.databaseHealthUseCase.Execute(ctx)
}
