//go:build wireinject
// +build wireinject

package di

import (
	"os"

	repo2 "github.com/seventeenthearth/sudal/internal/feature/health/data/repo"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"
	userConnect "github.com/seventeenthearth/sudal/internal/feature/user/protocol"

	"github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/redis"
	"github.com/seventeenthearth/sudal/internal/infrastructure/firebase"

	"github.com/google/wire"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"

	userApplication "github.com/seventeenthearth/sudal/internal/feature/user/application"
	userRepo "github.com/seventeenthearth/sudal/internal/feature/user/data/repo"
	userDomainRepo "github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"github.com/seventeenthearth/sudal/internal/infrastructure/cacheutil"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
	"go.uber.org/zap"
)

// ConfigSet is a Wire provider set for configuration
var ConfigSet = wire.NewSet(
	ProvideConfig,
)

// HealthConnectSet is a Wire provider set for Connect-go health service (gRPC only)
var HealthConnectSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	repo2.NewHealthRepository,
	wire.Bind(new(repo.HealthRepository), new(*repo2.HealthRepository)),
	application.NewPingUseCase,
	application.NewHealthCheckUseCase,
	application.NewDatabaseHealthUseCase,
	application.NewService,
	healthConnect.NewHealthManager,
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
	ProvideUserService,
	ProvideFirebaseHandler,
	userConnect.NewUserHandler,
)

// FirebaseSet is a Wire provider set for Firebase-related dependencies
var FirebaseSet = wire.NewSet(
	ProvideConfig,
	ProvideLogger,
	ProvideUserRepository,
	ProvidePostgresManager,
	ProvideFirebaseHandler,
)

// ProvideConfig provides the application configuration
func ProvideConfig() *config.Config {
	return config.GetConfig()
}

// ProvidePostgresManager provides a PostgreSQL connection manager
func ProvidePostgresManager(cfg *config.Config) (postgres.PostgresManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return postgres.NewPostgresManager(cfg)
}

// ProvideRedisManager provides a Redis connection manager
func ProvideRedisManager(cfg *config.Config) (redis.RedisManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return redis.NewRedisManager(cfg)
}

// ProvideCacheUtil provides a cache utility instance
func ProvideCacheUtil(redisManager redis.RedisManager) cacheutil.CacheUtil {
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
func ProvideUserRepository(pgManager postgres.PostgresManager, logger *zap.Logger) userDomainRepo.UserRepository {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return userRepo.NewUserRepoImpl(pgManager.GetDB(), logger)
}

// ProvideUserService provides a user application service instance
func ProvideUserService(repository userDomainRepo.UserRepository) userApplication.UserService {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return userApplication.NewService(repository)
}

// ProvideFirebaseHandler provides a Firebase handler instance
func ProvideFirebaseHandler(cfg *config.Config, userRepo userDomainRepo.UserRepository, logger *zap.Logger) (firebase.AuthVerifier, error) {
	// In test environment, return a stub handler to avoid real SDK init
	if isTestEnvironmentWire() {
		// For tests, we return nil to allow upper layers to skip auth, or tests can inject mocks explicitly.
		return nil, nil
	}

	// Use GOOGLE_APPLICATION_CREDENTIALS environment variable if set, otherwise use config
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsFile == "" {
		credentialsFile = cfg.FirebaseCredentialsJSON
	}

	return firebase.NewFirebaseHandler(credentialsFile, userRepo, logger)
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
func InitializeHealthConnectHandler() (*healthConnect.HealthManager, error) {
	wire.Build(HealthConnectSet)
	return nil, nil // Wire will fill this in
}

// InitializePostgresManager initializes and returns a PostgreSQL connection manager
func InitializePostgresManager() (postgres.PostgresManager, error) {
	wire.Build(DatabaseSet)
	return nil, nil // Wire will fill this in
}

// InitializeRedisManager initializes and returns a Redis connection manager
func InitializeRedisManager() (redis.RedisManager, error) {
	wire.Build(RedisSet)
	return nil, nil // Wire will fill this in
}

// InitializeCacheUtil initializes and returns a cache utility
func InitializeCacheUtil() (cacheutil.CacheUtil, error) {
	wire.Build(CacheSet)
	return nil, nil // Wire will fill this in
}

// InitializeUserConnectHandler initializes and returns a Connect-go user handler (gRPC only)
func InitializeUserConnectHandler() (*userConnect.UserManager, error) {
	wire.Build(UserSet)
	return nil, nil // Wire will fill this in
}

// InitializeFirebaseHandler initializes and returns a Firebase auth verifier
func InitializeFirebaseHandler() (firebase.AuthVerifier, error) {
	wire.Build(FirebaseSet)
	return nil, nil // Wire will fill this in
}

// OpenAPISet is a Wire provider set for OpenAPI-related dependencies (REST API)
var OpenAPISet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	repo2.NewHealthRepository,
	wire.Bind(new(repo.HealthRepository), new(*repo2.HealthRepository)),
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
