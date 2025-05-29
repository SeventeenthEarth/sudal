package application_test

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

var _ = ginkgo.Describe("PingUseCase", func() {
	var (
		useCase application.PingUseCase
		ctx     context.Context
	)

	ginkgo.BeforeEach(func() {
		ctx = context.Background()
	})

	ginkgo.Describe("NewPingUseCase", func() {
		ginkgo.It("should create a new ping use case", func() {
			// Act
			useCase = application.NewPingUseCase()

			// Assert
			gomega.Expect(useCase).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Execute", func() {
		var (
			result *entity.HealthStatus
			err    error
		)

		ginkgo.BeforeEach(func() {
			useCase = application.NewPingUseCase()
		})

		ginkgo.JustBeforeEach(func() {
			result, err = useCase.Execute(ctx)
		})

		ginkgo.It("should return an 'ok' status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(result).NotTo(gomega.BeNil())
			gomega.Expect(result.Status).To(gomega.Equal("ok"))
		})
	})
})
