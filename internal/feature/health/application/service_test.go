package application_test

import (
	"context"
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("Service", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockHealthRepository
		service  application.Service
		ctx      context.Context
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockHealthRepository(ctrl)
		ctx = context.Background()
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("NewService", func() {
		ginkgo.It("should create a new service", func() {
			// Act
			service = application.NewService(mockRepo)

			// Assert
			gomega.Expect(service).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Ping", func() {
		var (
			result *entity.HealthStatus
			err    error
		)

		ginkgo.BeforeEach(func() {
			// We need to create a service with our mocks, but the NewService function
			// creates its own use cases. For this test, we'll just verify the behavior
			// of the real service with the real PingUseCase.
			service = application.NewService(mockRepo)
		})

		ginkgo.JustBeforeEach(func() {
			result, err = service.Ping(ctx)
		})

		ginkgo.It("should return an 'ok' status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result).NotTo(gomega.BeNil())
			gomega.Expect(result.Status).To(gomega.Equal("ok"))
		})
	})

	ginkgo.Describe("Check", func() {
		var (
			result *entity.HealthStatus
			err    error
		)

		ginkgo.BeforeEach(func() {
			service = application.NewService(mockRepo)
		})

		ginkgo.JustBeforeEach(func() {
			result, err = service.Check(ctx)
		})

		ginkgo.Context("when the repository returns a status successfully", func() {
			var expectedStatus *entity.HealthStatus

			ginkgo.BeforeEach(func() {
				expectedStatus = entity.NewHealthStatus("test-healthy")
				mockRepo.EXPECT().GetStatus(gomock.Any()).Return(expectedStatus, nil)
			})

			ginkgo.It("should return the status without error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.Status).To(gomega.Equal(expectedStatus.Status))
			})
		})

		ginkgo.Context("when the repository returns an error", func() {
			var expectedError error

			ginkgo.BeforeEach(func() {
				expectedError = errors.New("repository error")
				mockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return the error and nil status", func() {
				gomega.Expect(err).To(gomega.Equal(expectedError))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("CheckDatabase", func() {
		var (
			ctx    context.Context
			result *entity.DatabaseStatus
			err    error
		)

		ginkgo.BeforeEach(func() {
			ctx = context.Background()
			service = application.NewService(mockRepo)
		})

		ginkgo.JustBeforeEach(func() {
			result, err = service.CheckDatabase(ctx)
		})

		ginkgo.Context("when the repository returns a database status successfully", func() {
			var expectedDbStatus *entity.DatabaseStatus

			ginkgo.BeforeEach(func() {
				stats := &entity.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    1,
					InUse:              0,
					Idle:               1,
				}
				expectedDbStatus = entity.HealthyDatabaseStatus("Database is healthy", stats)
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(expectedDbStatus, nil)
			})

			ginkgo.It("should return the database status without error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.Status).To(gomega.Equal(expectedDbStatus.Status))
				gomega.Expect(result.Message).To(gomega.Equal(expectedDbStatus.Message))
				gomega.Expect(result.Stats).To(gomega.Equal(expectedDbStatus.Stats))
			})
		})

		ginkgo.Context("when the repository returns an error", func() {
			var expectedError error

			ginkgo.BeforeEach(func() {
				expectedError = errors.New("database repository error")
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return the error and nil database status", func() {
				gomega.Expect(err).To(gomega.Equal(expectedError))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
