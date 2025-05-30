package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/gen/go/user/v1/userv1connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/middleware"
	"github.com/seventeenthearth/sudal/internal/infrastructure/openapi"
	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	port   string
	// For testing purposes
	shutdownSignal chan os.Signal
	// Mutex to protect concurrent access to shutdownSignal
	mutex sync.Mutex
}

// NewServer creates a new HTTP server
func NewServer(port string) *Server {
	if port == "" {
		port = "8080" // Default port if not specified
	}

	return &Server{
		port: port,
	}
}

// SetHTTPServer allows setting a custom HTTP server for testing
func (s *Server) SetHTTPServer(server *http.Server) {
	s.server = server
}

// TriggerShutdown triggers a shutdown signal for testing
func (s *Server) TriggerShutdown() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.shutdownSignal != nil {
		s.shutdownSignal <- syscall.SIGINT
	}
}

// Start initializes and starts the HTTP server
func (s *Server) Start() error {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Initialize handlers using dependency injection
	// Initialize Connect-go health service handler (gRPC only)
	healthConnectHandler, err := di.InitializeHealthConnectHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize health connect handler: %w", err)
	}
	// Initialize Connect-go user service handler (gRPC only)
	userConnectHandler, err := di.InitializeUserConnectHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize user connect handler: %w", err)
	}
	// Initialize OpenAPI handler for REST endpoints
	openAPIHandler, err := di.InitializeOpenAPIHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize OpenAPI handler: %w", err)
	}
	// Initialize Swagger UI handler
	swaggerHandler := di.InitializeSwaggerHandler()

	// Register OpenAPI routes for REST endpoints (/api/ping, /api/healthz, /api/health/database)
	openAPIServer, err := openapi.NewServer(openAPIHandler)
	if err != nil {
		return fmt.Errorf("failed to create OpenAPI server: %w", err)
	}
	mux.Handle("/api/", openAPIServer)

	// Register Swagger UI routes
	mux.HandleFunc("/docs", swaggerHandler.ServeSwaggerUI)
	mux.HandleFunc("/docs/", swaggerHandler.ServeSwaggerUI)
	mux.HandleFunc("/api/openapi.yaml", swaggerHandler.ServeOpenAPISpec)

	// Register Connect-go routes with gRPC support
	// These are restricted to gRPC-only via middleware
	healthPath, healthConnectHTTPHandler := healthv1connect.NewHealthServiceHandler(healthConnectHandler)
	mux.Handle(healthPath, healthConnectHTTPHandler)

	// Register UserService Connect-go routes
	userPath, userHTTPHandler := userv1connect.NewUserServiceHandler(userConnectHandler)
	mux.Handle(userPath, userHTTPHandler)

	// Apply middleware stack
	// 1. Protocol filtering for specified paths (gRPC-only restriction)
	protocolFilterHandler := middleware.ProtocolFilterMiddleware(middleware.GetGRPCOnlyPaths())(mux)
	// 2. Request logging
	httpHandler := middleware.RequestLogger(protocolFilterHandler)

	// Configure HTTP/2 server for gRPC support
	h2s := &http2.Server{
		// Allow HTTP/2 connections without prior knowledge
		// This is important for gRPC clients
		IdleTimeout: 60 * time.Second,
	}

	// Wrap handler with h2c (HTTP/2 Cleartext) to support gRPC over HTTP/2 without TLS
	// This allows both HTTP/1.1 and HTTP/2 clients to connect
	h2cHandler := h2c.NewHandler(httpHandler, h2s)

	// Create the HTTP server with HTTP/2 support
	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      h2cHandler,
		ReadTimeout:  0, // Disable read timeout for gRPC streaming
		WriteTimeout: 0, // Disable write timeout for gRPC streaming
		IdleTimeout:  60 * time.Second,
	}

	// Configure HTTP/2 on the server
	if err := http2.ConfigureServer(s.server, h2s); err != nil {
		return fmt.Errorf("failed to configure HTTP/2 server: %w", err)
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Info("Server listening", zap.String("port", s.port))
		serverErrors <- s.server.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	s.mutex.Lock()
	if s.shutdownSignal == nil {
		s.shutdownSignal = make(chan os.Signal, 1)
	}
	shutdown := s.shutdownSignal
	s.mutex.Unlock()
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error starting server: %w", err)

	case <-shutdown:
		log.Info("Server is shutting down...")

		// Create a deadline for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		err := s.server.Shutdown(ctx)
		if err != nil {
			// Force shutdown if graceful shutdown fails
			closeErr := s.server.Close()
			if closeErr != nil {
				return fmt.Errorf("could not stop server gracefully: %w, and failed to close: %w", err, closeErr)
			}
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
