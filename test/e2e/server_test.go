package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

// E2E tests that connect to the server running in Docker
var _ = ginkgo.Describe("Server E2E Tests", func() {
	var serverURL string

	// Setup before running tests
	ginkgo.BeforeEach(func() {
		// Get server port from environment or use default
		serverPort := "8080" // Default port
		if port := os.Getenv("SERVER_PORT"); port != "" {
			serverPort = port
		}

		// Set the server URL
		serverURL = fmt.Sprintf("http://localhost:%s", serverPort)

		// Check if server is accessible
		// We'll retry a few times in case the server is still starting up
		var err error
		var resp *http.Response
		for i := 0; i < 5; i++ {
			resp, err = http.Get(serverURL + "/ping")
			if err == nil {
				_ = resp.Body.Close() // Ignoring close error during connection check
				break
			}
			// Wait before retrying
			time.Sleep(1 * time.Second)
		}

		// If we still can't connect, fail the test
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to connect to server at "+serverURL+". Make sure the server is running with 'make run' in a separate terminal.")
	})

	// Health check endpoint test
	ginkgo.It("should respond to health check", func() {
		// Make request to health endpoint
		resp, err := http.Get(serverURL + "/healthz")
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to connect to health endpoint")
		defer func() { _ = resp.Body.Close() }() // Properly handle close error

		// Check status code
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK), "Health endpoint returned non-200 status code")

		// Parse and check response
		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to decode JSON response")
		gomega.Expect(result["status"]).To(gomega.Equal("healthy"), "Health endpoint did not return 'healthy' status")
	})

	// Ping endpoint test
	ginkgo.It("should respond to ping", func() {
		// Make request to ping endpoint
		resp, err := http.Get(serverURL + "/ping")
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to connect to ping endpoint")
		defer func() { _ = resp.Body.Close() }() // Properly handle close error

		// Check status code
		gomega.Expect(resp.StatusCode).To(gomega.Equal(http.StatusOK), "Ping endpoint returned non-200 status code")

		// Parse and check response
		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to decode JSON response")
		gomega.Expect(result["status"]).To(gomega.Equal("ok"), "Ping endpoint did not return 'ok' status")
	})

	// Add more endpoint tests as needed
})
