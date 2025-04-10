package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	healthApp "github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthData "github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthHandler "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

var _ = ginkgo.Describe("Health Endpoints", func() {
	var (
		repo     *healthData.Repository
		service  healthApp.Service
		handler  *healthHandler.Handler
		recorder *httptest.ResponseRecorder
	)

	ginkgo.BeforeEach(func() {
		// Create a new health repository
		repo = healthData.NewRepository()

		// Create a new health service
		service = healthApp.NewService(repo)

		// Create a new health handler
		handler = healthHandler.NewHandler(service)

		// Create a new recorder to capture the response
		recorder = httptest.NewRecorder()
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
			var status domain.Status
			err := json.NewDecoder(recorder.Body).Decode(&status)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check the status
			gomega.Expect(status.Status).To(gomega.Equal("ok"))
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
			var status domain.Status
			err := json.NewDecoder(recorder.Body).Decode(&status)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Check the status
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})
	})
})
