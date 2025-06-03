package integration_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/mocks"
	testMocks "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = Describe("gRPC Metadata and Error Handling Integration Tests", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockHealthRepository
		service  application.HealthService
		handler  *healthConnect.HealthManager
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
		handler = healthConnect.NewHealthManager(service)

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

	Describe("gRPC Metadata Handling", func() {
		Context("when sending custom headers with gRPC-Web", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should handle standard HTTP headers correctly", func() {
				// Given: gRPC-Web client
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request with standard headers
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				req.Header().Set("User-Agent", "integration-test-client/1.0")
				req.Header().Set("Accept-Encoding", "gzip, deflate")
				req.Header().Set("X-Request-ID", "test-request-12345")

				resp, err := client.Check(ctx, req)

				// Then: Request should succeed and response should contain headers
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify response headers
				Expect(resp.Header()).NotTo(BeNil())
				Expect(resp.Header().Get("Content-Type")).To(ContainSubstring("grpc-web"))
			})

			It("should handle custom application headers", func() {
				// Given: gRPC-Web client
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request with custom application headers
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				req.Header().Set("X-Client-Version", "2.1.0")
				req.Header().Set("X-Environment", "integration-test")
				req.Header().Set("X-Trace-ID", "trace-abc123")
				req.Header().Set("X-Custom-Auth", "bearer-token-xyz")

				resp, err := client.Check(ctx, req)

				// Then: Request should succeed with custom headers
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
			})

			It("should handle headers with special characters", func() {
				// Given: gRPC-Web client
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request with headers containing special characters
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				req.Header().Set("X-Special-Chars", "test-value-with-dashes_and_underscores")
				req.Header().Set("X-Numeric", "12345")
				req.Header().Set("X-Mixed", "Value123-Test_456")

				resp, err := client.Check(ctx, req)

				// Then: Request should succeed despite special characters
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
			})
		})

		Context("when sending custom headers with HTTP/JSON", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should handle JSON-specific headers correctly", func() {
				// Given: HTTP/JSON client
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					// Default HTTP/JSON protocol
				)

				// When: Making request with JSON-specific headers
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				req.Header().Set("Accept", "application/json")
				req.Header().Set("Content-Type", "application/json")
				req.Header().Set("X-JSON-Client", "true")

				resp, err := client.Check(ctx, req)

				// Then: Request should succeed with JSON headers
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify JSON response headers (Connect-Go may use application/proto)
				contentType := resp.Header().Get("Content-Type")
				Expect(contentType).To(Or(ContainSubstring("json"), ContainSubstring("proto")))
			})
		})
	})

	Describe("Error Handling Scenarios", func() {
		Context("when service returns different types of errors", func() {
			It("should handle repository errors appropriately", func() {
				// Given: Mock configured to return repository error
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("database connection failed"))

				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return internal error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				connectErr, ok := err.(*connect.Error)
				Expect(ok).To(BeTrue())
				Expect(connectErr.Code()).To(Equal(connect.CodeInternal))
				Expect(connectErr.Message()).To(ContainSubstring("database connection failed"))
			})

			It("should handle timeout errors gracefully", func() {
				// Given: Mock that simulates slow response
				mockRepo.EXPECT().GetStatus(gomock.Any()).DoAndReturn(func(ctx context.Context) (*entity.HealthStatus, error) {
					// Simulate slow operation
					select {
					case <-time.After(2 * time.Second):
						return entity.HealthyStatus(), nil
					case <-ctx.Done():
						return nil, ctx.Err()
					}
				}).AnyTimes()

				client := healthv1connect.NewHealthServiceClient(
					&http.Client{Timeout: 500 * time.Millisecond}, // Short timeout
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request with short timeout
				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return timeout error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})

			It("should handle context cancellation appropriately", func() {
				// Given: Mock that waits for context cancellation
				mockRepo.EXPECT().GetStatus(gomock.Any()).DoAndReturn(func(ctx context.Context) (*entity.HealthStatus, error) {
					<-ctx.Done()
					return nil, ctx.Err()
				}).AnyTimes()

				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request and cancelling context
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				cancel() // Cancel immediately

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return cancellation error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})

		Context("when handling different error codes", func() {
			It("should map service errors to appropriate gRPC codes", func() {
				// Test different error scenarios
				errorScenarios := []struct {
					name         string
					mockError    error
					expectedCode connect.Code
				}{
					{
						name:         "generic service error",
						mockError:    fmt.Errorf("generic service error"),
						expectedCode: connect.CodeInternal,
					},
					{
						name:         "database connection error",
						mockError:    fmt.Errorf("failed to connect to database"),
						expectedCode: connect.CodeInternal,
					},
					{
						name:         "timeout error",
						mockError:    context.DeadlineExceeded,
						expectedCode: connect.CodeInternal,
					},
				}

				for _, scenario := range errorScenarios {
					By(fmt.Sprintf("testing %s", scenario.name))

					// Given: Mock configured with specific error
					testMocks.SetUnhealthyStatus(mockRepo, scenario.mockError)

					client := healthv1connect.NewHealthServiceClient(
						http.DefaultClient,
						baseURL,
						connect.WithGRPCWeb(),
					)

					// When: Making health check request
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					req := connect.NewRequest(&healthv1.CheckRequest{})
					resp, err := client.Check(ctx, req)
					cancel()

					// Then: Should return expected error code
					Expect(err).To(HaveOccurred(), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(resp).To(BeNil(), fmt.Sprintf("Scenario: %s", scenario.name))

					connectErr, ok := err.(*connect.Error)
					Expect(ok).To(BeTrue(), fmt.Sprintf("Scenario: %s", scenario.name))
					Expect(connectErr.Code()).To(Equal(scenario.expectedCode), fmt.Sprintf("Scenario: %s", scenario.name))
				}
			})
		})
	})

	Describe("Protocol-Specific Error Handling", func() {
		Context("when comparing error handling across protocols", func() {
			It("should return consistent error types for gRPC-Web and HTTP/JSON", func() {
				// Given: Mock configured to return error
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("service unavailable"))

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

				// When: Making requests with both protocols
				req := connect.NewRequest(&healthv1.CheckRequest{})

				grpcWebResp, grpcWebErr := grpcWebClient.Check(ctx, req)
				httpResp, httpErr := httpClient.Check(ctx, req)

				// Then: Both should return consistent error types
				Expect(grpcWebErr).To(HaveOccurred())
				Expect(httpErr).To(HaveOccurred())

				Expect(grpcWebResp).To(BeNil())
				Expect(httpResp).To(BeNil())

				// Both should be connect errors with same code
				grpcWebConnectErr, ok1 := grpcWebErr.(*connect.Error)
				httpConnectErr, ok2 := httpErr.(*connect.Error)

				Expect(ok1).To(BeTrue())
				Expect(ok2).To(BeTrue())

				Expect(grpcWebConnectErr.Code()).To(Equal(connect.CodeInternal))
				Expect(httpConnectErr.Code()).To(Equal(connect.CodeInternal))
				Expect(grpcWebConnectErr.Code()).To(Equal(httpConnectErr.Code()))
			})
		})
	})
})
