package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	healthApp "github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthData "github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	healthHandler "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	internalMocks "github.com/seventeenthearth/sudal/internal/mocks"

	testMocks "github.com/seventeenthearth/sudal/test/integration/helpers"
)

var _ = ginkgo.Describe("Health Endpoints", func() {
	var (
		repo     *healthData.HealthRepository
		service  healthApp.HealthService
		handler  *healthHandler.Handler
		recorder *httptest.ResponseRecorder
	)

	ginkgo.BeforeEach(func() {
		// Create a new health repository
		repo = healthData.NewRepository(nil) // nil for test environment

		// Create a new health service
		service = healthApp.NewService(repo)

		// Create a new health handler
		handler = healthHandler.NewHandler(service)

		// Create a new recorder to capture the response
		recorder = httptest.NewRecorder()
	})

	ginkgo.Describe("RegisterRoutes", func() {
		ginkgo.It("should register routes to the mux", func() {
			// Create a new ServeMux
			mux := http.NewServeMux()

			// Register routes
			handler.RegisterRoutes(mux)

			// Test ping route
			req := httptest.NewRequest("GET", "/ping", nil)
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)

			// Check the status code
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

			// Parse the response body
			var pingStatus entity.HealthStatus
			err := json.NewDecoder(recorder.Body).Decode(&pingStatus)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(pingStatus.Status).To(gomega.Equal("ok"))

			// Test health route
			req = httptest.NewRequest("GET", "/healthz", nil)
			recorder = httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)

			// Check the status code
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

			// Parse the response body
			var healthStatus entity.HealthStatus
			err = json.NewDecoder(recorder.Body).Decode(&healthStatus)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(healthStatus.Status).To(gomega.Equal("healthy"))
		})
	})

	ginkgo.Describe("Ping Endpoint", func() {
		var req *http.Request

		ginkgo.BeforeEach(func() {
			// Create a new HTTP request
			req = httptest.NewRequest("GET", "/ping", nil)
		})

		ginkgo.JustBeforeEach(func() {
			// Call the ping handler
			handler.Ping(recorder, req)
		})

		ginkgo.It("should return a 200 OK with 'ok' status", func() {
			// Check the status code
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

			// Check the content type
			gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

			// Parse the response body
			var status entity.HealthStatus
			err := json.NewDecoder(recorder.Body).Decode(&status)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check the status
			gomega.Expect(status.Status).To(gomega.Equal("ok"))
		})

		ginkgo.Context("when JSON encoding fails", func() {
			ginkgo.It("should handle encoding errors", func() {
				// Create a failing response writer
				frw := testMocks.NewFailingResponseWriter()

				// Call the ping handler with the failing response writer
				handler.Ping(frw, req)

				// Check that the status code is 500
				gomega.Expect(frw.Code).To(gomega.Equal(http.StatusInternalServerError))
			})
		})
	})

	ginkgo.Describe("Health Endpoint", func() {
		var req *http.Request

		ginkgo.BeforeEach(func() {
			// Create a new HTTP request
			req = httptest.NewRequest("GET", "/healthz", nil)
		})

		ginkgo.JustBeforeEach(func() {
			// Call the health handler
			handler.Health(recorder, req)
		})

		ginkgo.It("should return a 200 OK with 'healthy' status", func() {
			// Check the status code
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

			// Check the content type
			gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

			// Parse the response body
			var status entity.HealthStatus
			err := json.NewDecoder(recorder.Body).Decode(&status)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check the status
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})

		ginkgo.Context("when the service returns an error", func() {
			var (
				mockService  *internalMocks.MockHealthService
				mockHandler  *healthHandler.Handler
				mockRecorder *httptest.ResponseRecorder
				ctrl         *gomock.Controller
			)

			ginkgo.BeforeEach(func() {
				// Create a mock service that returns an error
				ctrl = gomock.NewController(ginkgo.GinkgoT())
				mockService = internalMocks.NewMockHealthService(ctrl)
				mockService.EXPECT().Check(gomock.Any()).Return(nil, fmt.Errorf("service error")).AnyTimes()

				// Create a handler with the mock service
				mockHandler = healthHandler.NewHandler(mockService)

				// Create a new recorder to capture the response
				mockRecorder = httptest.NewRecorder()

				// Call the health handler
				mockHandler.Health(mockRecorder, req)
			})

			ginkgo.AfterEach(func() {
				ctrl.Finish()
			})

			ginkgo.It("should return a 500 Internal Server Error", func() {
				// Check the status code
				gomega.Expect(mockRecorder.Code).To(gomega.Equal(http.StatusInternalServerError))
			})
		})
	})
})
