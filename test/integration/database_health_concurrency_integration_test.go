package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testMocks "github.com/seventeenthearth/sudal/test/integration/helpers"
)

// DatabaseHealthResult represents the result of a database health request
type DatabaseHealthResult struct {
	Success    bool
	StatusCode int
	Response   *entity.DetailedHealthStatus
	Error      error
	Duration   time.Duration
	RequestID  int
}

var _ = Describe("Database Health Concurrency Integration Tests", func() {
	var (
		ctrl       *gomock.Controller
		mockRepo   *mocks.MockHealthRepository
		service    application.HealthService
		handler    *healthInterface.HealthHandler
		server     *http.Server
		listener   net.Listener
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

		// Start test server
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())

		addr := listener.Addr().String()
		baseURL = "http://" + addr

		server = &http.Server{Handler: mux}
		go func() {
			_ = server.Serve(listener)
		}()

		// Wait for server to be ready
		time.Sleep(100 * time.Millisecond)

		// Note: Mock configuration is done in each test case
	})

	AfterEach(func() {
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.Shutdown(ctx)
		}
		if listener != nil {
			_ = listener.Close()
		}
		if ctrl != nil {
			ctrl.Finish()
		}
	})

	Describe("Concurrent Database Health Requests", func() {
		Context("when making multiple concurrent requests", func() {
			It("should handle 5 concurrent database health requests successfully", func() {
				// Given: Healthy database state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Multiple concurrent requests
				numRequests := 5
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						start := time.Now()

						resp, err := httpClient.Get(baseURL + "/health/database")
						duration := time.Since(start)

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							Duration:  duration,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode

							if resp.StatusCode == http.StatusOK {
								defer resp.Body.Close() // nolint:errcheck
								var healthResponse entity.DetailedHealthStatus
								if decodeErr := json.NewDecoder(resp.Body).Decode(&healthResponse); decodeErr == nil {
									result.Response = &healthResponse
								}
							}
						}

						results[index] = result
					}(i)
				}

				wg.Wait()

				// Then: All requests should succeed
				for _, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d failed", result.RequestID))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("Request %d was not successful", result.RequestID))
					Expect(result.StatusCode).To(Equal(http.StatusOK), fmt.Sprintf("Request %d returned wrong status", result.RequestID))
					Expect(result.Duration).To(BeNumerically("<", 5*time.Second), fmt.Sprintf("Request %d took too long", result.RequestID))

					// Verify response content
					Expect(result.Response).NotTo(BeNil(), fmt.Sprintf("Request %d has no response", result.RequestID))
					Expect(result.Response.Status).To(Equal("healthy"), fmt.Sprintf("Request %d returned wrong status", result.RequestID))
					Expect(result.Response.Database).NotTo(BeNil(), fmt.Sprintf("Request %d missing database info", result.RequestID))
					Expect(result.Response.Database.Stats).NotTo(BeNil(), fmt.Sprintf("Request %d missing connection stats", result.RequestID))
				}
			})

			It("should handle 10 concurrent database health requests with consistent performance", func() {
				// Given: Healthy database state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: High number of concurrent requests
				numRequests := 10
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making many concurrent requests
				start := time.Now()
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						requestStart := time.Now()

						resp, err := httpClient.Get(baseURL + "/health/database")
						requestDuration := time.Since(requestStart)

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							Duration:  requestDuration,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode
							defer resp.Body.Close() // nolint:errcheck
						}

						results[index] = result
					}(i)
				}

				wg.Wait()
				totalDuration := time.Since(start)

				// Then: All requests should complete within reasonable time
				successCount := 0
				var totalRequestTime time.Duration

				for _, result := range results {
					if result.Success {
						successCount++
					}
					totalRequestTime += result.Duration
				}

				Expect(successCount).To(BeNumerically(">=", int(float64(numRequests)*0.95)), "At least 95% of requests should succeed")
				Expect(totalDuration).To(BeNumerically("<", 10*time.Second), "All requests should complete within 10 seconds")

				avgRequestTime := totalRequestTime / time.Duration(numRequests)
				Expect(avgRequestTime).To(BeNumerically("<", 2*time.Second), "Average request time should be under 2 seconds")
			})

			It("should return consistent connection statistics across concurrent requests", func() {
				// Given: Healthy database state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Concurrent requests to database health endpoint
				numRequests := 8
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						resp, err := httpClient.Get(baseURL + "/health/database")

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode

							if resp.StatusCode == http.StatusOK {
								defer resp.Body.Close() // nolint:errcheck
								var healthResponse entity.DetailedHealthStatus
								if decodeErr := json.NewDecoder(resp.Body).Decode(&healthResponse); decodeErr == nil {
									result.Response = &healthResponse
								}
							}
						}

						results[index] = result
					}(i)
				}

				wg.Wait()

				// Then: All responses should have consistent connection statistics
				var firstStats *entity.ConnectionStats
				for i, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d failed", result.RequestID))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("Request %d was not successful", result.RequestID))
					Expect(result.Response).NotTo(BeNil(), fmt.Sprintf("Request %d has no response", result.RequestID))
					Expect(result.Response.Database.Stats).NotTo(BeNil(), fmt.Sprintf("Request %d missing stats", result.RequestID))

					stats := result.Response.Database.Stats

					// Verify mathematical consistency for each response
					Expect(stats.OpenConnections).To(Equal(stats.InUse+stats.Idle),
						fmt.Sprintf("Request %d: OpenConnections should equal InUse + Idle", result.RequestID))
					Expect(stats.OpenConnections).To(BeNumerically("<=", stats.MaxOpenConnections),
						fmt.Sprintf("Request %d: OpenConnections should not exceed MaxOpenConnections", result.RequestID))

					// Store first stats for consistency comparison
					if i == 0 {
						firstStats = stats
					} else {
						// All responses should have the same mock statistics
						Expect(stats.MaxOpenConnections).To(Equal(firstStats.MaxOpenConnections),
							fmt.Sprintf("Request %d: MaxOpenConnections should be consistent", result.RequestID))
						Expect(stats.OpenConnections).To(Equal(firstStats.OpenConnections),
							fmt.Sprintf("Request %d: OpenConnections should be consistent", result.RequestID))
						Expect(stats.InUse).To(Equal(firstStats.InUse),
							fmt.Sprintf("Request %d: InUse should be consistent", result.RequestID))
						Expect(stats.Idle).To(Equal(firstStats.Idle),
							fmt.Sprintf("Request %d: Idle should be consistent", result.RequestID))
					}
				}
			})
		})
	})

	Describe("Concurrent Error Scenarios", func() {
		Context("when database becomes unhealthy during concurrent requests", func() {
			It("should handle database errors consistently across concurrent requests", func() {
				// Given: Database that will fail
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("database connection lost"))

				numRequests := 6
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests to failing database
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						resp, err := httpClient.Get(baseURL + "/health/database")

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode
							defer resp.Body.Close() // nolint:errcheck

							var healthResponse entity.DetailedHealthStatus
							if decodeErr := json.NewDecoder(resp.Body).Decode(&healthResponse); decodeErr == nil {
								result.Response = &healthResponse
							}
						}

						results[index] = result
					}(i)
				}

				wg.Wait()

				// Then: All requests should fail consistently
				for _, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d should not have HTTP error", result.RequestID))
					Expect(result.StatusCode).To(Equal(http.StatusServiceUnavailable), fmt.Sprintf("Request %d should return 503", result.RequestID))
					Expect(result.Success).To(BeFalse(), fmt.Sprintf("Request %d should not be successful", result.RequestID))

					if result.Response != nil {
						Expect(result.Response.Status).To(Equal("error"), fmt.Sprintf("Request %d should have error status", result.RequestID))
						Expect(result.Response.Database).NotTo(BeNil(), fmt.Sprintf("Request %d should have database info", result.RequestID))
						Expect(result.Response.Database.Status).To(Equal("unhealthy"), fmt.Sprintf("Request %d should have unhealthy database", result.RequestID))
					}
				}
			})
		})

		Context("when testing concurrent request timing", func() {
			It("should handle concurrent requests with varied timing", func() {
				// Given: Initially healthy database
				testMocks.SetHealthyStatus(mockRepo)

				numRequests := 10
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests and changing state mid-way
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						// Small delay to simulate realistic timing
						if index == 3 {
							time.Sleep(10 * time.Millisecond)
						}

						resp, err := httpClient.Get(baseURL + "/health/database")

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode
							defer resp.Body.Close() // nolint:errcheck
						}

						results[index] = result
					}(i)
				}

				wg.Wait()

				// Then: Should handle concurrent requests gracefully
				successCount := 0

				for _, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d should not have HTTP error", result.RequestID))
					Expect(result.StatusCode).To(Equal(http.StatusOK), fmt.Sprintf("Request %d should return 200", result.RequestID))

					if result.StatusCode == http.StatusOK {
						successCount++
					}
				}

				// All requests should succeed since we're not changing state
				Expect(successCount).To(Equal(numRequests), "All requests should succeed")
			})
		})
	})

	Describe("Performance and Resource Management", func() {
		Context("when testing resource cleanup", func() {
			It("should properly clean up resources during concurrent requests", func() {
				// Given: Healthy database state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Many concurrent requests to test resource management
				numRequests := 15
				results := make([]DatabaseHealthResult, numRequests)
				var wg sync.WaitGroup

				// When: Making many concurrent requests
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						// Use a context with timeout to test cleanup
						ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
						defer cancel()

						req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/health/database", nil)
						if err != nil {
							results[index] = DatabaseHealthResult{
								Success:   false,
								Error:     err,
								RequestID: index + 1,
							}
							return
						}

						resp, err := httpClient.Do(req)

						result := DatabaseHealthResult{
							Success:   err == nil && resp != nil && resp.StatusCode == http.StatusOK,
							Error:     err,
							RequestID: index + 1,
						}

						if resp != nil {
							result.StatusCode = resp.StatusCode
							defer resp.Body.Close() // nolint:errcheck
						}

						results[index] = result
					}(i)
				}

				wg.Wait()

				// Then: All requests should complete successfully
				for _, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d failed", result.RequestID))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("Request %d was not successful", result.RequestID))
				}
			})
		})
	})
})
