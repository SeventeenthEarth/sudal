package di_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("DatabaseHealthHandler Unit Tests", func() {
	var (
		ctrl          *gomock.Controller
		mockDBManager *mocks.MockPostgresManager
		handler       *di.DatabaseHealthHandler
		ctx           context.Context
		recorder      *httptest.ResponseRecorder
		request       *http.Request
	)

	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)

		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockDBManager = mocks.NewMockPostgresManager(ctrl)
		ctx = context.Background()

		// Create HTTP test recorder and request
		recorder = httptest.NewRecorder()
		var err error
		request, err = http.NewRequestWithContext(ctx, http.MethodGet, "/health/database", nil)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		// Clear environment variables to ensure clean test state
		os.Unsetenv("GO_TEST")     // nolint:errcheck
		os.Unsetenv("GINKGO_TEST") // nolint:errcheck

		// Reset config to ensure clean state
		config.ResetViper()

		// Load a minimal test config to prevent panic in IsTestEnvironment
		testConfig := &config.Config{
			AppEnv:      "production",
			Environment: "production",
		}
		config.SetConfig(testConfig)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
		// Clean up environment variables
		os.Unsetenv("GO_TEST")     // nolint:errcheck
		os.Unsetenv("GINKGO_TEST") // nolint:errcheck
		config.ResetViper()
	})

	ginkgo.Describe("NewDatabaseHealthHandler", func() {
		ginkgo.Context("when creating a new database health handler", func() {
			ginkgo.It("should create handler with database manager", func() {
				// When
				handler = di.NewDatabaseHealthHandler(mockDBManager)

				// Then
				gomega.Expect(handler).ToNot(gomega.BeNil())
			})

			ginkgo.It("should create handler with nil database manager", func() {
				// When
				handler = di.NewDatabaseHealthHandler(nil)

				// Then
				gomega.Expect(handler).ToNot(gomega.BeNil())
			})
		})

		ginkgo.Context("when in test environment", func() {
			ginkgo.BeforeEach(func() {
				// Set test environment
				os.Setenv("GO_TEST", "1") // nolint:errcheck
			})

			ginkgo.It("should return mock handler when GO_TEST is set", func() {
				// When
				handler = di.NewDatabaseHealthHandler(mockDBManager)

				// Then
				gomega.Expect(handler).ToNot(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("HandleDatabaseHealth", func() {
		ginkgo.BeforeEach(func() {
			// Create handler manually with proper initialization for testing
			handler = createTestDatabaseHealthHandler(mockDBManager)
		})

		ginkgo.Context("when database manager is available and healthy", func() {
			ginkgo.It("should return healthy status with connection stats", func() {
				// Given
				expectedHealthStatus := &postgres.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats: &postgres.ConnectionStats{
						MaxOpenConnections: 25,
						OpenConnections:    5,
						InUse:              2,
						Idle:               3,
						WaitCount:          0,
						WaitDuration:       0,
						MaxIdleClosed:      0,
						MaxLifetimeClosed:  0,
					},
				}
				mockDBManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When
				handler.HandleDatabaseHealth(recorder, request)

				// Then
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var response di.DatabaseHealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("healthy"))
				gomega.Expect(response.Message).To(gomega.Equal("Database is healthy"))
				gomega.Expect(response.Database).ToNot(gomega.BeNil())
				gomega.Expect(response.Database.Status).To(gomega.Equal("healthy"))
			})
		})

		ginkgo.Context("when database manager health check fails", func() {
			ginkgo.It("should return unhealthy status with error", func() {
				// Given
				expectedError := errors.New("database connection failed")
				mockDBManager.EXPECT().HealthCheck(gomock.Any()).Return(nil, expectedError)

				// When
				handler.HandleDatabaseHealth(recorder, request)

				// Then
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusServiceUnavailable))

				var response di.DatabaseHealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("unhealthy"))
				gomega.Expect(response.Message).To(gomega.Equal("Database health check failed"))
				gomega.Expect(response.Error).To(gomega.Equal("database connection failed"))
			})
		})

		ginkgo.Context("when database manager is nil", func() {
			ginkgo.BeforeEach(func() {
				handler = di.NewDatabaseHealthHandler(nil)
			})

			ginkgo.It("should return error when database manager is not available", func() {
				// When
				handler.HandleDatabaseHealth(recorder, request)

				// Then
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusServiceUnavailable))

				var response di.DatabaseHealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("error"))
				gomega.Expect(response.Message).To(gomega.Equal("Database manager not available"))
				gomega.Expect(response.Error).To(gomega.Equal("Database manager is nil"))
			})
		})

		ginkgo.Context("when in test environment", func() {
			ginkgo.BeforeEach(func() {
				os.Setenv("GO_TEST", "1") // nolint:errcheck
			})

			ginkgo.It("should return mock response", func() {
				// When
				handler.HandleDatabaseHealth(recorder, request)

				// Then
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

				var response di.DatabaseHealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("healthy"))
				gomega.Expect(response.Message).To(gomega.Equal("Mock database is healthy"))
				gomega.Expect(response.Database).ToNot(gomega.BeNil())
				gomega.Expect(response.Database.Status).To(gomega.Equal("healthy"))
			})
		})

		ginkgo.Context("when request context has timeout", func() {
			ginkgo.BeforeEach(func() {
				// Create request with short timeout
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				var err error
				request, err = http.NewRequestWithContext(timeoutCtx, http.MethodGet, "/health/database", nil)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
			})

			ginkgo.It("should handle context timeout gracefully", func() {
				// Given
				mockDBManager.EXPECT().HealthCheck(gomock.Any()).DoAndReturn(func(ctx context.Context) (*postgres.HealthStatus, error) {
					// Simulate slow operation
					time.Sleep(10 * time.Millisecond)
					return nil, context.DeadlineExceeded
				})

				// When
				handler.HandleDatabaseHealth(recorder, request)

				// Then
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusServiceUnavailable))

				var response di.DatabaseHealthResponse
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				gomega.Expect(err).ToNot(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("unhealthy"))
			})
		})
	})

	ginkgo.Describe("IsTestEnvironment", func() {
		ginkgo.Context("when checking test environment detection", func() {
			ginkgo.It("should return false when no test environment variables are set", func() {
				// Given - clean environment (already done in BeforeEach)

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeFalse())
			})

			ginkgo.It("should return true when GO_TEST is set to 1", func() {
				// Given
				os.Setenv("GO_TEST", "1") // nolint:errcheck

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return true when GINKGO_TEST is set to 1", func() {
				// Given
				os.Setenv("GINKGO_TEST", "1") // nolint:errcheck

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return false when GO_TEST is set to other value", func() {
				// Given
				os.Setenv("GO_TEST", "0") // nolint:errcheck

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeFalse())
			})
		})

		ginkgo.Context("when checking config-based test environment detection", func() {
			ginkgo.It("should return true when config AppEnv is test", func() {
				// Given
				testConfig := &config.Config{
					AppEnv: "test",
				}
				config.SetConfig(testConfig)

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return true when config Environment is test", func() {
				// Given
				testConfig := &config.Config{
					Environment: "test",
				}
				config.SetConfig(testConfig)

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeTrue())
			})

			ginkgo.It("should return false when config indicates production", func() {
				// Given
				testConfig := &config.Config{
					AppEnv:      "production",
					Environment: "production",
				}
				config.SetConfig(testConfig)

				// When
				result := di.IsTestEnvironment()

				// Then
				gomega.Expect(result).To(gomega.BeFalse())
			})
		})
	})

	ginkgo.Describe("NewMockDatabaseHealthHandler", func() {
		ginkgo.Context("when creating a mock database health handler", func() {
			ginkgo.It("should create handler with nil database manager", func() {
				// When
				mockHandler := di.NewMockDatabaseHealthHandler()

				// Then
				gomega.Expect(mockHandler).ToNot(gomega.BeNil())
			})
		})
	})
})

// Helper function to create a DatabaseHealthHandler for testing
func createTestDatabaseHealthHandler(dbManager postgres.PostgresManager) *di.DatabaseHealthHandler {
	// Temporarily clear environment variables to ensure we don't get mock handler
	originalGoTest := os.Getenv("GO_TEST")
	originalGinkgoTest := os.Getenv("GINKGO_TEST")

	os.Unsetenv("GO_TEST")     // nolint:errcheck
	os.Unsetenv("GINKGO_TEST") // nolint:errcheck

	// Create handler using the actual constructor
	handler := di.NewDatabaseHealthHandler(dbManager)

	// Restore original environment variables
	if originalGoTest != "" {
		os.Setenv("GO_TEST", originalGoTest) // nolint:errcheck
	}
	if originalGinkgoTest != "" {
		os.Setenv("GINKGO_TEST", originalGinkgoTest) // nolint:errcheck
	}

	return handler
}
