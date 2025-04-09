package server_test

import (
	"net"
	"testing"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
)

func TestNewServer(t *testing.T) {
	// Test with empty port
	srv1 := server.NewServer("")
	if srv1 == nil {
		t.Fatal("Expected server to not be nil when port is empty")
	}

	// Test with specific port
	srv2 := server.NewServer("9090")
	if srv2 == nil {
		t.Fatal("Expected server to not be nil when port is specified")
	}
}

func TestServer_Start(t *testing.T) {
	// Test server error handling
	t.Run("ServerError", func(t *testing.T) {
		// Arrange - create a server with a port that's already in use
		// First, start a server on port 9092
		listener, err := net.Listen("tcp", ":9092")
		if err != nil {
			t.Fatalf("Failed to create listener: %v", err)
		}
		defer listener.Close()

		// Now create our test server on the same port
		srv := server.NewServer("9092")

		// Start the server and expect an error
		errCh := make(chan error, 1)
		go func() {
			errCh <- srv.Start()
		}()

		// Wait for the error
		select {
		case err := <-errCh:
			// We expect an error since the port is already in use
			if err == nil {
				t.Error("Expected an error when starting server on an in-use port")
			}
		case <-time.After(2 * time.Second):
			t.Error("Timeout waiting for server error")
		}
	})
}
