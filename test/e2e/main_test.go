package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	healthApp "github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthData "github.com/seventeenthearth/sudal/internal/feature/health/data"
	healthHandler "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

var _ = ginkgo.Describe("End-to-End Tests", func() {
	ginkgo.Describe("Health Endpoints", func() {
		var (
			handler *healthHandler.Handler
			repo    *healthData.Repository
			service healthApp.Service
		)

		ginkgo.BeforeEach(func() {
			// 실제 서버 대신 핸들러를 직접 테스트
			repo = healthData.NewRepository()
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
	})
})
