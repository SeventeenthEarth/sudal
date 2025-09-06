package server

import (
	"fmt"
	"net/http"

	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/gen/go/quiz/v1/quizv1connect"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"
	quizConnect "github.com/seventeenthearth/sudal/internal/feature/quiz/protocol"
	userConnect "github.com/seventeenthearth/sudal/internal/feature/user/protocol"
	"github.com/seventeenthearth/sudal/internal/infrastructure/apispec"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
)

// ServiceRegistry manages all service handlers and their initialization
type ServiceRegistry struct {
	// gRPC service handlers
	HealthHandler *healthConnect.HealthManager
	QuizHandler   *quizConnect.QuizManager
	UserHandler   *userConnect.UserManager

	// REST service handlers
	OpenAPIHandler *openapi.OpenAPIHandler
	SwaggerHandler *openapi.SwaggerHandler
}

// NewServiceRegistry creates and initializes all service handlers
func NewServiceRegistry() (*ServiceRegistry, error) {
	// Initialize Connect-go health service handler (gRPC only)
	healthHandler, err := di.InitializeHealthConnectHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize health connect handler: %w", err)
	}

	// Initialize Connect-go user service handler (gRPC only)
	userHandler, err := di.InitializeUserConnectHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user connect handler: %w", err)
	}

	// Initialize Connect-go quiz service handler (gRPC only)
	quizHandler, err := di.InitializeQuizConnectHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize quiz connect handler: %w", err)
	}

	// Initialize OpenAPI handler for REST endpoints
	openAPIHandler, err := di.InitializeOpenAPIHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAPI handler: %w", err)
	}

	// Initialize Swagger UI handler
	swaggerHandler := di.InitializeSwaggerHandler()

	return &ServiceRegistry{
		HealthHandler:  healthHandler,
		QuizHandler:    quizHandler,
		UserHandler:    userHandler,
		OpenAPIHandler: openAPIHandler,
		SwaggerHandler: swaggerHandler,
	}, nil
}

// RouteRegistrar handles route registration with appropriate middleware chains
type RouteRegistrar struct {
	mux      *http.ServeMux
	chains   *ServiceChains
	registry *ServiceRegistry
}

// NewRouteRegistrar creates a new route registrar
func NewRouteRegistrar(mux *http.ServeMux, chains *ServiceChains, registry *ServiceRegistry) *RouteRegistrar {
	return &RouteRegistrar{
		mux:      mux,
		chains:   chains,
		registry: registry,
	}
}

// RegisterRESTRoutes registers all REST API routes
func (rr *RouteRegistrar) RegisterRESTRoutes() error {
	// Register OpenAPI routes for REST endpoints (/api/ping, /api/healthz, /api/health/database)
	openAPIServer, err := openapi.NewServer(rr.registry.OpenAPIHandler)
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI server: %w", err)
	}

	// Apply public HTTP middleware chain to REST endpoints
	restHandler := rr.chains.PublicHTTP.Apply(openAPIServer)
	rr.mux.Handle("/api/", restHandler)

	// Register Swagger UI routes (no middleware needed)
	rr.mux.HandleFunc("/docs", rr.registry.SwaggerHandler.ServeSwaggerUI)
	rr.mux.HandleFunc("/docs/", rr.registry.SwaggerHandler.ServeSwaggerUI)
	rr.mux.HandleFunc("/api/openapi.yaml", rr.registry.SwaggerHandler.ServeOpenAPISpec)

	return nil
}

// RegisterGRPCRoutes registers all gRPC service routes
func (rr *RouteRegistrar) RegisterGRPCRoutes() {
	// Register Health Service (public, gRPC-only)
	rr.registerHealthService()

	// Register Quiz Service (selective authentication, gRPC-only)
	rr.registerQuizService()

	// Register User Service (selective authentication, gRPC-only)
	rr.registerUserService()
}

// registerHealthService registers the health service with public access
func (rr *RouteRegistrar) registerHealthService() {
	// Create health service handler without authentication
	healthPath, healthHTTPHandler := healthv1connect.NewHealthServiceHandler(
		rr.registry.HealthHandler,
		rr.chains.PublicGRPC.ToConnectOptions()...,
	)

	// Apply gRPC-only HTTP middleware chain
	healthHandler := rr.chains.GRPCOnlyHTTP.Apply(healthHTTPHandler)
	rr.mux.Handle(healthPath, healthHandler)
}

// registerUserService registers the user service with selective authentication
func (rr *RouteRegistrar) registerUserService() {
	// Create user service handler with selective authentication
	userPath, userHTTPHandler := userv1connect.NewUserServiceHandler(
		rr.registry.UserHandler,
		rr.chains.SelectiveGRPC.ToConnectOptions()...,
	)

	// Apply gRPC-only HTTP middleware chain
	userHandler := rr.chains.GRPCOnlyHTTP.Apply(userHTTPHandler)
	rr.mux.Handle(userPath, userHandler)
}

// registerQuizService registers the quiz service with selective authentication
func (rr *RouteRegistrar) registerQuizService() {
	// Create quiz service handler with selective authentication
	quizPath, quizHTTPHandler := quizv1connect.NewQuizServiceHandler(
		rr.registry.QuizHandler,
		rr.chains.SelectiveGRPC.ToConnectOptions()...,
	)

	// Apply gRPC-only HTTP middleware chain
	quizHandler := rr.chains.GRPCOnlyHTTP.Apply(quizHTTPHandler)
	rr.mux.Handle(quizPath, quizHandler)
}

// ServiceConfiguration holds the configuration for service chains
type ServiceConfiguration struct {
	// Define which procedures require authentication
	ProtectedProcedures []string
}

// GetDefaultServiceConfiguration returns the default service configuration
func GetDefaultServiceConfiguration() *ServiceConfiguration {
	return &ServiceConfiguration{
		ProtectedProcedures: apispec.ProtectedProcedures(),
	}
}
