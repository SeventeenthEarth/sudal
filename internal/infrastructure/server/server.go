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

	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/middleware"
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
	healthHandler := di.InitializeHealthHandler()
	// Initialize Connect-go health service handler
	healthConnectHandler := di.InitializeHealthConnectHandler()
	// Initialize database health handler using DI
	dbHealthHandler, err := di.InitializeDatabaseHealthHandler()
	if err != nil {
		return fmt.Errorf("failed to initialize database health handler: %w", err)
	}

	// Register REST routes
	healthHandler.RegisterRoutes(mux)
	// Register database health check route
	mux.HandleFunc("/health/database", dbHealthHandler.HandleDatabaseHealth)

	// Register Connect-go routes
	// This path pattern will handle both gRPC and HTTP/JSON requests
	path, handler := healthv1connect.NewHealthServiceHandler(healthConnectHandler)
	mux.Handle(path, handler)

	// Apply middleware
	httpHandler := middleware.RequestLogger(mux)

	// Create the HTTP server
	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
