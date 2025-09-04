//go:build wireinject
// +build wireinject

package di

import (
	"context"
	"fmt"
	"os"

	repo2 "github.com/seventeenthearth/sudal/internal/feature/health/data/repo"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"
	userConnect "github.com/seventeenthearth/sudal/internal/feature/user/protocol"

	firebaseadm "firebase.google.com/go/v4"

	"github.com/google/wire"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"

	userApplication "github.com/seventeenthearth/sudal/internal/feature/user/application"
	userRepo "github.com/seventeenthearth/sudal/internal/feature/user/data/repo"
	userDomainRepo "github.com/seventeenthearth/sudal/internal/feature/user/domain/repo"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
	scache "github.com/seventeenthearth/sudal/internal/service/cache"
	sconfig "github.com/seventeenthearth/sudal/internal/service/config"
	"github.com/seventeenthearth/sudal/internal/service/firebaseauth"
	slogger "github.com/seventeenthearth/sudal/internal/service/logger"
	spostgres "github.com/seventeenthearth/sudal/internal/service/postgres"
	sredis "github.com/seventeenthearth/sudal/internal/service/redis"
	ssql "github.com/seventeenthearth/sudal/internal/service/sql"
	ssqlpg "github.com/seventeenthearth/sudal/internal/service/sql/postgres"
	"go.uber.org/zap"
	"google.golang.org/api/option"
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
	ProvideRedisKV,
	ProvideCacheUtil,
)

// UserSet is a Wire provider set for user-related dependencies (gRPC only)
var UserSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	ProvideLogger,
	ProvideSQLExecutor, // Provide minimal SQL surface for repos
	ProvideUserRepository,
	ProvideUserService,
	ProvideTokenVerifier,
	userConnect.NewUserHandler,
)

// FirebaseSet is a Wire provider set for Firebase-related dependencies
// FirebaseSet removed in Stage C (replaced by TokenVerifier)

// ProvideConfig provides the application configuration
func ProvideConfig() *sconfig.Config {
	return sconfig.GetConfig()
}

// ProvidePostgresManager provides a PostgreSQL connection manager
func ProvidePostgresManager(cfg *sconfig.Config) (spostgres.PostgresManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return spostgres.NewPostgresManager(cfg)
}

// ProvideRedisManager provides a Redis connection manager
func ProvideRedisManager(cfg *sconfig.Config) (sredis.RedisManager, error) {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil, nil
	}
	return sredis.NewRedisManager(cfg)
}

// ProvideCacheUtil provides a cache utility instance
func ProvideCacheUtil(kv sredis.KV) scache.CacheUtil {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return scache.NewCacheUtil(kv)
}

// ProvideRedisKV adapts RedisManager's client into the service KV interface
func ProvideRedisKV(manager sredis.RedisManager) sredis.KV {
	if isTestEnvironmentWire() || manager == nil {
		return nil
	}
	client := manager.GetClient()
	if client == nil {
		return nil
	}
	return sredis.NewKVFromClient(client)
}

// ProvideLogger provides a logger instance
func ProvideLogger() *zap.Logger {
	return slogger.GetLogger()
}

// ProvideUserRepository provides a user repository instance using the minimal SQL executor
func ProvideUserRepository(exec ssql.Executor, logger *zap.Logger) userDomainRepo.UserRepository {
	// In tests, or when executor is not available, return nil to allow mocks
	if isTestEnvironmentWire() || exec == nil {
		return nil
	}
	return userRepo.NewUserRepo(exec, logger)
}

// ProvideSQLExecutor provides a thin SQL executor backed by *sql.DB
// Note: This only wires the constructor; repositories will be migrated to depend on this in later PRs.
func ProvideSQLExecutor(pgManager spostgres.PostgresManager) ssql.Executor {
	if isTestEnvironmentWire() || pgManager == nil {
		return nil
	}
	exec, _ := ProvideSQLExecutorAndTransactor(pgManager)
	return exec
}

// ProvideSQLTransactor provides a transactor for beginning transactions
func ProvideSQLTransactor(pgManager spostgres.PostgresManager) ssql.Transactor {
	if isTestEnvironmentWire() || pgManager == nil {
		return nil
	}
	_, tx := ProvideSQLExecutorAndTransactor(pgManager)
	return tx
}

// ProvideSQLExecutorAndTransactor provides both an executor and transactor.
func ProvideSQLExecutorAndTransactor(pgManager spostgres.PostgresManager) (ssql.Executor, ssql.Transactor) {
	if isTestEnvironmentWire() || pgManager == nil {
		return nil, nil
	}
	return ssqlpg.NewFromDB(pgManager.GetDB())
}

// ProvideUserService provides a user application service instance
func ProvideUserService(repository userDomainRepo.UserRepository) userApplication.UserService {
	// Check if we're in test environment and return nil to use mock
	if isTestEnvironmentWire() {
		return nil
	}
	return userApplication.NewService(repository)
}

// ProvideFirebaseHandler removed in Stage C

// ProvideTokenVerifier provides a Firebase-based TokenVerifier implementation
func ProvideTokenVerifier(cfg *sconfig.Config, logger *zap.Logger) (firebaseauth.TokenVerifier, error) {
	if isTestEnvironmentWire() {
		return nil, nil
	}

	// Resolve credentials path
	credentialsFile := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsFile == "" {
		credentialsFile = cfg.FirebaseCredentialsJSON
	}
	if credentialsFile == "" {
		err := fmt.Errorf("firebase credentials file not provided via GOOGLE_APPLICATION_CREDENTIALS or config")
		logger.Error("Failed to initialize TokenVerifier", zap.Error(err))
		return nil, err
	}

	// Initialize Firebase app and auth client
	app, err := firebaseadm.NewApp(context.Background(), nil, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		logger.Error("Failed to initialize Firebase app for TokenVerifier",
			zap.String("credentials_file", credentialsFile), zap.Error(err))
		return nil, err
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		logger.Error("Failed to get Firebase Auth client for TokenVerifier", zap.Error(err))
		return nil, err
	}

	return firebaseauth.NewFirebaseTokenVerifier(client, logger), nil
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
	cfg := sconfig.GetConfig()
	if cfg != nil {
		if cfg.AppEnv == "test" {
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
func InitializePostgresManager() (spostgres.PostgresManager, error) {
	wire.Build(DatabaseSet)
	return nil, nil // Wire will fill this in
}

// InitializeRedisManager initializes and returns a Redis connection manager
func InitializeRedisManager() (sredis.RedisManager, error) {
	wire.Build(RedisSet)
	return nil, nil // Wire will fill this in
}

// InitializeCacheUtil initializes and returns a cache utility
func InitializeCacheUtil() (scache.CacheUtil, error) {
	wire.Build(CacheSet)
	return nil, nil // Wire will fill this in
}

// InitializeUserConnectHandler initializes and returns a Connect-go user handler (gRPC only)
func InitializeUserConnectHandler() (*userConnect.UserManager, error) {
	wire.Build(UserSet)
	return nil, nil // Wire will fill this in
}

// InitializeFirebaseHandler removed in Stage C

// TokenVerifierSet is a Wire provider set for TokenVerifier only (for middleware chains)
var TokenVerifierSet = wire.NewSet(
	ProvideConfig,
	ProvideLogger,
	ProvideTokenVerifier,
)

// InitializeTokenVerifier initializes and returns a TokenVerifier
func InitializeTokenVerifier() (firebaseauth.TokenVerifier, error) {
	wire.Build(TokenVerifierSet)
	return nil, nil // Wire will fill this in
}

// UserServiceSet is a Wire provider set for building only the user service
var UserServiceSet = wire.NewSet(
	ProvideConfig,
	ProvidePostgresManager,
	ProvideLogger,
	ProvideSQLExecutor,
	ProvideUserRepository,
	ProvideUserService,
)

// InitializeUserService initializes and returns a user application service
func InitializeUserService() (userApplication.UserService, error) {
	wire.Build(UserServiceSet)
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
