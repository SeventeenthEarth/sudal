package integration_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testMocks "github.com/seventeenthearth/sudal/test/integration/mocks"
)

var _ = Describe("gRPC Concurrency Integration Tests", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockHealthRepository
		service  application.Service
		handler  *healthConnect.HealthServiceHandler
		server   *http.Server
		listener net.Listener
		baseURL  string
	)

	BeforeEach(func() {
		// Initialize gomock controller
		ctrl = gomock.NewController(GinkgoT())
		mockRepo = mocks.NewMockHealthRepository(ctrl)

		// Create service with mock repository
		service = application.NewService(mockRepo)
		handler = healthConnect.NewHealthServiceHandler(service)

		// Setup test server
		mux := http.NewServeMux()
		path, connectHandler := healthv1connect.NewHealthServiceHandler(handler)
		mux.Handle(path, connectHandler)

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

	Describe("Concurrent Connect-Go gRPC-Web Requests", func() {
		Context("when making multiple concurrent requests", func() {
			It("should handle 10 concurrent gRPC-Web requests successfully", func() {
				// Given: Service configured for healthy state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Multiple clients with gRPC-Web protocol
				numRequests := 10
				results := make([]testMocks.ConcurrentTestResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						start := time.Now()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							connect.WithGRPCWeb(),
						)

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						duration := time.Since(start)

						results[index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil && resp.Msg.Status == healthv1.ServingStatus_SERVING_STATUS_SERVING,
							Error:    err,
							Duration: duration,
							Protocol: "grpc-web",
							Metadata: make(map[string]string),
						}

						if resp != nil {
							for key, values := range resp.Header() {
								if len(values) > 0 {
									results[index].Metadata[key] = values[0]
								}
							}
						}
					}(i)
				}

				wg.Wait()

				// Then: All requests should succeed
				successCount := 0
				for i, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Request %d failed", i+1))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("Request %d was not successful", i+1))
					Expect(result.Duration).To(BeNumerically("<", 5*time.Second), fmt.Sprintf("Request %d took too long", i+1))

					if result.Success {
						successCount++
					}
				}

				Expect(successCount).To(Equal(numRequests), "All requests should succeed")
			})

			It("should handle 25 concurrent gRPC-Web requests with consistent performance", func() {
				// Given: Service configured for healthy state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: High number of concurrent clients
				numRequests := 25
				results := make([]testMocks.ConcurrentTestResult, numRequests)
				var wg sync.WaitGroup

				// When: Making many concurrent requests
				start := time.Now()
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						requestStart := time.Now()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							connect.WithGRPCWeb(),
						)

						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						requestDuration := time.Since(requestStart)

						results[index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil,
							Error:    err,
							Duration: requestDuration,
							Protocol: "grpc-web",
							Metadata: make(map[string]string),
						}
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
				Expect(totalDuration).To(BeNumerically("<", 15*time.Second), "All requests should complete within 15 seconds")

				avgRequestTime := totalRequestTime / time.Duration(numRequests)
				Expect(avgRequestTime).To(BeNumerically("<", 2*time.Second), "Average request time should be under 2 seconds")
			})
		})
	})

	Describe("Concurrent Connect-Go HTTP/JSON Requests", func() {
		Context("when making multiple concurrent requests", func() {
			It("should handle 15 concurrent HTTP/JSON requests successfully", func() {
				// Given: Service configured for healthy state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Multiple clients with HTTP/JSON protocol
				numRequests := 15
				results := make([]testMocks.ConcurrentTestResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						start := time.Now()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							// Default protocol is HTTP/JSON
						)

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						duration := time.Since(start)

						results[index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil && resp.Msg.Status == healthv1.ServingStatus_SERVING_STATUS_SERVING,
							Error:    err,
							Duration: duration,
							Protocol: "http",
							Metadata: make(map[string]string),
						}
					}(i)
				}

				wg.Wait()

				// Then: All requests should succeed
				for i, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("HTTP/JSON request %d failed", i+1))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("HTTP/JSON request %d was not successful", i+1))
				}
			})
		})
	})

	Describe("Mixed Protocol Concurrent Requests", func() {
		Context("when making concurrent requests with different protocols", func() {
			It("should handle mixed gRPC-Web and HTTP/JSON requests consistently", func() {
				// Given: Service configured for healthy state
				testMocks.SetHealthyStatus(mockRepo)

				// Given: Mixed protocol clients
				numRequestsPerProtocol := 10
				totalRequests := numRequestsPerProtocol * 2
				results := make([]testMocks.ConcurrentTestResult, totalRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests with different protocols
				for i := 0; i < numRequestsPerProtocol; i++ {
					// gRPC-Web requests
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						start := time.Now()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							connect.WithGRPCWeb(),
						)

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						duration := time.Since(start)

						results[index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil && resp.Msg.Status == healthv1.ServingStatus_SERVING_STATUS_SERVING,
							Error:    err,
							Duration: duration,
							Protocol: "grpc-web",
							Metadata: make(map[string]string),
						}
					}(i)

					// HTTP/JSON requests
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						start := time.Now()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							// Default HTTP/JSON protocol
						)

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						duration := time.Since(start)

						results[numRequestsPerProtocol+index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil && resp.Msg.Status == healthv1.ServingStatus_SERVING_STATUS_SERVING,
							Error:    err,
							Duration: duration,
							Protocol: "http",
							Metadata: make(map[string]string),
						}
					}(i)
				}

				wg.Wait()

				// Then: All requests should succeed regardless of protocol
				grpcWebSuccessCount := 0
				httpSuccessCount := 0

				for i, result := range results {
					Expect(result.Error).NotTo(HaveOccurred(), fmt.Sprintf("Mixed protocol request %d failed", i+1))
					Expect(result.Success).To(BeTrue(), fmt.Sprintf("Mixed protocol request %d was not successful", i+1))

					if result.Protocol == "grpc-web" {
						grpcWebSuccessCount++
					} else if result.Protocol == "http" {
						httpSuccessCount++
					}
				}

				Expect(grpcWebSuccessCount).To(Equal(numRequestsPerProtocol), "All gRPC-Web requests should succeed")
				Expect(httpSuccessCount).To(Equal(numRequestsPerProtocol), "All HTTP/JSON requests should succeed")
			})
		})
	})

	Describe("Concurrent Error Scenarios", func() {
		Context("when service becomes unhealthy during concurrent requests", func() {
			It("should handle service errors consistently across concurrent requests", func() {
				// Given: Service that will fail
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("service unavailable"))

				numRequests := 8
				results := make([]testMocks.ConcurrentTestResult, numRequests)
				var wg sync.WaitGroup

				// When: Making concurrent requests to failing service
				for i := 0; i < numRequests; i++ {
					wg.Add(1)
					go func(index int) {
						defer wg.Done()

						client := healthv1connect.NewHealthServiceClient(
							http.DefaultClient,
							baseURL,
							connect.WithGRPCWeb(),
						)

						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()

						req := connect.NewRequest(&healthv1.CheckRequest{})
						resp, err := client.Check(ctx, req)

						results[index] = testMocks.ConcurrentTestResult{
							Success:  err == nil && resp != nil,
							Error:    err,
							Duration: 0,
							Protocol: "grpc-web",
							Metadata: make(map[string]string),
						}
					}(i)
				}

				wg.Wait()

				// Then: All requests should fail consistently
				for i, result := range results {
					Expect(result.Error).To(HaveOccurred(), fmt.Sprintf("Request %d should have failed", i+1))
					Expect(result.Success).To(BeFalse(), fmt.Sprintf("Request %d should not be successful", i+1))

					// Verify it's a connect error
					connectErr, ok := result.Error.(*connect.Error)
					Expect(ok).To(BeTrue(), fmt.Sprintf("Request %d should return connect error", i+1))
					Expect(connectErr.Code()).To(Equal(connect.CodeInternal), fmt.Sprintf("Request %d should return internal error", i+1))
				}
			})
		})
	})
})
