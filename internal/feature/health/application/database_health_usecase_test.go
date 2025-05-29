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

var _ = ginkgo.Describe("DatabaseHealthUseCase", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockHealthRepository
		useCase  application.DatabaseHealthUseCase
	)

	ginkgo.BeforeEach(func() {
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockRepo = mocks.NewMockHealthRepository(ctrl)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("NewDatabaseHealthUseCase", func() {
		ginkgo.It("should create a new database health use case", func() {
			// Act
			useCase = application.NewDatabaseHealthUseCase(mockRepo)

			// Assert
			gomega.Expect(useCase).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Execute", func() {
		var (
			ctx              context.Context
			expectedDbStatus *entity.DatabaseStatus
			actualDbStatus   *entity.DatabaseStatus
			err              error
		)

		ginkgo.BeforeEach(func() {
			ctx = context.Background()
			useCase = application.NewDatabaseHealthUseCase(mockRepo)
		})

		ginkgo.JustBeforeEach(func() {
			actualDbStatus, err = useCase.Execute(ctx)
		})

		ginkgo.Context("when the repository returns a successful database status", func() {
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
				gomega.Expect(actualDbStatus).To(gomega.Equal(expectedDbStatus))
				gomega.Expect(actualDbStatus.Status).To(gomega.Equal("healthy"))
				gomega.Expect(actualDbStatus.Message).To(gomega.Equal("Database is healthy"))
				gomega.Expect(actualDbStatus.Stats).NotTo(gomega.BeNil())
				gomega.Expect(actualDbStatus.Stats.MaxOpenConnections).To(gomega.Equal(25))
			})
		})

		ginkgo.Context("when the repository returns an error", func() {
			var expectedError error

			ginkgo.BeforeEach(func() {
				expectedError = errors.New("database connection failed")
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(nil, expectedError)
			})

			ginkgo.It("should return the error", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err).To(gomega.Equal(expectedError))
				gomega.Expect(actualDbStatus).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when the repository returns an unhealthy database status", func() {
			ginkgo.BeforeEach(func() {
				expectedDbStatus = entity.UnhealthyDatabaseStatus("Database connection failed")
				expectedError := errors.New("database connection failed")
				mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(expectedDbStatus, expectedError)
			})

			ginkgo.It("should return the unhealthy status with error", func() {
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(actualDbStatus).To(gomega.Equal(expectedDbStatus))
				gomega.Expect(actualDbStatus.Status).To(gomega.Equal("unhealthy"))
				gomega.Expect(actualDbStatus.Message).To(gomega.Equal("Database connection failed"))
				gomega.Expect(actualDbStatus.Stats).To(gomega.BeNil())
			})
		})
	})
})
