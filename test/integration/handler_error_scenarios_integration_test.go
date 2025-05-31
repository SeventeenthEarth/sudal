package integration_test

import (
	"context"
	"fmt"
	repo2 "github.com/seventeenthearth/sudal/internal/feature/health/data/repo"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/repo"
	healthInterface "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

var _ = Describe("HealthHandler Error Scenarios Integration Tests", func() {
	var (
		handler *healthInterface.HealthHandler
		service application.HealthService
		repo    repo.HealthRepository
	)

	BeforeEach(func() {
		// Create repository and service
		repo = repo2.NewHealthRepository(nil)
		service = application.NewService(repo)
		handler = healthInterface.NewHealthHandler(service)
	})

	Describe("JSON Encoding Error Scenarios", func() {
		Context("when JSON encoding fails", func() {
			It("should handle Ping endpoint JSON encoding errors", func() {
				// Given: A request and a failing response writer
				req := httptest.NewRequest("GET", "/ping", nil)
				failingWriter := &FailingJSONWriter{
					ResponseRecorder: httptest.NewRecorder(),
					ShouldFailWrite:  true,
				}

				// When: Calling the ping handler with failing writer
				handler.Ping(failingWriter, req)

				// Then: Should attempt to write and handle the error
				Expect(failingWriter.WriteAttempted).To(BeTrue())
			})

			It("should handle Health endpoint JSON encoding errors", func() {
				// Given: A request and a failing response writer
				req := httptest.NewRequest("GET", "/healthz", nil)
				failingWriter := &FailingJSONWriter{
					ResponseRecorder: httptest.NewRecorder(),
					ShouldFailWrite:  true,
				}

				// When: Calling the health handler with failing writer
				handler.Health(failingWriter, req)

				// Then: Should attempt to write and handle the error
				Expect(failingWriter.WriteAttempted).To(BeTrue())
			})

			It("should handle DatabaseHealth endpoint JSON encoding errors for success response", func() {
				// Given: A request and a failing response writer
				req := httptest.NewRequest("GET", "/health/database", nil)
				failingWriter := &FailingJSONWriter{
					ResponseRecorder: httptest.NewRecorder(),
					ShouldFailWrite:  true,
				}

				// When: Calling the database health handler with failing writer
				handler.DatabaseHealth(failingWriter, req)

				// Then: Should attempt to write and handle the error
				Expect(failingWriter.WriteAttempted).To(BeTrue())
			})

			It("should handle DatabaseHealth endpoint JSON encoding errors for error response", func() {
				// Given: A service that returns an error and a failing response writer
				errorService := &ErrorService{
					ShouldFailCheckDatabase: true,
					ErrorToReturn:           fmt.Errorf("database connection failed"),
				}
				errorHandler := healthInterface.NewHealthHandler(errorService)

				req := httptest.NewRequest("GET", "/health/database", nil)
				failingWriter := &FailingJSONWriter{
					ResponseRecorder: httptest.NewRecorder(),
					ShouldFailWrite:  true,
				}

				// When: Calling the database health handler with failing writer
				errorHandler.DatabaseHealth(failingWriter, req)

				// Then: Should attempt to write and handle the error
				Expect(failingWriter.WriteAttempted).To(BeTrue())
			})
		})

		Context("when service returns errors", func() {
			It("should handle Ping service errors", func() {
				// Given: A service that returns an error
				errorService := &ErrorService{
					ShouldFailPing: true,
					ErrorToReturn:  fmt.Errorf("ping service error"),
				}
				errorHandler := healthInterface.NewHealthHandler(errorService)

				req := httptest.NewRequest("GET", "/ping", nil)
				recorder := httptest.NewRecorder()

				// When: Calling the ping handler
				errorHandler.Ping(recorder, req)

				// Then: Should return 500 Internal Server Error
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})

			It("should handle Health service errors", func() {
				// Given: A service that returns an error
				errorService := &ErrorService{
					ShouldFailCheck: true,
					ErrorToReturn:   fmt.Errorf("health check service error"),
				}
				errorHandler := healthInterface.NewHealthHandler(errorService)

				req := httptest.NewRequest("GET", "/healthz", nil)
				recorder := httptest.NewRecorder()

				// When: Calling the health handler
				errorHandler.Health(recorder, req)

				// Then: Should return 500 Internal Server Error
				Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
			})
		})
	})
})

// FailingJSONWriter is a custom ResponseWriter that can simulate JSON encoding failures
type FailingJSONWriter struct {
	*httptest.ResponseRecorder
	ShouldFailWrite bool
	WriteAttempted  bool
}

func (f *FailingJSONWriter) Write(data []byte) (int, error) {
	f.WriteAttempted = true
	if f.ShouldFailWrite {
		return 0, fmt.Errorf("simulated JSON write failure")
	}
	return f.ResponseRecorder.Write(data)
}

// ErrorService is a mock service that can return errors for testing
type ErrorService struct {
	ShouldFailPing          bool
	ShouldFailCheck         bool
	ShouldFailCheckDatabase bool
	ErrorToReturn           error
}

func (e *ErrorService) Ping(ctx context.Context) (*entity.HealthStatus, error) {
	if e.ShouldFailPing {
		return nil, e.ErrorToReturn
	}
	return entity.OkStatus(), nil
}

func (e *ErrorService) Check(ctx context.Context) (*entity.HealthStatus, error) {
	if e.ShouldFailCheck {
		return nil, e.ErrorToReturn
	}
	return entity.HealthyStatus(), nil
}

func (e *ErrorService) CheckDatabase(ctx context.Context) (*entity.DatabaseStatus, error) {
	if e.ShouldFailCheckDatabase {
		return nil, e.ErrorToReturn
	}
	stats := &entity.ConnectionStats{
		MaxOpenConnections: 25,
		OpenConnections:    5,
		InUse:              2,
		Idle:               3,
		WaitCount:          0,
		WaitDuration:       0,
		MaxIdleClosed:      0,
		MaxLifetimeClosed:  0,
	}
	return entity.HealthyDatabaseStatus("Database is healthy", stats), nil
}
