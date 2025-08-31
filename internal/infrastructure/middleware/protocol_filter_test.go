package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/infrastructure/middleware"
)

var _ = ginkgo.Describe("ProtocolFilterMiddleware", func() {
	var (
		handler       http.Handler
		recorder      *httptest.ResponseRecorder
		grpcOnlyPaths []string
	)

	ginkgo.BeforeEach(func() {
		grpcOnlyPaths = []string{
			"/health.v1.HealthService/",
			"/user.v1.UserService/",
		}

		// Create a simple test handler that returns 200 OK
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success")) // nolint:errcheck
		})

		// Wrap with protocol filter middleware
		handler = middleware.ProtocolFilterMiddleware(grpcOnlyPaths)(testHandler)
		recorder = httptest.NewRecorder()
	})

	ginkgo.Describe("Non-gRPC paths", func() {
		ginkgo.It("should allow all requests to non-restricted paths", func() {
			// Test regular REST API path
			req := httptest.NewRequest("GET", "/api/ping", nil)
			req.Header.Set("Content-Type", "application/json")

			handler.ServeHTTP(recorder, req)

			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			gomega.Expect(recorder.Body.String()).To(gomega.Equal("success"))
		})

		ginkgo.It("should allow requests to documentation paths", func() {
			req := httptest.NewRequest("GET", "/docs", nil)
			req.Header.Set("Content-Type", "text/html")

			handler.ServeHTTP(recorder, req)

			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
		})
	})

	ginkgo.Describe("gRPC-only paths", func() {
		ginkgo.Context("when receiving HTTP/JSON requests", func() {
			ginkgo.It("should block requests to health service", func() {
				req := httptest.NewRequest("POST", "/health.v1.HealthService/Check", strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")

				handler.ServeHTTP(recorder, req)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
			})

			ginkgo.It("should block requests to user service", func() {
				req := httptest.NewRequest("POST", "/user.v1.UserService/RegisterUser", strings.NewReader(`{}`))
				req.Header.Set("Content-Type", "application/json")

				handler.ServeHTTP(recorder, req)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
			})
		})

		ginkgo.Context("when receiving gRPC requests", func() {
			ginkgo.It("should allow gRPC requests with application/grpc content type", func() {
				req := httptest.NewRequest("POST", "/health.v1.HealthService/Check", strings.NewReader("grpc-data"))
				req.Header.Set("Content-Type", "application/grpc+proto")
				req.Header.Set("TE", "trailers")

				handler.ServeHTTP(recorder, req)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				gomega.Expect(recorder.Body.String()).To(gomega.Equal("success"))
			})

			ginkgo.It("should allow gRPC-Web requests", func() {
				req := httptest.NewRequest("POST", "/user.v1.UserService/GetUserProfile", strings.NewReader("grpc-web-data"))
				req.Header.Set("Content-Type", "application/grpc-web+proto")

				handler.ServeHTTP(recorder, req)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.It("should allow HTTP/2 requests with gRPC indicators", func() {
				req := httptest.NewRequest("POST", "/health.v1.HealthService/Check", strings.NewReader("data"))
				req.ProtoMajor = 2
				req.ProtoMinor = 0
				req.Header.Set("TE", "trailers")

				handler.ServeHTTP(recorder, req)

				gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
			})

			ginkgo.DescribeTable("should allow Connect requests",
				func(contentType, body string) {
					req := httptest.NewRequest("POST", "/user.v1.UserService/GetUserProfile", strings.NewReader(body))
					req.Header.Set("Content-Type", contentType)

					handler.ServeHTTP(recorder, req)

					gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
				},
				ginkgo.Entry("with application/connect+proto", "application/connect+proto", "connect-proto-data"),
				ginkgo.Entry("with application/connect+json", "application/connect+json", "{\"x\":1}"),
			)
		})
	})

	ginkgo.Describe("GetGRPCOnlyPaths", func() {
		ginkgo.It("should return the correct gRPC-only paths", func() {
			paths := middleware.GetGRPCOnlyPaths()

			gomega.Expect(paths).To(gomega.ContainElement("/health.v1.HealthService/"))
			gomega.Expect(paths).To(gomega.ContainElement("/user.v1.UserService/"))
			gomega.Expect(len(paths)).To(gomega.Equal(2))
		})
	})

	ginkgo.Describe("Edge cases", func() {
		ginkgo.It("should handle requests with no content type", func() {
			req := httptest.NewRequest("POST", "/health.v1.HealthService/Check", nil)
			// No Content-Type header

			handler.ServeHTTP(recorder, req)

			// Should be blocked as it's not clearly a gRPC request
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
		})

		ginkgo.It("should handle GET requests to gRPC paths", func() {
			req := httptest.NewRequest("GET", "/user.v1.UserService/RegisterUser", nil)

			handler.ServeHTTP(recorder, req)

			// Should be blocked as gRPC typically uses POST
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusNotFound))
		})

		ginkgo.It("should handle partial path matches correctly", func() {
			req := httptest.NewRequest("POST", "/health.v1.HealthService", strings.NewReader(`{}`))
			req.Header.Set("Content-Type", "application/json")

			handler.ServeHTTP(recorder, req)

			// Should be allowed as it doesn't match the exact prefix pattern
			gomega.Expect(recorder.Code).To(gomega.Equal(http.StatusOK))
		})
	})
})
