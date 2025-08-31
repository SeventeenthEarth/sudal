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

var _ = Describe("Database Health Error Scenarios Integration Tests", func() {
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

	Describe("Database Connection Error Scenarios", func() {
		Context("when database connection fails", func() {
			It("should handle connection timeout errors", func() {
				// Given: Mock configured to return connection timeout error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("connection timeout: dial tcp 127.0.0.1:5432: i/o timeout"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with timeout error details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Status).To(Equal("error"))
				Expect(healthResponse.Message).To(Equal("Database health check failed"))
				Expect(healthResponse.Database).NotTo(BeNil())
				Expect(healthResponse.Database.Status).To(Equal("unhealthy"))
				Expect(healthResponse.Database.Message).To(ContainSubstring("connection timeout"))
			})

			It("should handle connection refused errors", func() {
				// Given: Mock configured to return connection refused error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("connection refused: dial tcp 127.0.0.1:5432: connect: connection refused"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with connection refused details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("connection refused"))
			})

			It("should handle authentication errors", func() {
				// Given: Mock configured to return authentication error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: password authentication failed for user \"testuser\""))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with authentication error details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("authentication failed"))
			})

			It("should handle database not found errors", func() {
				// Given: Mock configured to return database not found error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: database \"nonexistent_db\" does not exist"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with database not found details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("does not exist"))
			})
		})

		Context("when database pool is exhausted", func() {
			It("should handle connection pool exhaustion", func() {
				// Given: Mock configured to simulate pool exhaustion
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, fmt.Errorf("connection pool exhausted: all connections are in use")).AnyTimes()

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with pool exhaustion details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("connection pool exhausted"))
			})

			It("should handle max connections reached", func() {
				// Given: Mock configured to simulate max connections reached
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, fmt.Errorf("pq: sorry, too many clients already")).AnyTimes()

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with max connections error
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("too many clients"))
			})
		})
	})

	Describe("Context and Timeout Error Scenarios", func() {
		Context("when request context is cancelled", func() {
			It("should handle context cancellation gracefully", func() {
				// Given: Mock that waits for context cancellation
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).DoAndReturn(func(ctx context.Context) (*entity.DatabaseStatus, error) {
					<-ctx.Done()
					return nil, ctx.Err()
				}).AnyTimes()

				// When: Making request with cancelled context
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately

				req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/health/database", nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = httpClient.Do(req)

				// Then: Should handle cancellation appropriately
				Expect(err).To(HaveOccurred()) // HTTP client should return error for cancelled context
			})

			It("should handle request timeout", func() {
				// Given: Mock that simulates slow database response
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).DoAndReturn(func(ctx context.Context) (*entity.DatabaseStatus, error) {
					select {
					case <-time.After(3 * time.Second):
						return &entity.DatabaseStatus{
							Status:  "healthy",
							Message: "Slow response",
							Stats:   &entity.ConnectionStats{},
						}, nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				}).AnyTimes()

				// When: Making request with short timeout
				client := &http.Client{Timeout: 500 * time.Millisecond}
				resp, err := client.Get(baseURL + "/health/database")

				// Then: Should timeout
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})

		Context("when database operations timeout", func() {
			It("should handle database query timeout", func() {
				// Given: Mock configured to return query timeout error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: canceling statement due to statement timeout"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with timeout details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("statement timeout"))
			})

			It("should handle connection timeout during health check", func() {
				// Given: Mock configured to return connection timeout during health check
				testhelpers.SetUnhealthyStatus(mockRepo, context.DeadlineExceeded)

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with timeout error
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("context deadline exceeded"))
			})
		})
	})

	Describe("Database State Error Scenarios", func() {
		Context("when database is in maintenance mode", func() {
			It("should handle database maintenance mode", func() {
				// Given: Mock configured to simulate maintenance mode
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: the database system is starting up"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with maintenance details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("starting up"))
			})

			It("should handle database shutdown", func() {
				// Given: Mock configured to simulate database shutdown
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: the database system is shutting down"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with shutdown details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("shutting down"))
			})
		})

		Context("when database has permission issues", func() {
			It("should handle insufficient privileges", func() {
				// Given: Mock configured to simulate permission error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("pq: permission denied for relation health_check"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return service unavailable with permission details
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))

				defer resp.Body.Close() // nolint:errcheck

				var healthResponse entity.DetailedHealthStatus
				err = json.NewDecoder(resp.Body).Decode(&healthResponse)
				Expect(err).NotTo(HaveOccurred())

				Expect(healthResponse.Database.Message).To(ContainSubstring("permission denied"))
			})
		})
	})

	Describe("Error Response Format Validation", func() {
		Context("when validating error response structure", func() {
			It("should always include required fields in error responses", func() {
				// Given: Mock configured to return various types of errors
				errorScenarios := []struct {
					name  string
					error error
				}{
					{"connection error", fmt.Errorf("connection failed")},
					{"timeout error", context.DeadlineExceeded},
					{"authentication error", fmt.Errorf("authentication failed")},
					{"permission error", fmt.Errorf("permission denied")},
				}

				for _, scenario := range errorScenarios {
					By(fmt.Sprintf("testing %s", scenario.name))

					// Given: Mock configured with specific error
					testhelpers.SetUnhealthyStatus(mockRepo, scenario.error)

					// When: Making GET request to database health endpoint
					resp, err := httpClient.Get(baseURL + "/health/database")

					// Then: Should return properly formatted error response
					Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable), fmt.Sprintf("Scenario: %s", scenario.name))

					defer resp.Body.Close() // nolint:errcheck

					var healthResponse entity.DetailedHealthStatus
					err = json.NewDecoder(resp.Body).Decode(&healthResponse)
					Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Scenario: %s", scenario.name))

					// Verify required fields are present
					Expect(healthResponse.Status).To(Equal("error"), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(healthResponse.Message).To(Equal("Database health check failed"), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(healthResponse.Timestamp).NotTo(BeEmpty(), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(healthResponse.Database).NotTo(BeNil(), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(healthResponse.Database.Status).To(Equal("unhealthy"), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(healthResponse.Database.Message).NotTo(BeEmpty(), fmt.Sprintf("Scenario: %s", scenario.name))

					// Verify timestamp format
					_, timeErr := time.Parse(time.RFC3339, healthResponse.Timestamp)
					Expect(timeErr).NotTo(HaveOccurred(), fmt.Sprintf("Scenario: %s - invalid timestamp format", scenario.name))
				}
			})

			It("should include proper content type for error responses", func() {
				// Given: Mock configured to return error
				testhelpers.SetUnhealthyStatus(mockRepo, fmt.Errorf("database error"))

				// When: Making GET request to database health endpoint
				resp, err := httpClient.Get(baseURL + "/health/database")

				// Then: Should return JSON content type even for errors
				Expect(err).NotTo(HaveOccurred())
				Expect(resp.StatusCode).To(Equal(http.StatusServiceUnavailable))
				Expect(resp.Header.Get("Content-Type")).To(Equal("application/json"))

				defer resp.Body.Close() // nolint:errcheck
			})
		})
	})
})
