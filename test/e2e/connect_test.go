package e2e_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
)

// E2E tests for Connect-Go service running in Docker
var _ = ginkgo.Describe("Connect-Go E2E Tests", func() {
	var (
		serverURL string
		client    healthv1connect.HealthServiceClient
	)

	// Setup before running tests
	ginkgo.BeforeEach(func() {
		// Get server port from environment or use default
		serverPort := "8080" // Default port
		if port := os.Getenv("SERVER_PORT"); port != "" {
			serverPort = port
		}

		// Set the server URL
		serverURL = fmt.Sprintf("http://localhost:%s", serverPort)

		// Create a Connect client
		client = healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			serverURL,
		)

		// Check if server is accessible
		// We'll retry a few times in case the server is still starting up
		var err error
		var resp *http.Response
		for i := 0; i < 5; i++ {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err = client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))
			cancel()

			if err == nil {
				break
			}

			// Try a simple HTTP request as fallback
			resp, err = http.Get(serverURL + "/ping")
			if err == nil {
				_ = resp.Body.Close() // Ignoring close error during connection check
				break
			}

			// Wait before retrying
			time.Sleep(1 * time.Second)
		}

		// If we still can't connect, fail the test
		gomega.Expect(err).NotTo(gomega.HaveOccurred(),
			"Failed to connect to server at "+serverURL+". Make sure the server is running in Docker on port "+serverPort)
	})

	// Test Connect-Go client
	ginkgo.Describe("Health Service", func() {
		ginkgo.Context("using Connect-Go client", func() {
			ginkgo.It("should return SERVING status", func() {
				// Create context with timeout
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// Make request using Connect client
				resp, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))

				// Verify response
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Connect client request failed")
				gomega.Expect(resp).NotTo(gomega.BeNil(), "Response should not be nil")
				gomega.Expect(resp.Msg).NotTo(gomega.BeNil(), "Response message should not be nil")
				gomega.Expect(resp.Msg.Status).To(gomega.Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING),
					"Health service should return SERVING status")
			})
		})

		ginkgo.Context("using HTTP/JSON", func() {
			ginkgo.It("should handle HTTP/JSON requests", func() {
				// Create JSON request body
				jsonBody := strings.NewReader(`{}`)

				// Create HTTP request
				req, err := http.NewRequest(
					"POST",
					serverURL+"/health.v1.HealthService/Check",
					jsonBody,
				)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to create HTTP request")

				// Set content type for JSON
				req.Header.Set("Content-Type", "application/json")

				// Send request
				resp, err := http.DefaultClient.Do(req)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "HTTP request failed")
				gomega.Expect(resp).NotTo(gomega.BeNil(), "Response should not be nil")
				gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK), "HTTP status should be 200 OK")

				// Parse response
				var response struct {
					Status string `json:"status"`
				}

				err = json.NewDecoder(resp.Body).Decode(&response)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to decode JSON response")
				gomega.Expect(response.Status).To(gomega.Equal("SERVING_STATUS_SERVING"), "Health service should return SERVING_STATUS_SERVING status")

				// Close the response body
				defer func() {
					_ = resp.Body.Close() // Ignoring close error
				}()
			})
		})
	})

	// Test error cases
	ginkgo.Describe("Error Handling", func() {
		ginkgo.It("should reject requests with invalid content type", func() {
			// Create request with invalid content type
			jsonBody := strings.NewReader(`{}`)
			req, err := http.NewRequest(
				"POST",
				serverURL+"/health.v1.HealthService/Check",
				jsonBody,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to create HTTP request")

			// Set invalid content type
			req.Header.Set("Content-Type", "text/plain")

			// Send request
			resp, err := http.DefaultClient.Do(req)
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), "HTTP request failed")
			gomega.Expect(resp).NotTo(gomega.BeNil(), "Response should not be nil")

			// Should return 415 Unsupported Media Type
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusUnsupportedMediaType),
				"Server should reject invalid content type with 415 status")

			// Close the response body
			defer func() {
				_ = resp.Body.Close() // Ignoring close error
			}()
		})

		ginkgo.It("should return 404 for non-existent endpoints", func() {
			// Create request to non-existent endpoint
			jsonBody := strings.NewReader(`{}`)
			req, err := http.NewRequest(
				"POST",
				serverURL+"/health.v1.HealthService/NonExistentMethod",
				jsonBody,
			)
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to create HTTP request")

			// Set content type
			req.Header.Set("Content-Type", "application/json")

			// Send request
			resp, err := http.DefaultClient.Do(req)
			gomega.Expect(err).NotTo(gomega.HaveOccurred(), "HTTP request failed")
			gomega.Expect(resp).NotTo(gomega.BeNil(), "Response should not be nil")

			// Should return 404 Not Found
			gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusNotFound),
				"Server should return 404 for non-existent endpoints")

			// Close the response body
			defer func() {
				_ = resp.Body.Close() // Ignoring close error
			}()
		})
	})

	// Performance tests
	ginkgo.Describe("Performance", func() {
		ginkgo.It("should handle multiple concurrent requests", func() {
			// Number of concurrent requests
			numRequests := 10

			// Channel to collect results
			results := make(chan error, numRequests)

			// Make concurrent requests
			for i := 0; i < numRequests; i++ {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					// Make request using Connect client
					_, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))
					results <- err
				}()
			}

			// Collect and verify results
			for i := 0; i < numRequests; i++ {
				err := <-results
				gomega.Expect(err).NotTo(gomega.HaveOccurred(),
					"Concurrent request failed")
			}
		})
	})
})
