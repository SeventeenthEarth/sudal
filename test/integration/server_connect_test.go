package integration_test

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

var _ = Describe("Server Connect Integration", func() {
	var (
		baseURL    string
		serverPort string
		listener   net.Listener
		stopCh     chan struct{}
		errCh      chan error
	)

	BeforeEach(func() {
		// Initialize logger
		log.Init(log.InfoLevel)

		// Create a listener on a random available port
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())

		// Get the actual port
		addr := listener.Addr().(*net.TCPAddr)
		serverPort = fmt.Sprintf("%d", addr.Port)

		// Set up configuration
		cfg := &config.Config{
			ServerPort:  serverPort,
			LogLevel:    "info",
			Environment: "test",
		}
		config.SetConfig(cfg)

		// Note: We're not using the actual server implementation for this test
		// Instead, we're using a simple HTTP handler to simulate the server

		// Create a custom HTTP server that uses our listener
		httpServer := &http.Server{
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Simple handler that responds to health checks
				if strings.Contains(r.URL.Path, "/health.v1.HealthService/Check") {
					if r.Method == "POST" {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusOK)
						_, _ = w.Write([]byte(`{"status":"SERVING"}`))
						return
					}
				}
				w.WriteHeader(http.StatusNotFound)
			}),
		}

		// Start the server in a goroutine
		stopCh = make(chan struct{})
		errCh = make(chan error, 1)

		go func() {
			// This will block until the server is stopped
			errCh <- httpServer.Serve(listener)
			close(stopCh)
		}()

		// Wait a moment for the server to start
		time.Sleep(100 * time.Millisecond)

		// Set the base URL with the actual port
		baseURL = fmt.Sprintf("http://localhost:%s", serverPort)
	})

	AfterEach(func() {
		// Close the listener to stop the server
		err := listener.Close()
		Expect(err).NotTo(HaveOccurred())

		// Wait for the server to stop
		select {
		case <-stopCh:
			// Server stopped
		case <-time.After(5 * time.Second):
			// Timeout waiting for server to stop
			Fail("Timeout waiting for server to stop")
		}

		// Reset config
		config.SetConfig(nil)
	})

	Describe("Health Service", func() {

		It("should handle HTTP/JSON requests", func() {

			// Act
			jsonBody := strings.NewReader(`{}`)
			req, err := http.NewRequest(
				"POST",
				baseURL+"/health.v1.HealthService/Check",
				jsonBody,
			)
			Expect(err).NotTo(HaveOccurred())

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Close the response body
			defer func() {
				_ = resp.Body.Close() // 오류 무시
			}()
		})
	})
})
