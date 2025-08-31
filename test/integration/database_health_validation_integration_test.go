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
	testhelpers "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = Describe("Database Health Validation Integration Tests", func() {
	var (
		ctrl       *gomock.Controller
		mockRepo   *mocks.MockHealthRepository
		service    application.HealthService
		handler    *healthInterface.HealthHandler
		testServer *testhelpers.TestServer
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
		testServer, err = testhelpers.NewTestServer(mux)
		Expect(err).NotTo(HaveOccurred())
		baseURL = testServer.BaseURL

		// Note: Mock configuration is done in each test case
	})

	AfterEach(func() {
		if testServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			Expect(testServer.Close(ctx)).To(Succeed())
		}
		if ctrl != nil {
			ctrl.Finish()
		}
	})

	Describe("Connection Statistics Validation", func() {
		Context("when validating connection statistics consistency", func() {
			It("should validate that OpenConnections equals InUse plus Idle", func() {
				// Given: Mock configured with specific connection statistics
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    15,
					InUse:              9,
					Idle:               6,
					WaitCount:          0,
					WaitDuration:       0,
					MaxIdleClosed:      0,
					MaxLifetimeClosed:  0,
				}
				dbStatus := &entity.DatabaseStatus{
					Status:  "healthy",
					Message: "Database is healthy",
					Stats:   stats,
				}
				testhelpers.SetDatabaseStatus(mockRepo, dbStatus)

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should validate mathematical consistency
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				returnedStats := healthResponse.Database.Stats
				Expect(returnedStats.OpenConnections).To(Equal(returnedStats.InUse+returnedStats.Idle),
					"OpenConnections should equal InUse + Idle")
				Expect(returnedStats.OpenConnections).To(Equal(15))
				Expect(returnedStats.InUse).To(Equal(9))
				Expect(returnedStats.Idle).To(Equal(6))
			})

			It("should validate that OpenConnections does not exceed MaxOpenConnections", func() {
				// Given: Mock configured with connection statistics at limit
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 20,
					OpenConnections:    20,
					InUse:              15,
					Idle:               5,
					WaitCount:          3,
					WaitDuration:       200 * time.Millisecond,
					MaxIdleClosed:      2,
					MaxLifetimeClosed:  1,
				}
				dbStatus := &entity.DatabaseStatus{
					Status:  "healthy",
					Message: "Database at capacity",
					Stats:   stats,
				}
				testhelpers.SetDatabaseStatus(mockRepo, dbStatus)

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should validate capacity constraints
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				returnedStats := healthResponse.Database.Stats
				Expect(returnedStats.OpenConnections).To(BeNumerically("<=", returnedStats.MaxOpenConnections),
					"OpenConnections should not exceed MaxOpenConnections")
				Expect(returnedStats.OpenConnections).To(Equal(returnedStats.MaxOpenConnections),
					"Should be at maximum capacity")
				Expect(returnedStats.WaitCount).To(BeNumerically(">", 0),
					"Should have waiting requests when at capacity")
			})

			DescribeTable("should validate connection statistics for various scenarios",
				func(maxOpen, open, inUse, idle int, waitCount int64, waitDuration time.Duration, description string) {
					// Given: Mock configured with specific connection pool state
					stats := &entity.ConnectionStats{
						MaxOpenConnections: maxOpen,
						OpenConnections:    open,
						InUse:              inUse,
						Idle:               idle,
						WaitCount:          waitCount,
						WaitDuration:       waitDuration,
						MaxIdleClosed:      0,
						MaxLifetimeClosed:  0,
					}
					dbStatus := &entity.DatabaseStatus{
						Status:  "healthy",
						Message: description,
						Stats:   stats,
					}
					testhelpers.SetDatabaseStatus(mockRepo, dbStatus)

					// When: Making GET request to database health endpoint
					resp, err := httpClient.Get(baseURL + "/health/database")

					// Then: Should validate all constraints
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.StatusCode).To(Equal(http.StatusOK))

					defer resp.Body.Close() // nolint:errcheck

					var healthResponse entity.DetailedHealthStatus
					err = json.NewDecoder(resp.Body).Decode(&healthResponse)
					Expect(err).NotTo(HaveOccurred())

					returnedStats := healthResponse.Database.Stats

					// Validate mathematical consistency
					Expect(returnedStats.OpenConnections).To(Equal(returnedStats.InUse+returnedStats.Idle),
						fmt.Sprintf("%s: OpenConnections should equal InUse + Idle", description))
					Expect(returnedStats.OpenConnections).To(BeNumerically("<=", returnedStats.MaxOpenConnections),
						fmt.Sprintf("%s: OpenConnections should not exceed MaxOpenConnections", description))
					Expect(returnedStats.InUse).To(BeNumerically(">=", 0),
						fmt.Sprintf("%s: InUse should be non-negative", description))
					Expect(returnedStats.Idle).To(BeNumerically(">=", 0),
						fmt.Sprintf("%s: Idle should be non-negative", description))
					Expect(returnedStats.WaitCount).To(BeNumerically(">=", 0),
						fmt.Sprintf("%s: WaitCount should be non-negative", description))
					Expect(returnedStats.WaitDuration).To(BeNumerically(">=", 0),
						fmt.Sprintf("%s: WaitDuration should be non-negative", description))
				},
				Entry("minimal usage", 10, 2, 1, 1, int64(0), 0*time.Millisecond, "Minimal connection usage"),
				Entry("moderate usage", 25, 12, 8, 4, int64(1), 50*time.Millisecond, "Moderate connection usage"),
				Entry("high usage", 50, 45, 40, 5, int64(5), 200*time.Millisecond, "High connection usage"),
				Entry("at capacity", 20, 20, 20, 0, int64(10), 500*time.Millisecond, "At maximum capacity"),
				Entry("single connection", 1, 1, 1, 0, int64(0), 0*time.Millisecond, "Single connection pool"),
			)
		})

		Context("when validating edge cases", func() {
			It("should handle zero connections scenario", func() {
				// Given: Mock configured with zero connections (startup scenario)
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    0,
					InUse:              0,
					Idle:               0,
					WaitCount:          0,
					WaitDuration:       0,
					MaxIdleClosed:      0,
					MaxLifetimeClosed:  0,
				}
				dbStatus := &entity.DatabaseStatus{
					Status:  "healthy",
					Message: "Database starting up",
					Stats:   stats,
				}
				testhelpers.SetDatabaseStatus(mockRepo, dbStatus)

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should handle zero connections gracefully
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				returnedStats := healthResponse.Database.Stats
				Expect(returnedStats.OpenConnections).To(Equal(0))
				Expect(returnedStats.InUse).To(Equal(0))
				Expect(returnedStats.Idle).To(Equal(0))
				Expect(returnedStats.OpenConnections).To(Equal(returnedStats.InUse + returnedStats.Idle))
			})

			It("should handle maximum connections scenario", func() {
				// Given: Mock configured with maximum connections
				maxConns := 100
				stats := &entity.ConnectionStats{
					MaxOpenConnections: maxConns,
					OpenConnections:    maxConns,
					InUse:              maxConns,
					Idle:               0,
					WaitCount:          50,
					WaitDuration:       2 * time.Second,
					MaxIdleClosed:      10,
					MaxLifetimeClosed:  5,
				}
				dbStatus := &entity.DatabaseStatus{
					Status:  "healthy",
					Message: "Database at maximum capacity",
					Stats:   stats,
				}
				testhelpers.SetDatabaseStatus(mockRepo, dbStatus)

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should handle maximum connections scenario
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				returnedStats := healthResponse.Database.Stats
				Expect(returnedStats.OpenConnections).To(Equal(maxConns))
				Expect(returnedStats.InUse).To(Equal(maxConns))
				Expect(returnedStats.Idle).To(Equal(0))
				Expect(returnedStats.WaitCount).To(BeNumerically(">", 0))
				Expect(returnedStats.WaitDuration).To(BeNumerically(">", 0))
			})
		})
	})

	Describe("Response Format Validation", func() {
		Context("when validating response structure", func() {
			BeforeEach(func() {
				testhelpers.SetHealthyStatus(mockRepo)
			})

			It("should include all required fields in successful response", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should include all required fields
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Validate top-level fields
				Expect(healthResponse.Status).NotTo(BeEmpty())
				Expect(healthResponse.Message).NotTo(BeEmpty())
				Expect(healthResponse.Timestamp).NotTo(BeEmpty())
				Expect(healthResponse.Database).NotTo(BeNil())

				// Validate database fields
				Expect(healthResponse.Database.Status).NotTo(BeEmpty())
				Expect(healthResponse.Database.Message).NotTo(BeEmpty())
				Expect(healthResponse.Database.Stats).NotTo(BeNil())

				// Validate connection stats fields
				stats := healthResponse.Database.Stats
				Expect(stats.MaxOpenConnections).To(BeNumerically(">=", 0))
				Expect(stats.OpenConnections).To(BeNumerically(">=", 0))
				Expect(stats.InUse).To(BeNumerically(">=", 0))
				Expect(stats.Idle).To(BeNumerically(">=", 0))
				Expect(stats.WaitCount).To(BeNumerically(">=", 0))
				Expect(stats.WaitDuration).To(BeNumerically(">=", 0))
				Expect(stats.MaxIdleClosed).To(BeNumerically(">=", 0))
				Expect(stats.MaxLifetimeClosed).To(BeNumerically(">=", 0))
			})

			It("should use proper JSON field names", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should use correct JSON field names
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				// Parse as raw JSON to check field names
				var rawResponse map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&rawResponse)
				Expect(err).NotTo(HaveOccurred())

				// Validate top-level field names
				Expect(rawResponse).To(HaveKey("status"))
				Expect(rawResponse).To(HaveKey("message"))
				Expect(rawResponse).To(HaveKey("timestamp"))
				Expect(rawResponse).To(HaveKey("database"))

				// Validate database field names
				database, ok := rawResponse["database"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(database).To(HaveKey("status"))
				Expect(database).To(HaveKey("message"))
				Expect(database).To(HaveKey("stats"))

				// Validate stats field names
				stats, ok := database["stats"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(stats).To(HaveKey("max_open_connections"))
				Expect(stats).To(HaveKey("open_connections"))
				Expect(stats).To(HaveKey("in_use"))
				Expect(stats).To(HaveKey("idle"))
				Expect(stats).To(HaveKey("wait_count"))
				Expect(stats).To(HaveKey("wait_duration"))
				Expect(stats).To(HaveKey("max_idle_closed"))
				Expect(stats).To(HaveKey("max_lifetime_closed"))
			})

			It("should use RFC3339 timestamp format", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should use RFC3339 timestamp format
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				// Validate timestamp format
				timestamp := healthResponse.Timestamp
				Expect(timestamp).NotTo(BeEmpty())

				// Parse timestamp to verify it's valid RFC3339
				parsedTime, err := time.Parse(time.RFC3339, timestamp)
				Expect(err).NotTo(HaveOccurred())

				// Verify timestamp is recent (within last 5 seconds)
				now := time.Now()
				Expect(parsedTime).To(BeTemporally("~", now, 5*time.Second))

				// Verify timestamp is in UTC
				Expect(parsedTime.Location()).To(Equal(time.UTC))
			})
		})

		Context("when validating data types", func() {
			BeforeEach(func() {
				testhelpers.SetHealthyStatus(mockRepo)
			})

			It("should return correct data types for all fields", func() {
				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return correct data types
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusOK))

				defer resp.Body.Close() // nolint:errcheck

				// Parse as raw JSON to check data types
				var rawResponse map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&rawResponse)
				Expect(err).NotTo(HaveOccurred())

				// Validate top-level data types
				Expect(rawResponse["status"]).To(BeAssignableToTypeOf(""))
				Expect(rawResponse["message"]).To(BeAssignableToTypeOf(""))
				Expect(rawResponse["timestamp"]).To(BeAssignableToTypeOf(""))
				Expect(rawResponse["database"]).To(BeAssignableToTypeOf(map[string]interface{}{}))

				// Validate database data types
				database := rawResponse["database"].(map[string]interface{})
				Expect(database["status"]).To(BeAssignableToTypeOf(""))
				Expect(database["message"]).To(BeAssignableToTypeOf(""))
				Expect(database["stats"]).To(BeAssignableToTypeOf(map[string]interface{}{}))

				// Validate stats data types
				stats := database["stats"].(map[string]interface{})
				Expect(stats["max_open_connections"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["open_connections"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["in_use"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["idle"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["wait_count"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["wait_duration"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["max_idle_closed"]).To(BeAssignableToTypeOf(float64(0)))
				Expect(stats["max_lifetime_closed"]).To(BeAssignableToTypeOf(float64(0)))
			})
		})
	})
})
