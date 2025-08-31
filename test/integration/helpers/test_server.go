package helpers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// TestServer provides a lightweight HTTP server helper for integration tests.
// It binds to a random available port (127.0.0.1:0), exposes BaseURL, and
// ensures graceful shutdown to avoid port leaks and test flakiness.
type TestServer struct {
	Server   *http.Server
	BaseURL  string
	listener net.Listener
}

// NewTestServer starts an HTTP server with the provided mux on a random port.
// It waits briefly to ensure the server is accepting connections.
func NewTestServer(handler http.Handler) (*TestServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	srv := &http.Server{Handler: handler}
	ts := &TestServer{
		Server:   srv,
		BaseURL:  "http://" + listener.Addr().String(),
		listener: listener,
	}

	go func() {
		// Intentionally ignore error: server will return http.ErrServerClosed on Shutdown
		_ = srv.Serve(listener)
	}()

	// Wait for the server to be ready by polling for a connection.
	for i := 0; i < 20; i++ { // Poll for up to 1 second
		conn, err := net.DialTimeout("tcp", listener.Addr().String(), 50*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return ts, nil // Server is ready
		}
		time.Sleep(50 * time.Millisecond)
	}

	// If we can't connect, shutdown and return an error.
	_ = srv.Shutdown(context.Background())
	return nil, fmt.Errorf("server failed to start on %s", listener.Addr().String())
}

// Close gracefully shuts down the server.
func (ts *TestServer) Close(ctx context.Context) error {
	if ts.Server != nil {
		// Shutdown will also close the listener.
		return ts.Server.Shutdown(ctx)
	}
	return nil
}
