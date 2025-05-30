//go:build wireinject
// +build wireinject

package di

import (
	"os"

	"github.com/google/wire"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"

	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	userRepo "github.com/seventeenthearth/sudal/internal/feature/user/data/repo"
	userDomainRepo "github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	userConnect "github.com/seventeenthearth/sudal/internal/feature/user/interface/connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
	"go.uber.org/zap"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../../../mocks/mock_di_initializer.go -package=mocks github.com/seventeenthearth/sudal/internal/infrastructure/di DatabaseHealthInitializer

// DatabaseHealthInitializer interface for dependency injection initialization
type DatabaseHealthInitializer interface {
	InitializeDatabaseHealthHandler() (*DatabaseHealthHandler, error)
}

// ConfigSet is a Wire provider set for configuration
var ConfigSet = wire.NewSet(
	ProvideConfig,
)

// HealthConnectSet is a Wire provider set for Connect-go health service (gRPC only)
var HealthConnectSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	data.NewRepository,
	wire.Bind(new(repo.HealthRepository), new(*data.HealthRepository)),
	application.NewPingUseCase,
	application.NewHealthCheckUseCase,
	application.NewDatabaseHealthUseCase,
	application.NewService,
	healthConnect.NewHealthServiceHandler,
)

// DatabaseSet is a Wire provider set for database-related dependencies
var DatabaseSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
)

// DatabaseHealthSet is a Wire provider set for database health check handler
var DatabaseHealthSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	NewDatabaseHealthHandler,
)

// RedisSet is a Wire provider set for Redis-related dependencies
var RedisSet = wire.NewSet(
	ProvideConfig,
	ProvideRedisManager,
)

// CacheSet is a Wire provider set for cache-related dependencies
var CacheSet = wire.NewSet(
	ProvideConfig,
	ProvideRedisManager,
	ProvideCacheUtil,
)

// UserSet is a Wire provider set for user-related dependencies (gRPC only)
var UserSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	ProvideLogger,
	ProvideUserRepository,
	userConnect.NewUserService,
)

// ProvideConfig provides the application configuration
func ProvideConfig() *config.Config {
	return config.GetConfig()
}

// ProvidePostgresManager provides a PostgreSQL connection manager
func ProvidePostgresManager(cfg *config.Config) (database.PostgresManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return database.NewPostgresManager(cfg)
}

// ProvideRedisManager provides a Redis connection manager
func ProvideRedisManager(cfg *config.Config) (database.RedisManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return database.NewRedisManager(cfg)
}

// ProvideCacheUtil provides a cache utility instance
func ProvideCacheUtil(redisManager database.RedisManager) cacheutil.CacheUtil {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return cacheutil.NewCacheUtil(redisManager)
}

// ProvideLogger provides a logger instance
func ProvideLogger() *zap.Logger {
	return log.GetLogger()
}

// ProvideUserRepository provides a user repository instance
func ProvideUserRepository(pgManager database.PostgresManager, logger *zap.Logger) userDomainRepo.UserRepository {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return userRepo.NewUserRepoImpl(pgManager.GetDB(), logger)
}

// isTestEnvironmentWire checks if we're running in a test environment for wire
func isTestEnvironmentWire() bool {
	// Check environment variables that indicate test mode
	goTest := os.Getenv("GO_TEST")
	ginkgoTest := os.Getenv("GINKGO_TEST")

	if goTest == "1" || ginkgoTest == "1" {
		return true
	}

	// Check if config indicates test environment
	cfg := config.GetConfig()
	if cfg != nil {
		if cfg.AppEnv == "test" || cfg.Environment == "test" {
			return true
		}
	}

	return false
}

// InitializeHealthConnectHandler initializes and returns a Connect-go health service handler (gRPC only)
func InitializeHealthConnectHandler() (*healthConnect.HealthServiceHandler, error) {
	wire.Build(HealthConnectSet)
	return nil, nil // Wire will fill this in
}

// InitializePostgresManager initializes and returns a PostgreSQL connection manager
func InitializePostgresManager() (database.PostgresManager, error) {
	wire.Build(DatabaseSet)
	return nil, nil // Wire will fill this in
}

// InitializeDatabaseHealthHandler initializes and returns a database health handler
func InitializeDatabaseHealthHandler() (*DatabaseHealthHandler, error) {
	wire.Build(DatabaseHealthSet)
	return nil, nil // Wire will fill this in
}

// InitializeRedisManager initializes and returns a Redis connection manager
func InitializeRedisManager() (database.RedisManager, error) {
	wire.Build(RedisSet)
	return nil, nil // Wire will fill this in
}

// InitializeCacheUtil initializes and returns a cache utility
func InitializeCacheUtil() (cacheutil.CacheUtil, error) {
	wire.Build(CacheSet)
	return nil, nil // Wire will fill this in
}

// InitializeUserConnectHandler initializes and returns a Connect-go user service handler (gRPC only)
func InitializeUserConnectHandler() (*userConnect.UserService, error) {
	wire.Build(UserSet)
	return nil, nil // Wire will fill this in
}

// DefaultDatabaseHealthInitializer is the default implementation of DatabaseHealthInitializer
type DefaultDatabaseHealthInitializer struct{}

// NewDefaultDatabaseHealthInitializer creates a new default database health initializer
func NewDefaultDatabaseHealthInitializer() DatabaseHealthInitializer {
	return &DefaultDatabaseHealthInitializer{}
}

// InitializeDatabaseHealthHandler implements DatabaseHealthInitializer interface
func (d *DefaultDatabaseHealthInitializer) InitializeDatabaseHealthHandler() (*DatabaseHealthHandler, error) {
	return InitializeDatabaseHealthHandler()
}

// OpenAPISet is a Wire provider set for OpenAPI-related dependencies (REST API)
var OpenAPISet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	data.NewRepository,
	wire.Bind(new(repo.HealthRepository), new(*data.HealthRepository)),
	application.NewPingUseCase,
	application.NewHealthCheckUseCase,
	application.NewDatabaseHealthUseCase,
	application.NewService,
	NewOpenAPIHandler,
)

// NewOpenAPIHandler creates a new OpenAPI handler
func NewOpenAPIHandler(service application.HealthService) *openapi.OpenAPIHandler {
	return openapi.NewOpenAPIHandler(service)
}

// InitializeOpenAPIHandler initializes and returns an OpenAPI handler (REST API)
func InitializeOpenAPIHandler() (*openapi.OpenAPIHandler, error) {
	wire.Build(OpenAPISet)
	return nil, nil // Wire will fill this in
}

// InitializeSwaggerHandler initializes and returns a Swagger UI handler
func InitializeSwaggerHandler() *openapi.SwaggerHandler {
	return openapi.NewSwaggerHandler("api/openapi.yaml")
}
