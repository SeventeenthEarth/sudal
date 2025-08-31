package integration_test

import (
	"context"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	testHelpers "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = Describe("Server Connect Integration", func() {
	var (
		baseURL    string
		testServer *testHelpers.TestServer
	)

	BeforeEach(func() {
		// Initialize logger
		log.Init(log.InfoLevel)

		// Set up configuration (port is implicit via helper BaseURL)
		cfg := &config.Config{LogLevel: "info", AppEnv: "test"}
		config.SetConfig(cfg)

		// Note: We're not using the actual server implementation for this test
		// Instead, we're using a simple HTTP handler to simulate the server

		// Create mux with simple handler and run via helper
		mux := http.NewServeMux()
		mux.HandleFunc("/health.v1.HealthService/Check", func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status":"SERVING"}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		})

		var err error
		testServer, err = testHelpers.NewTestServer(mux)
		Expect(err).NotTo(HaveOccurred())
		baseURL = testServer.BaseURL
	})

	AfterEach(func() {
		// Gracefully shutdown test server
		if testServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = testServer.Close(ctx)
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
