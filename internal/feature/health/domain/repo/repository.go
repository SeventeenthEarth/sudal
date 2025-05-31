package repo

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../../mocks/mock_health_repository.go -package=mocks -mock_names=HealthRepository=MockHealthRepository github.com/seventeenthearth/sudal/internal/feature/health/domain/repo HealthRepository

// HealthRepository defines the protocol for health data access
// This is defined in the domain layer to maintain the dependency rule
// where data layer depends on domain, not the other way around
type HealthRepository interface {
	GetStatus(ctx context.Context) (*entity.HealthStatus, error)
	GetDatabaseStatus(ctx context.Context) (*entity.DatabaseStatus, error)
}
