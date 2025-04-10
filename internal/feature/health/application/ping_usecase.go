package application

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_ping_usecase.go -package=mocks github.com/seventeenthearth/sudal/internal/feature/health/application PingUseCase

// PingUseCase defines the interface for the ping functionality
type PingUseCase interface {
	Execute(ctx context.Context) (*domain.Status, error)
}

// pingUseCase implements the PingUseCase interface
type pingUseCase struct {
	// No dependencies needed for this simple use case
}

// NewPingUseCase creates a new ping use case
func NewPingUseCase() PingUseCase {
	return &pingUseCase{}
}

// Execute returns a simple status to indicate the service is alive
func (uc *pingUseCase) Execute(ctx context.Context) (*domain.Status, error) {
	return domain.OkStatus(), nil
}
