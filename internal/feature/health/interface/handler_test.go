package interfaces_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	interfaces "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("Handler", func() {
	var (
		ctrl        *gomock.Controller
		mockService *mocks.MockService
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockService = mocks.NewMockService(ctrl)
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
			var expectedStatus *domain.Status

			ginkgo.BeforeEach(func() {
				expectedStatus = domain.NewStatus("test-ok")
				mockService.EXPECT().Ping(gomock.Any()).Return(expectedStatus, nil)
			})

			ginkgo.It("should return a 200 OK with the correct status", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var status domain.Status
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
			var expectedStatus *domain.Status

			ginkgo.BeforeEach(func() {
				expectedStatus = domain.NewStatus("test-healthy")
				mockService.EXPECT().Check(gomock.Any()).Return(expectedStatus, nil)
			})

			ginkgo.It("should return a 200 OK with the correct status", func() {
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Header().Get("Content-Type")).To(gomega.Equal("application/json"))

				var status domain.Status
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
})
