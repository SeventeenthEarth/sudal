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
	Listener net.Listener
	BaseURL  string
}

// NewTestServer starts an HTTP server with the provided mux on a random port.
// It waits briefly to ensure the server is accepting connections.
func NewTestServer(mux *http.ServeMux) (*TestServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	srv := &http.Server{Handler: mux}
	ts := &TestServer{
		Server:   srv,
		Listener: listener,
		BaseURL:  "http://" + listener.Addr().String(),
	}

	go func() {
		// Intentionally ignore error: server will return http.ErrServerClosed on Shutdown
		_ = srv.Serve(listener)
	}()

	// Small wait to reduce flakiness in fast tests
	time.Sleep(100 * time.Millisecond)

	return ts, nil
}

// Close gracefully shuts down the server and closes the listener.
func (ts *TestServer) Close(ctx context.Context) error {
	var firstErr error
	if ts.Server != nil {
		if err := ts.Server.Shutdown(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if ts.Listener != nil {
		if err := ts.Listener.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
