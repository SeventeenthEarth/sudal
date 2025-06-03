package server

import (
	"fmt"
	"net/http"

	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"
	userConnect "github.com/seventeenthearth/sudal/internal/feature/user/protocol"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
)

// ServiceRegistry manages all service handlers and their initialization
type ServiceRegistry struct {
	// gRPC service handlers
	HealthHandler *healthConnect.HealthManager
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

	// Initialize OpenAPI handler for REST endpoints
	openAPIHandler, err := di.InitializeOpenAPIHandler()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAPI handler: %w", err)
	}

	// Initialize Swagger UI handler
	swaggerHandler := di.InitializeSwaggerHandler()

	return &ServiceRegistry{
		HealthHandler:  healthHandler,
		UserHandler:    userHandler,
		OpenAPIHandler: openAPIHandler,
		SwaggerHandler: swaggerHandler,
	}, nil
}

// RouteRegistrar handles route registration with appropriate middleware chains
type RouteRegistrar struct {
	mux     *http.ServeMux
	chains  *ServiceChains
	registry *ServiceRegistry
}

// NewRouteRegistrar creates a new route registrar
func NewRouteRegistrar(mux *http.ServeMux, chains *ServiceChains, registry *ServiceRegistry) *RouteRegistrar {
	return &RouteRegistrar{
		mux:     mux,
		chains:  chains,
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

// ServiceConfiguration holds the configuration for service chains
type ServiceConfiguration struct {
	// Define which procedures require authentication
	ProtectedProcedures []string
}

// GetDefaultServiceConfiguration returns the default service configuration
func GetDefaultServiceConfiguration() *ServiceConfiguration {
	return &ServiceConfiguration{
		ProtectedProcedures: []string{
			"/user.v1.UserService/GetUserProfile",
			"/user.v1.UserService/UpdateUserProfile",
		},
	}
}
