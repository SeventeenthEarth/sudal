//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

// HealthSet is a Wire provider set for health-related dependencies
var HealthSet = wire.NewSet(
	data.NewRepository,
	wire.Bind(new(domain.Repository), new(*data.Repository)),
	application.NewPingUseCase,
	application.NewHealthCheckUseCase,
	application.NewService,
	healthInterface.NewHandler,
)

// InitializeHealthHandler initializes and returns a health handler with all its dependencies
func InitializeHealthHandler() *healthInterface.Handler {
	wire.Build(HealthSet)
	return nil // Wire will fill this in
}
