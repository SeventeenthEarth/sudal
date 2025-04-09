package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
)

// Server represents the HTTP server
type Server struct {
	server *http.Server
	port   string
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

// Start initializes and starts the HTTP server
func (s *Server) Start() error {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Initialize handlers using dependency injection
	healthHandler := di.InitializeHealthHandler()

	// Register routes
	healthHandler.RegisterRoutes(mux)

	// Create the HTTP server
	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server listening on port %s", s.port)
		serverErrors <- s.server.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a signal or an error
	select {
	case err := <-serverErrors:
		return fmt.Errorf("error starting server: %w", err)

	case <-shutdown:
		log.Println("Server is shutting down...")

		// Create a deadline for the shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		err := s.server.Shutdown(ctx)
		if err != nil {
			// Force shutdown if graceful shutdown fails
			s.server.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
