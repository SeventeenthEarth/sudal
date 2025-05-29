package application_test

import (
	"context"
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("HealthCheckUseCase", func() {
	var (
		ctrl     *gomock.Controller
		mockRepo *mocks.MockHealthRepository
		useCase  application.HealthCheckUseCase
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

	ginkgo.Describe("NewHealthCheckUseCase", func() {
		ginkgo.It("should create a new health check use case", func() {
			// Act
			useCase = application.NewHealthCheckUseCase(mockRepo)

			// Assert
			gomega.Expect(useCase).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Execute", func() {
		ginkgo.Context("when the repository returns a status successfully", func() {
			var (
				expectedStatus *entity.HealthStatus
				result         *entity.HealthStatus
				err            error
			)

			ginkgo.BeforeEach(func() {
				expectedStatus = entity.NewHealthStatus("test-healthy")
				mockRepo.EXPECT().GetStatus(gomock.Any()).Return(expectedStatus, nil)
				useCase = application.NewHealthCheckUseCase(mockRepo)
			})

			ginkgo.JustBeforeEach(func() {
				result, err = useCase.Execute(ctx)
			})

			ginkgo.It("should return the status without error", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(result).NotTo(gomega.BeNil())
				gomega.Expect(result.Status).To(gomega.Equal(expectedStatus.Status))
			})
		})

		ginkgo.Context("when the repository returns an error", func() {
			var (
				expectedError error
				result        *entity.HealthStatus
				err           error
			)

			ginkgo.BeforeEach(func() {
				expectedError = errors.New("repository error")
				mockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, expectedError)
				useCase = application.NewHealthCheckUseCase(mockRepo)
			})

			ginkgo.JustBeforeEach(func() {
				result, err = useCase.Execute(ctx)
			})

			ginkgo.It("should return the error and nil status", func() {
				gomega.Expect(err).To(gomega.Equal(expectedError))
				gomega.Expect(result).To(gomega.BeNil())
			})
		})
	})
})
