package interfaces_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	interfaces "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("Handler", func() {
	var (
		ctrl        *gomock.Controller
		mockService *mocks.MockHealthService
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockService = mocks.NewMockHealthService(ctrl)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("NewHandler", func() {
		ginkgo.It("should create a new handler", func() {
			// Act
			handler := interfaces.NewHandler(mockService)

			// Assert
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Ping", func() {
		var (
			handler  *interfaces.Handler
			req      *http.Request
			recorder *httptest.ResponseRecorder
		)

		ginkgo.JustBeforeEach(func() {
			handler = interfaces.NewHandler(mockService)
			req = httptest.NewRequest("GET", "/ping", nil)
			recorder = httptest.NewRecorder()
			handler.Ping(recorder, req)
		})

		ginkgo.Context("when the service returns a status successfully", func() {
			var expectedStatus *entity.HealthStatus

			ginkgo.BeforeEach(func() {
				expectedStatus = entity.NewHealthStatus("test-ok")
				mockService.EXPECT().Ping(gomock.Any()).Return(expectedStatus, nil)
			})

			ginkgo.It("should return a 200 OK with the correct status", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var status entity.HealthStatus
				err := json.NewDecoder(recorder.Body).Decode(&status)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(status.Status).To(gomega.Equal(expectedStatus.Status))
			})
		})

		ginkgo.Context("when the service returns an error", func() {
			ginkgo.BeforeEach(func() {
				expectedError := fmt.Errorf("service error")
				mockService.EXPECT().Ping(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return a 500 Internal Server Error", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusInternalServerError))
			})
		})
	})

	ginkgo.Describe("Health", func() {
		var (
			handler  *interfaces.Handler
			req      *http.Request
			recorder *httptest.ResponseRecorder
		)

		ginkgo.JustBeforeEach(func() {
			handler = interfaces.NewHandler(mockService)
			req = httptest.NewRequest("GET", "/healthz", nil)
			recorder = httptest.NewRecorder()
			handler.Health(recorder, req)
		})

		ginkgo.Context("when the service returns a status successfully", func() {
			var expectedStatus *entity.HealthStatus

			ginkgo.BeforeEach(func() {
				expectedStatus = entity.NewHealthStatus("test-healthy")
				mockService.EXPECT().Check(gomock.Any()).Return(expectedStatus, nil)
			})

			ginkgo.It("should return a 200 OK with the correct status", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var status entity.HealthStatus
				err := json.NewDecoder(recorder.Body).Decode(&status)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(status.Status).To(gomega.Equal(expectedStatus.Status))
			})
		})

		ginkgo.Context("when the service returns an error", func() {
			ginkgo.BeforeEach(func() {
				expectedError := fmt.Errorf("service error")
				mockService.EXPECT().Check(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return a 500 Internal Server Error", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusInternalServerError))
			})
		})
	})

	ginkgo.Describe("DatabaseHealth", func() {
		var (
			handler  *interfaces.Handler
			req      *http.Request
			recorder *httptest.ResponseRecorder
		)

		ginkgo.JustBeforeEach(func() {
			handler = interfaces.NewHandler(mockService)
			req = httptest.NewRequest("GET", "/health/database", nil)
			recorder = httptest.NewRecorder()
			handler.DatabaseHealth(recorder, req)
		})

		ginkgo.Context("when the service returns a database status successfully", func() {
			var expectedDbStatus *entity.DatabaseStatus

			ginkgo.BeforeEach(func() {
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    1,
					InUse:              0,
					Idle:               1,
				}
				expectedDbStatus = entity.HealthyDatabaseStatus("Database is healthy", stats)
				mockService.EXPECT().CheckDatabase(gomock.Any()).Return(expectedDbStatus, nil)
			})

			ginkgo.It("should return a 200 OK with the correct database status", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var response entity.DetailedHealthStatus
				err := json.NewDecoder(recorder.Body).Decode(&response)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("healthy"))
				gomega.Expect(response.Message).To(gomega.Equal("Database is healthy"))
				gomega.Expect(response.Database).NotTo(gomega.BeNil())
				gomega.Expect(response.Database.Status).To(gomega.Equal(expectedDbStatus.Status))
				gomega.Expect(response.Database.Message).To(gomega.Equal(expectedDbStatus.Message))
				gomega.Expect(response.Timestamp).NotTo(gomega.BeEmpty())
			})
		})

		ginkgo.Context("when the service returns an error", func() {
			ginkgo.BeforeEach(func() {
				expectedError := fmt.Errorf("database service error")
				mockService.EXPECT().CheckDatabase(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return a 503 Service Unavailable", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusServiceUnavailable))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var response entity.DetailedHealthStatus
				err := json.NewDecoder(recorder.Body).Decode(&response)
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(response.Status).To(gomega.Equal("error"))
				gomega.Expect(response.Message).To(gomega.Equal("Database health check failed"))
				gomega.Expect(response.Database).NotTo(gomega.BeNil())
				gomega.Expect(response.Database.Status).To(gomega.Equal("unhealthy"))
			})
		})
	})

	ginkgo.Describe("RegisterRoutes", func() {
		ginkgo.It("should register routes without panicking", func() {
			// Arrange
			handler := interfaces.NewHandler(mockService)
			mux := http.NewServeMux()

			// Act & Assert - This should not panic
			gomega.Expect(func() {
				handler.RegisterRoutes(mux)
			}).NotTo(gomega.Panic())
		})
	})

	ginkgo.Describe("Error handling edge cases", func() {
		var (
			edgeHandler  *interfaces.Handler
			edgeRecorder *httptest.ResponseRecorder
			edgeRequest  *http.Request
		)

		ginkgo.BeforeEach(func() {
			edgeHandler = interfaces.NewHandler(mockService)
			edgeRecorder = httptest.NewRecorder()
			edgeRequest = httptest.NewRequest(http.MethodGet, "/", nil)
		})

		ginkgo.Context("when service returns nil status", func() {
			ginkgo.BeforeEach(func() {
				mockService.EXPECT().Ping(gomock.Any()).Return(nil, nil)
				edgeHandler.Ping(edgeRecorder, edgeRequest)
			})

			ginkgo.It("should handle nil status gracefully", func() {
				gomega.Expect(edgeRecorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.Context("when service returns nil database status", func() {
			ginkgo.BeforeEach(func() {
				mockService.EXPECT().CheckDatabase(gomock.Any()).Return(nil, nil)
				edgeHandler.DatabaseHealth(edgeRecorder, edgeRequest)
			})

			ginkgo.It("should handle nil database status gracefully", func() {
				gomega.Expect(edgeRecorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})

		ginkgo.Context("when service returns nil health status", func() {
			ginkgo.BeforeEach(func() {
				mockService.EXPECT().Check(gomock.Any()).Return(nil, nil)
				edgeHandler.Health(edgeRecorder, edgeRequest)
			})

			ginkgo.It("should handle nil health status gracefully", func() {
				gomega.Expect(edgeRecorder.Code).To(gomega.Equal(http.StatusOK))
			})
		})
	})
})
