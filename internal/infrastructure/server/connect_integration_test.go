package server_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	"github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

var _ = ginkgo.Describe("Connect Integration", func() {
	// Initialize logger before all tests to avoid race conditions
	ginkgo.BeforeEach(func() {
		// Initialize the logger with info level
		log.Init(log.InfoLevel)
	})

	ginkgo.Describe("HealthService Connect Handler", func() {
		var (
			recorder *httptest.ResponseRecorder
			handler  http.Handler
		)

		ginkgo.BeforeEach(func() {
			// Ensure we have a valid config for dependency injection
			cfg := &config.Config{
				ServerPort:  "8080",
				LogLevel:    "info",
				Environment: "test",
			}
			config.SetConfig(cfg)

			// Create a test recorder for capturing HTTP responses
			recorder = httptest.NewRecorder()

			// Create a mock health service that always returns healthy
			mockHealthService := &mockHealthService{
				status: domain.HealthyStatus(),
			}

			// Create the Connect handler
			healthHandler := connect.NewHealthServiceHandler(mockHealthService)
			path, connectHandler := healthv1connect.NewHealthServiceHandler(healthHandler)

			// Create a router and register the Connect handler
			mux := http.NewServeMux()
			mux.Handle(path, connectHandler)
			handler = mux
		})

		ginkgo.AfterEach(func() {
			// Reset config after tests
			config.SetConfig(nil)
		})

		ginkgo.Context("when receiving a health check request via HTTP/JSON", func() {
			ginkgo.It("should return a successful response with SERVING status", func() {
				// Arrange
				reqBody := bytes.NewBufferString("{}")
				req := httptest.NewRequest(http.MethodPost, "/health.v1.HealthService/Check", reqBody)
				req.Header.Set("Content-Type", "application/json")

				// Act
				handler.ServeHTTP(recorder, req)

				// Assert
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

				// Parse the response
				var response struct {
					Status string `json:"status"`
				}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())

				// Check the status
				gomega.Expect(response.Status).To(gomega.Equal("SERVING_STATUS_SERVING"))
			})
		})

		ginkgo.Context("when receiving a health check request with invalid content type", func() {
			ginkgo.It("should return a 415 Unsupported Media Type error", func() {
				// Arrange
				reqBody := bytes.NewBufferString("{}")
				req := httptest.NewRequest(http.MethodPost, "/health.v1.HealthService/Check", reqBody)
				req.Header.Set("Content-Type", "text/plain") // Invalid content type

				// Act
				handler.ServeHTTP(recorder, req)

				// Assert
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusUnsupportedMediaType))
			})
		})
	})
})

// Mock implementation of the health service for testing
type mockHealthService struct {
	status *domain.Status
	err    error
}

func (m *mockHealthService) Check(ctx context.Context) (*domain.Status, error) {
	return m.status, m.err
}

func (m *mockHealthService) Ping(ctx context.Context) (*domain.Status, error) {
	return domain.OkStatus(), nil
}

func (m *mockHealthService) CheckDatabase(ctx context.Context) (*domain.DatabaseStatus, error) {
	// Return a default healthy database status for tests
	stats := &domain.ConnectionStats{
		MaxOpenConnections: 25,
		OpenConnections:    1,
		InUse:              0,
		Idle:               1,
		WaitCount:          0,
		WaitDuration:       0,
		MaxIdleClosed:      0,
		MaxLifetimeClosed:  0,
	}
	return domain.HealthyDatabaseStatus("Mock database connection is healthy", stats), m.err
}
