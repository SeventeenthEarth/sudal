package integration_test

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/interface/connect"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

var _ = Describe("Health Connect Service Integration", func() {
	var (
		server   *http.Server
		client   healthv1connect.HealthServiceClient
		baseURL  string
		listener net.Listener
		mockRepo *mockRepository
	)

	BeforeEach(func() {
		// Initialize logger
		log.Init(log.InfoLevel)

		// Create a mock repository that can be configured for different test scenarios
		mockRepo = &mockRepository{
			status: domain.HealthyStatus(),
			err:    nil,
		}

		// Create a service with the mock repository
		service := application.NewService(mockRepo)

		// Create the Connect handler
		healthHandler := healthConnect.NewHealthServiceHandler(service)
		path, handler := healthv1connect.NewHealthServiceHandler(healthHandler)

		// Create a router and register the Connect handler
		mux := http.NewServeMux()
		mux.Handle(path, handler)

		// Start a test server on a random port
		var err error
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		Expect(err).NotTo(HaveOccurred())

		// Get the server address
		addr := listener.Addr().String()
		baseURL = "http://" + addr

		// Create the HTTP server
		server = &http.Server{
			Handler: mux,
		}

		// Start the server in a goroutine
		go func() {
			_ = server.Serve(listener)
		}()

		// Create a Connect client
		client = healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			baseURL,
		)
	})

	AfterEach(func() {
		// Shutdown the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
		_ = listener.Close()
	})

	Describe("Check", func() {
		Context("when the service returns a healthy status", func() {
			BeforeEach(func() {
				mockRepo.status = domain.HealthyStatus()
				mockRepo.err = nil
			})

			It("should return a SERVING status", func() {
				// Act
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_SERVING))
			})
		})

		Context("when the service returns an unhealthy status", func() {
			BeforeEach(func() {
				mockRepo.status = domain.NewStatus("unhealthy")
				mockRepo.err = nil
			})

			It("should return a NOT_SERVING status", func() {
				// Act
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING))
			})
		})

		Context("when the service returns an unknown status", func() {
			BeforeEach(func() {
				mockRepo.status = domain.NewStatus("unknown_status")
				mockRepo.err = nil
			})

			It("should return an UNKNOWN status", func() {
				// Act
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))

				// Assert
				Expect(err).NotTo(HaveOccurred())
				Expect(resp).NotTo(BeNil())
				Expect(resp.Msg).NotTo(BeNil())
				Expect(resp.Msg.Status).To(Equal(healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED))
			})
		})

		Context("when the service returns an error", func() {
			BeforeEach(func() {
				mockRepo.status = nil
				mockRepo.err = errors.New("service error")
			})

			It("should return a connect error with internal code", func() {
				// Act
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				resp, err := client.Check(ctx, connect.NewRequest(&healthv1.CheckRequest{}))

				// Assert
				Expect(err).To(HaveOccurred())
				Expect(resp).To(BeNil())

				// Check that it's a connect error with the correct code
				connectErr, ok := err.(*connect.Error)
				Expect(ok).To(BeTrue())
				Expect(connectErr.Code()).To(Equal(connect.CodeInternal))
			})
		})
	})

	Describe("HTTP/JSON API", func() {
		It("should handle HTTP/JSON requests", func() {
			// Act
			jsonBody := strings.NewReader(`{}`)
			req, err := http.NewRequest(
				"POST",
				baseURL+"/health.v1.HealthService/Check",
				jsonBody,
			)
			Expect(err).NotTo(HaveOccurred())

			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)

			// Assert
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Close the response body
			defer func() {
				_ = resp.Body.Close() // 오류 무시
			}()
		})
	})
})

// mockRepository is a mock implementation of the domain.Repository interface
type mockRepository struct {
	status *domain.Status
	err    error
}

// GetStatus implements the domain.Repository interface
func (m *mockRepository) GetStatus(ctx context.Context) (*domain.Status, error) {
	return m.status, m.err
}
