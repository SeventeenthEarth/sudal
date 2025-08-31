package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/protocol"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testMocks "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = Describe("Database Health Integration Tests", func() {
	var (
		ctrl       *gomock.Controller
		mockRepo   *mocks.MockHealthRepository
		service    application.HealthService
		handler    *healthInterface.HealthHandler
		testServer *testMocks.TestServer
		baseURL    string
		httpClient *http.Client
	)

	BeforeEach(func() {
		// Initialize gomock controller
		ctrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockHealthRepository(ctrl)
		httpClient = &http.Client{Timeout: 10 * time.Second}

		// Create service with mock repository
		service = application.NewService(mockRepo)
		handler = healthInterface.NewHealthHandler(service)

		// Setup test server
		mux := http.NewServeMux()
		handler.RegisterRoutes(mux)

		// Start test server via helper
		var err error
		testServer, err = testMocks.NewTestServer(mux)
		Expect(err).NotTo(HaveOccurred())
		baseURL = testServer.BaseURL

		// Note: Mock configuration is done in each test case
	})

	AfterEach(func() {
		if testServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = testServer.Close(ctx)
		}
		if ctrl != nil {
			ctrl.Finish()
		}
	})

	Describe("Database Health Endpoint", func() {
		Context("when database is healthy", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should return 200 status with healthy database information", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return successful response
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				// Parse response
				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Verify response structure
				Expect(healthResponse.Status).To(Equal("healthy"))
				Expect(healthResponse.Message).To(Equal("Database is healthy"))
				Expect(healthResponse.Timestamp).NotTo(BeEmpty())
				Expect(healthResponse.Database).NotTo(BeNil())

				// Verify database status
				Expect(healthResponse.Database.Status).To(Equal("healthy"))
				Expect(healthResponse.Database.Message).To(ContainSubstring("healthy"))
				Expect(healthResponse.Database.Stats).NotTo(BeNil())
			})

			It("should include valid connection statistics", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return connection statistics
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Verify connection statistics
				stats := healthResponse.Database.Stats
				Expect(stats).NotTo(BeNil())
				Expect(stats.MaxOpenConnections).To(BeNumerically(">", 0))
				Expect(stats.OpenConnections).To(BeNumerically(">=", 0))
				Expect(stats.InUse).To(BeNumerically(">=", 0))
				Expect(stats.Idle).To(BeNumerically(">=", 0))
				Expect(stats.WaitCount).To(BeNumerically(">=", 0))
				Expect(stats.WaitDuration).To(BeNumerically(">=", 0))
				Expect(stats.MaxIdleClosed).To(BeNumerically(">=", 0))
				Expect(stats.MaxLifetimeClosed).To(BeNumerically(">=", 0))

				// Verify mathematical consistency
				Expect(stats.OpenConnections).To(Equal(stats.InUse + stats.Idle))
				Expect(stats.OpenConnections).To(BeNumerically("<=", stats.MaxOpenConnections))
			})

			It("should include timestamp in RFC3339 format", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should include valid timestamp
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Verify timestamp format
				Expect(healthResponse.Timestamp).NotTo(BeEmpty())

				// Parse timestamp to verify it's valid RFC3339
				parsedTime, err := time.Parse(time.RFC3339, healthResponse.Timestamp)
				Expect(err).NotTo(HaveOccurred())
				Expect(parsedTime).To(BeTemporally("~", time.Now(), 5*time.Second))
			})

			It("should return application/json content type", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return JSON content type
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				defer resp.Body.Close() // nolint:errcheck
			})
		})

		Context("when database is unhealthy", func() {
			It("should return 503 status with error information", func() {
				// Given: Mock configured with database connection error
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("database connection failed"))
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				// Parse response
				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Verify error response structure
				Expect(healthResponse.Status).To(Equal("error"))
				Expect(healthResponse.Message).To(Equal("Database health check failed"))
				Expect(healthResponse.Timestamp).NotTo(BeEmpty())
				Expect(healthResponse.Database).NotTo(BeNil())
				Expect(healthResponse.Database.Status).To(Equal("unhealthy"))
			})

			It("should include error details in response", func() {
				// Given: Mock configured with specific error
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("connection timeout"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should include error details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("connection timeout"))
			})
		})

		Context("when database is in degraded state", func() {
			BeforeEach(func() {
				testMocks.SetDegradedStatus(mockRepo)
			})

			It("should return degraded status with high connection usage", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return degraded status
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Verify degraded state indicators
				stats := healthResponse.Database.Stats
				Expect(stats).NotTo(BeNil())
				Expect(stats.OpenConnections).To(Equal(stats.MaxOpenConnections)) // All connections in use
				Expect(stats.InUse).To(Equal(stats.MaxOpenConnections))
				Expect(stats.Idle).To(Equal(0))
				Expect(stats.WaitCount).To(BeNumerically(">", 0)) // Requests waiting
				Expect(stats.WaitDuration).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("Connection Pool Statistics Validation", func() {
		Context("when validating connection statistics consistency", func() {
			BeforeEach(func() {
				// Configure mock with specific connection statistics
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    10,
					InUse:              6,
					Idle:               4,
					WaitCount:          2,
					WaitDuration:       100 * time.Millisecond,
					MaxIdleClosed:      5,
					MaxLifetimeClosed:  3,
				}
				dbStatus := &entity.DatabaseStatus{
					Status:  "healthy",
					Message: "Database is healthy",
					Stats:   stats,
				}
				testMocks.SetDatabaseStatus(mockRepo, dbStatus)
			})

			It("should validate mathematical consistency of connection statistics", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return mathematically consistent statistics
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				stats := healthResponse.Database.Stats

				// Validate mathematical relationships
				Expect(stats.OpenConnections).To(Equal(stats.InUse+stats.Idle),
					"OpenConnections should equal InUse + Idle")
				Expect(stats.OpenConnections).To(BeNumerically("<=", stats.MaxOpenConnections),
					"OpenConnections should not exceed MaxOpenConnections")
				Expect(stats.InUse).To(BeNumerically(">=", 0),
					"InUse should be non-negative")
				Expect(stats.Idle).To(BeNumerically(">=", 0),
					"Idle should be non-negative")
				Expect(stats.WaitCount).To(BeNumerically(">=", 0),
					"WaitCount should be non-negative")
				Expect(stats.WaitDuration).To(BeNumerically(">=", 0),
					"WaitDuration should be non-negative")
			})

			It("should include all required connection metrics", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should include all connection metrics
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				stats := healthResponse.Database.Stats

				// Verify all metrics are present
				Expect(stats.MaxOpenConnections).To(Equal(25))
				Expect(stats.OpenConnections).To(Equal(10))
				Expect(stats.InUse).To(Equal(6))
				Expect(stats.Idle).To(Equal(4))
				Expect(stats.WaitCount).To(Equal(int64(2)))
				Expect(stats.WaitDuration).To(Equal(100 * time.Millisecond))
				Expect(stats.MaxIdleClosed).To(Equal(int64(5)))
				Expect(stats.MaxLifetimeClosed).To(Equal(int64(3)))
			})
		})

		Context("when testing different connection pool scenarios", func() {
			DescribeTable("should handle various connection pool states",
				func(maxOpen, open, inUse, idle int, waitCount int64, expectedStatus string) {
					// Given: Mock configured with specific connection pool state
					stats := &entity.ConnectionStats{
						MaxOpenConnections: maxOpen,
						OpenConnections:    open,
						InUse:              inUse,
						Idle:               idle,
						WaitCount:          waitCount,
						WaitDuration:       time.Duration(waitCount) * 10 * time.Millisecond,
						MaxIdleClosed:      0,
						MaxLifetimeClosed:  0,
					}
					dbStatus := &entity.DatabaseStatus{
						Status:  "healthy",
						Message: "Database connection pool status",
						Stats:   stats,
					}
					testMocks.SetDatabaseStatus(mockRepo, dbStatus)

					// When: Making GET request to database health endpoint
					resp, err := httpClient.Get(baseURL + "/health/database")

					// Then: Should return expected status and valid statistics
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					defer resp.Body.Close() // nolint:errcheck

					var healthResponse entity.DetailedHealthStatus
					err = json.NewDecoder(resp.Body).Decode(&healthResponse)
					Expect(err).NotTo(HaveOccurred())

					Expect(healthResponse.Database.Status).To(Equal("healthy"))

					returnedStats := healthResponse.Database.Stats
					Expect(returnedStats.MaxOpenConnections).To(Equal(maxOpen))
					Expect(returnedStats.OpenConnections).To(Equal(open))
					Expect(returnedStats.InUse).To(Equal(inUse))
					Expect(returnedStats.Idle).To(Equal(idle))
					Expect(returnedStats.WaitCount).To(Equal(waitCount))

					// Verify mathematical consistency
					Expect(returnedStats.OpenConnections).To(Equal(returnedStats.InUse + returnedStats.Idle))
				},
				Entry("low usage pool", 25, 5, 2, 3, int64(0), "healthy"),
				Entry("medium usage pool", 25, 15, 10, 5, int64(1), "healthy"),
				Entry("high usage pool", 25, 23, 20, 3, int64(5), "healthy"),
				Entry("fully utilized pool", 25, 25, 25, 0, int64(10), "healthy"),
				Entry("minimal pool", 5, 2, 1, 1, int64(0), "healthy"),
			)
		})
	})
})
