package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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

// Start initializes and starts the HTTP server using middleware chains
func (s *Server) Start() error {
	// Initialize TokenVerifier and UserService for middleware chains
	tokenVerifier, err := di.InitializeTokenVerifier()
	if err != nil {
		return fmt.Errorf("failed to initialize token verifier: %w", err)
	}

	userService, err := di.InitializeUserService()
	if err != nil {
		return fmt.Errorf("failed to initialize user service: %w", err)
	}

	// Initialize service registry
	serviceRegistry, err := NewServiceRegistry()
	if err != nil {
		return fmt.Errorf("failed to initialize service registry: %w", err)
	}

	// Build middleware chains
	chainBuilder := NewMiddlewareChainBuilder(tokenVerifier, userService, log.GetLogger())
	serviceConfig := GetDefaultServiceConfiguration()
	chains := chainBuilder.BuildServiceChains(serviceConfig.ProtectedProcedures)

	// Setup routes with middleware chains
	mux := http.NewServeMux()
	routeRegistrar := NewRouteRegistrar(mux, chains, serviceRegistry)

	// Register all routes
	if err := routeRegistrar.RegisterRESTRoutes(); err != nil {
		return fmt.Errorf("failed to register REST routes: %w", err)
	}
	routeRegistrar.RegisterGRPCRoutes()

	// Create the final HTTP handler
	httpHandler := mux

	// Setup HTTP/2 server for gRPC support
	finalHandler, server, err := s.setupHTTP2Server(httpHandler)
	if err != nil {
		return fmt.Errorf("failed to setup HTTP/2 server: %w", err)
	}
	s.server = server
	s.server.Handler = finalHandler

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
		// Treat http.ErrServerClosed as a normal, graceful shutdown
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("Server closed gracefully")
			return nil
		}
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

// setupHTTP2Server configures HTTP/2 server for gRPC support
func (s *Server) setupHTTP2Server(handler http.Handler) (http.Handler, *http.Server, error) {
	// Configure HTTP/2 server for gRPC support
	h2s := &http2.Server{
		// Allow HTTP/2 connections without prior knowledge
		// This is important for gRPC clients
		IdleTimeout: 60 * time.Second,
	}

	// Wrap handler with h2c (HTTP/2 Cleartext) to support gRPC over HTTP/2 without TLS
	// This allows both HTTP/1.1 and HTTP/2 clients to connect
	h2cHandler := h2c.NewHandler(handler, h2s)

	// Create the HTTP server with HTTP/2 support
	server := &http.Server{
		Addr:         ":" + s.port,
		ReadTimeout:  0, // Disable read timeout for gRPC streaming
		WriteTimeout: 0, // Disable write timeout for gRPC streaming
		IdleTimeout:  60 * time.Second,
	}

	// Configure HTTP/2 on the server
	if err := http2.ConfigureServer(server, h2s); err != nil {
		return nil, nil, fmt.Errorf("failed to configure HTTP/2 server: %w", err)
	}

	return h2cHandler, server, nil
}
