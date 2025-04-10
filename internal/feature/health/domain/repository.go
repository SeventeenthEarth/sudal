package domain

import "context"

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_health_repository.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/domain Repository

// Repository defines the interface for health data access
// This is defined in the domain layer to maintain the dependency rule
// where data layer depends on domain, not the other way around
type Repository interface {
	GetStatus(ctx context.Context) (*Status, error)
}
