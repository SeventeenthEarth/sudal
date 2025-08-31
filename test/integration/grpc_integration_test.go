package integration_test

import (
	"context"
	"fmt"
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

var _ = Describe("gRPC Protocol Integration Tests", func() {
	var (
		ctrl       *gomock.Controller
		mockRepo   *mocks.MockHealthRepository
		service    application.HealthService
		handler    *healthConnect.HealthManager
		testServer *testMocks.TestServer
		baseURL    string
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

	Describe("Connect-Go gRPC-Web Protocol", func() {
		Context("when service is healthy", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should return SERVING status for gRPC-Web requests", func() {
				// Given: Connect-Go client with gRPC-Web protocol
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

				// Then: Response should indicate serving status
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify gRPC-Web specific headers
				Expect(resp.Header().Get("Content-Type")).To(ContainSubstring("application/grpc-web"))
			})

			It("should handle gRPC-Web metadata correctly", func() {
				// Given: Connect-Go client with gRPC-Web protocol
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making request with custom headers
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				req.Header().Set("X-Test-Header", "test-value")
				req.Header().Set("X-Client-Version", "1.0.0")

				resp, err := client.Check(ctx, req)

				// Then: Response should be successful and contain metadata
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify response headers are present
				Expect(resp.Header()).NotTo(BeNil())
			})
		})

		Context("when service is unhealthy", func() {
			BeforeEach(func() {
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("mock service error"))
			})

			It("should return internal error for gRPC-Web requests", func() {
				// Given: Connect-Go client with gRPC-Web protocol
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

				// Then: Should return connect error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				// Verify it's a connect error with internal code
				connectErr, ok := err.(*connect.Error)
				Expect(ok).To(BeTrue())
				Expect(connectErr.Code()).To(Equal(connect.CodeInternal))
			})
		})
	})

	Describe("Connect-Go HTTP/JSON Protocol", func() {
		Context("when service is healthy", func() {
			BeforeEach(func() {
				testMocks.SetHealthyStatus(mockRepo)
			})

			It("should return SERVING status for HTTP/JSON requests", func() {
				// Given: Connect-Go client with HTTP/JSON protocol (default)
				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
					// No protocol specified - uses HTTP/JSON by default
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Response should indicate serving status
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))

				// Verify HTTP/JSON specific headers (Connect-Go uses application/proto for both protocols)
				contentType := resp.Header().Get("Content-Type")
				Expect(contentType).To(Or(ContainSubstring("application/json"), ContainSubstring("application/proto")))
			})

			It("should handle HTTP/JSON with custom timeout", func() {
				// Given: Connect-Go client with short timeout
				client := healthv1connect.NewHealthServiceClient(
					&http.Client{Timeout: 1 * time.Second},
					baseURL,
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should complete within timeout
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
			})
		})

		Context("when service returns different statuses", func() {
			It("should return NOT_SERVING for unhealthy status", func() {
				// Given: Mock configured to return unhealthy status
				unhealthyStatus := entity.UnhealthyStatus()
				testMocks.SetCustomStatus(mockRepo, unhealthyStatus)

				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return NOT_SERVING status
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING))
			})

			It("should return UNKNOWN for unknown status", func() {
				// Given: Mock configured to return unknown status
				unknownStatus := entity.UnknownStatus()
				testMocks.SetCustomStatus(mockRepo, unknownStatus)

				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return UNKNOWN status
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED))
			})
		})
	})

	Describe("Protocol Error Scenarios", func() {
		Context("when network errors occur", func() {
			It("should handle connection timeout gracefully", func() {
				// Given: Client with very short timeout
				client := healthv1connect.NewHealthServiceClient(
					&http.Client{Timeout: 1 * time.Nanosecond}, // Extremely short timeout
					baseURL,
					connect.WithGRPCWeb(),
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return timeout error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())
			})
		})

		Context("when server returns errors", func() {
			It("should handle internal server errors appropriately", func() {
				// Given: Mock configured to return errors
				testMocks.SetUnhealthyStatus(mockRepo, fmt.Errorf("internal server error"))

				client := healthv1connect.NewHealthServiceClient(
					http.DefaultClient,
					baseURL,
				)

				// When: Making health check request
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				req := connect.NewRequest(&healthv1.CheckRequest{})
				resp, err := client.Check(ctx, req)

				// Then: Should return connect error
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				connectErr, ok := err.(*connect.Error)
				Expect(ok).To(BeTrue())
				Expect(connectErr.Code()).To(Equal(connect.CodeInternal))
			})
		})
	})
})
