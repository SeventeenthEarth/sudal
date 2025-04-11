//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/seventeenthearth/sudal/gen/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
)

// ConfigSet is a Wire provider set for configuration
var ConfigSet = wire.NewSet(
	ProvideConfig,
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

// HealthConnectSet is a Wire provider set for Connect-go health service
var HealthConnectSet = wire.NewSet(
	data.NewRepository,
	wire.Bind(new(domain.Repository), new(*data.Repository)),
	application.NewPingUseCase,
	application.NewHealthCheckUseCase,
	application.NewService,
	healthConnect.NewHealthServiceHandler,
	wire.Bind(new(healthv1connect.HealthServiceHandler), new(*healthConnect.HealthServiceHandler)),
)

// ProvideConfig provides the application configuration
func ProvideConfig() *config.Config {
	return config.GetConfig()
}

// InitializeHealthHandler initializes and returns a health handler with all its dependencies
func InitializeHealthHandler() *healthInterface.Handler {
	wire.Build(HealthSet)
	return nil // Wire will fill this in
}

// InitializeHealthConnectHandler initializes and returns a Connect-go health service handler
func InitializeHealthConnectHandler() healthv1connect.HealthServiceHandler {
	wire.Build(HealthConnectSet)
	return nil // Wire will fill this in
}
