package integration_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testMocks "github.com/seventeenthearth/sudal/test/integration/mocks"
)

var _ = Describe("gRPC Protocol Consistency Integration Tests", func() {
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

	Describe("Protocol Response Consistency", func() {
		Context("when service is healthy", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should return consistent SERVING status across all protocols", func() {
				// Given: Clients for different protocols
				grpcWebClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				httpClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					// Default HTTP/JSON protocol
				)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// When: Making requests with different protocols
				req := connect.NewRequest(&healthv1.CheckRequest{})

				grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, req)
				httpResp, httpErr := httpClient.Check(ctx, req)

				// Then: Both should return SERVING status
				Expect(grpcWebErr).NotTo(HaveOccurred())
				Expect(httpErr).NotTo(HaveOccurred())

				Expect(grpcWebResp).NotTo(BeNil())
				Expect(httpResp).NotTo(BeNil())

				Expect(grpcWebResp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
				Expect(httpResp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify both responses have the same semantic meaning
				Expect(grpcWebResp.Msg.Status).To(Equal(httpResp.Msg.Status))
			})

			It("should return consistent responses for multiple sequential requests", func() {
				// Given: Single client making multiple requests
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making multiple sequential requests
				numRequests := 5
				responses := make([]*connect.Response[healthv1.CheckResponse], numRequests)

				for i := 0; i < numRequests; i++ {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					req := connect.NewRequest(&healthv1.CheckRequest{})
					resp, err := client.Check(ctx, req)
					cancel()

					Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("Request %d failed", i+1))
					responses[i] = resp
				}

				// Then: All responses should be consistent
				expectedStatus := responses[0].Msg.Status
				for i, resp := range responses {
					Expect(resp.Msg.Status).To(Equal(expectedStatus), fmt.Sprintf("Response %d status inconsistent", i+1))
				}
			})
		})

		Context("when service is unhealthy", func() {
			BeforeEach(func() {
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("service error"))
			})

			It("should return consistent errors across all protocols", func() {
				// Given: Clients for different protocols
				grpcWebClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				httpClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// When: Making requests with different protocols
				req := connect.NewRequest(&healthv1.CheckRequest{})

				grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, req)
				httpResp, httpErr := httpClient.Check(ctx, req)

				// Then: Both should return consistent errors
				Expect(grpcWebErr).To(HaveOccurred())
				Expect(httpErr).To(HaveOccurred())

				Expect(grpcWebResp).To(BeNil())
				Expect(httpResp).To(BeNil())

				// Verify both return connect errors with same code
				grpcWebConnectErr, ok1 := grpcWebErr.(*connect.Error)
				httpConnectErr, ok2 := httpErr.(*connect.Error)

				Expect(ok1).To(BeTrue())
				Expect(ok2).To(BeTrue())

				Expect(grpcWebConnectErr.Code()).To(Equal(connect.CodeInternal))
				Expect(httpConnectErr.Code()).To(Equal(connect.CodeInternal))
				Expect(grpcWebConnectErr.Code()).To(Equal(httpConnectErr.Code()))
			})
		})

		Context("when service returns different status values", func() {
			DescribeTable("should map domain status to proto status consistently",
				func(domainStatus string, expectedProtoStatus healthv1.ServingStatus) {
					// Given: Mock configured with specific domain status
					customStatus := domain.NewStatus(domainStatus)
					testMocks.SetCustomStatus(mockRepo, customStatus)

					grpcWebClient := healthv1connect.NewHealthServiceClient(
						http.DefaultClient,
						baseURL,
						connect.WithGRPCWeb(),
					)

					httpClient := healthv1connect.NewHealthServiceClient(
						http.DefaultClient,
						baseURL,
					)

					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()

					// When: Making requests with different protocols
					req := connect.NewRequest(&healthv1.CheckRequest{})

					grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, req)
					httpResp, httpErr := httpClient.Check(ctx, req)

					// Then: Both should return the same mapped status
					Expect(grpcWebErr).NotTo(HaveOccurred())
					Expect(httpErr).NotTo(HaveOccurred())

					Expect(grpcWebResp.Msg.Status).To(Equal(expectedProtoStatus))
					Expect(httpResp.Msg.Status).To(Equal(expectedProtoStatus))
					Expect(grpcWebResp.Msg.Status).To(Equal(httpResp.Msg.Status))
				},
				Entry("healthy status", "healthy", healthv1.ServingStatus_SERVING_STATUS_SERVING),
				Entry("unhealthy status", "unhealthy", healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING),
				Entry("unknown status", "unknown", healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED),
				Entry("custom status", "custom", healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED),
			)
		})
	})

	Describe("Protocol Header Consistency", func() {
		Context("when making requests with custom headers", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should handle custom headers appropriately for each protocol", func() {
				// Given: Clients for different protocols
				grpcWebClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				httpClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				// When: Making requests with custom headers
				grpcWebReq := connect.NewRequest(&healthv1.CheckRequest{})
				grpcWebReq.Header().Set("X-Test-Header", "grpc-web-value")
				grpcWebReq.Header().Set("X-Client-Type", "grpc-web-client")

				httpReq := connect.NewRequest(&healthv1.CheckRequest{})
				httpReq.Header().Set("X-Test-Header", "http-value")
				httpReq.Header().Set("X-Client-Type", "http-client")

				grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, grpcWebReq)
				httpResp, httpErr := httpClient.Check(ctx, httpReq)

				// Then: Both should succeed and handle headers appropriately
				Expect(grpcWebErr).NotTo(HaveOccurred())
				Expect(httpErr).NotTo(HaveOccurred())

				Expect(grpcWebResp).NotTo(BeNil())
				Expect(httpResp).NotTo(BeNil())

				// Verify protocol-specific content types
				Expect(grpcWebResp.Header().Get("Content-Type")).To(ContainSubstring("grpc-web"))
				// Connect-Go may use application/proto for HTTP/JSON protocol
				httpContentType := httpResp.Header().Get("Content-Type")
				Expect(httpContentType).To(Or(ContainSubstring("json"), ContainSubstring("proto")))

				// Both should return the same business logic result
				Expect(grpcWebResp.Msg.Status).To(Equal(httpResp.Msg.Status))
			})
		})
	})

	Describe("Protocol Performance Consistency", func() {
		Context("when measuring response times", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should have comparable performance across protocols", func() {
				// Given: Clients for different protocols
				grpcWebClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				httpClient := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				numRequests := 10
				grpcWebTimes := make([]time.Duration, numRequests)
				httpTimes := make([]time.Duration, numRequests)

				// When: Making multiple requests and measuring time
				for i := 0; i < numRequests; i++ {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					req := connect.NewRequest(&healthv1.CheckRequest{})

					// Measure gRPC-Web request time
					start := time.Now()
					grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, req)
					grpcWebTimes[i] = time.Since(start)

					// Measure HTTP/JSON request time
					start = time.Now()
					httpResp, httpErr := httpClient.Check(ctx, req)
					httpTimes[i] = time.Since(start)

					cancel()

					Expect(grpcWebErr).NotTo(HaveOccurred())
					Expect(httpErr).NotTo(HaveOccurred())
					Expect(grpcWebResp.Msg.Status).To(Equal(httpResp.Msg.Status))
				}

				// Then: Performance should be comparable
				var grpcWebTotal, httpTotal time.Duration
				for i := 0; i < numRequests; i++ {
					grpcWebTotal += grpcWebTimes[i]
					httpTotal += httpTimes[i]

					// Each individual request should complete quickly
					Expect(grpcWebTimes[i]).To(BeNumerically("<", 2*time.Second))
					Expect(httpTimes[i]).To(BeNumerically("<", 2*time.Second))
				}

				grpcWebAvg := grpcWebTotal / time.Duration(numRequests)
				httpAvg := httpTotal / time.Duration(numRequests)

				// Average times should be reasonable and comparable
				Expect(grpcWebAvg).To(BeNumerically("<", 1*time.Second))
				Expect(httpAvg).To(BeNumerically("<", 1*time.Second))

				// Neither protocol should be significantly slower than the other
				// Allow up to 3x difference to account for protocol overhead
				ratio := float64(grpcWebAvg) / float64(httpAvg)
				if ratio > 1 {
					Expect(ratio).To(BeNumerically("<", 3.0), "gRPC-Web should not be more than 3x slower than HTTP")
				} else {
					Expect(1/ratio).To(BeNumerically("<", 3.0), "HTTP should not be more than 3x slower than gRPC-Web")
				}
			})
		})
	})
})
