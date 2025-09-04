package integration_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	healthConnect "github.com/seventeenthearth/sudal/internal/feature/health/protocol"

	"connectrpc.com/connect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/gen/go/health/v1/healthv1connect"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	testhelpers "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = Describe("Health Connect Service Integration", func() {
	var (
		testServer *testhelpers.TestServer
		client     healthv1connect.HealthServiceClient
		baseURL    string
		mockRepo   *mockRepository
	)

	BeforeEach(func() {
		// Initialize logger
		log.Init(log.InfoLevel)

		// Create a mock repository that can be configured for different test scenarios
		mockRepo = &mockRepository{
			status: entity.HealthyStatus(),
			err:    nil,
		}

		// Create a service with the mock repository
		service := application.NewService(mockRepo)

		// Create the Connect handler
		healthHandler := healthConnect.NewHealthManager(service)
		path, handler := healthv1connect.NewHealthServiceHandler(healthHandler)

		// Create a router and register the Connect handler
		mux := http.NewServeMux()
		mux.Handle(path, handler)

		// Start a test server via helper
		var err error
		testServer, err = testhelpers.NewTestServer(mux)
		Expect(err).NotTo(HaveOccurred())
		baseURL = testServer.BaseURL

		// Create a Connect client
		client = healthv1connect.NewHealthServiceClient(
			http.DefaultClient,
			baseURL,
		)
	})

	AfterEach(func() {
		// Shutdown the test server
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		Expect(testServer.Close(ctx)).To(Succeed())
	})

	Describe("Check", func() {
		Context("when the service returns a healthy status", func() {
			BeforeEach(func() {
				mockRepo.status = entity.HealthyStatus()
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
				mockRepo.status = entity.NewHealthStatus("unhealthy")
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
				mockRepo.status = entity.NewHealthStatus("unknown_status")
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

// mockRepository is a mock implementation of the repo.HealthRepository protocol
type mockRepository struct {
	status         *entity.HealthStatus
	databaseStatus *entity.DatabaseStatus
	err            error
}

// GetStatus implements the repo.HealthRepository protocol
func (m *mockRepository) GetStatus(ctx context.Context) (*entity.HealthStatus, error) {
	return m.status, m.err
}

// GetDatabaseStatus implements the repo.HealthRepository protocol
func (m *mockRepository) GetDatabaseStatus(ctx context.Context) (*entity.DatabaseStatus, error) {
	if m.databaseStatus != nil {
		return m.databaseStatus, m.err
	}
	// Return a default healthy database status for tests
	stats := &entity.ConnectionStats{
		MaxOpenConnections: 25,
		OpenConnections:    1,
		InUse:              0,
		Idle:               1,
		WaitCount:          0,
		WaitDuration:       0,
		MaxIdleClosed:      0,
		MaxLifetimeClosed:  0,
	}
	return entity.HealthyDatabaseStatus("Mock database connection is healthy", stats), m.err
}
