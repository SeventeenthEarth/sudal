package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	healthApp "github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthData "github.com/seventeenthearth/sudal/internal/feature/health/data"
	healthHandler "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

var _ = ginkgo.Describe("Integration Tests", func() {
	ginkgo.Describe("Health Endpoints", func() {
		var (
			handler *healthHandler.Handler
			repo    *healthData.HealthRepository
			service healthApp.HealthService
		)

		ginkgo.BeforeEach(func() {
			// 실제 서버 대신 핸들러를 직접 테스트
			repo = healthData.NewRepository(nil) // nil for test environment
			service = healthApp.NewService(repo)
			handler = healthHandler.NewHandler(service)
		})

		ginkgo.Context("Ping Endpoint", func() {
			var (
				req      *http.Request
				recorder *httptest.ResponseRecorder
				status   map[string]string
			)

			ginkgo.BeforeEach(func() {
				// 요청 및 응답 레코더 설정
				req = httptest.NewRequest("GET", "/ping", nil)
				recorder = httptest.NewRecorder()

				// 핸들러 호출
				handler.Ping(recorder, req)

				// 응답 파싱
				status = make(map[string]string)
				err := json.NewDecoder(recorder.Body).Decode(&status)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to decode response body")
			})

			ginkgo.It("should return a 200 OK with 'ok' status", func() {
				// 상태 코드 확인
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

				// 응답 내용 확인
				gomega.Expect(status["status"]).To(gomega.Equal("ok"))
			})
		})

		ginkgo.Context("Health Endpoint", func() {
			var (
				req      *http.Request
				recorder *httptest.ResponseRecorder
				status   map[string]string
			)

			ginkgo.BeforeEach(func() {
				// 요청 및 응답 레코더 설정
				req = httptest.NewRequest("GET", "/healthz", nil)
				recorder = httptest.NewRecorder()

				// 핸들러 호출
				handler.Health(recorder, req)

				// 응답 파싱
				status = make(map[string]string)
				err := json.NewDecoder(recorder.Body).Decode(&status)
				gomega.Expect(err).NotTo(gomega.HaveOccurred(), "Failed to decode response body")
			})

			ginkgo.It("should return a 200 OK with 'healthy' status", func() {
				// 상태 코드 확인
				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))

				// 응답 내용 확인
				gomega.Expect(status["status"]).To(gomega.Equal("healthy"))
			})
		})

		ginkgo.Context("JSON Encoding Error Scenarios", func() {
			ginkgo.Describe("Ping Endpoint JSON Encoding Error", func() {
				ginkgo.It("should handle JSON encoding errors gracefully", func() {
					// Given: A request that will cause JSON encoding to fail
					req := httptest.NewRequest("GET", "/ping", nil)
					recorder := httptest.NewRecorder()

					// Create a service that returns a status with problematic data for JSON encoding
					// We'll use a custom response writer that fails on Write
					failingWriter := &FailingResponseWriter{
						ResponseRecorder:  recorder,
						ShouldFailOnWrite: true,
					}

					// When: Calling the ping handler with failing writer
					handler.Ping(failingWriter, req)

					// Then: Should handle the encoding error
					gomega.Expect(failingWriter.WriteCallCount).To(gomega.BeNumerically(">", 0))
				})
			})

			ginkgo.Describe("Health Endpoint JSON Encoding Error", func() {
				ginkgo.It("should handle JSON encoding errors gracefully", func() {
					// Given: A request that will cause JSON encoding to fail
					req := httptest.NewRequest("GET", "/healthz", nil)
					recorder := httptest.NewRecorder()

					// Create a failing response writer
					failingWriter := &FailingResponseWriter{
						ResponseRecorder:  recorder,
						ShouldFailOnWrite: true,
					}

					// When: Calling the health handler with failing writer
					handler.Health(failingWriter, req)

					// Then: Should handle the encoding error
					gomega.Expect(failingWriter.WriteCallCount).To(gomega.BeNumerically(">", 0))
				})
			})

			ginkgo.Describe("DatabaseHealth Endpoint JSON Encoding Error", func() {
				ginkgo.It("should handle JSON encoding errors gracefully for success response", func() {
					// Given: A request that will cause JSON encoding to fail
					req := httptest.NewRequest("GET", "/health/database", nil)
					recorder := httptest.NewRecorder()

					// Create a failing response writer
					failingWriter := &FailingResponseWriter{
						ResponseRecorder:  recorder,
						ShouldFailOnWrite: true,
					}

					// When: Calling the database health handler with failing writer
					handler.DatabaseHealth(failingWriter, req)

					// Then: Should handle the encoding error
					gomega.Expect(failingWriter.WriteCallCount).To(gomega.BeNumerically(">", 0))
				})
			})
		})
	})
})

// FailingResponseWriter is a custom ResponseWriter that can simulate write failures
type FailingResponseWriter struct {
	*httptest.ResponseRecorder
	ShouldFailOnWrite bool
	WriteCallCount    int
}

func (f *FailingResponseWriter) Write(data []byte) (int, error) {
	f.WriteCallCount++
	if f.ShouldFailOnWrite {
		return 0, fmt.Errorf("simulated write failure")
	}
	return f.ResponseRecorder.Write(data)
}
