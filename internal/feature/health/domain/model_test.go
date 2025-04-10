package domain_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

var _ = ginkgo.Describe("Status", func() {
	ginkgo.Describe("NewStatus", func() {
		ginkgo.It("should create a new status with the given status string", func() {
			// Arrange
			expectedStatus := "test-status"

			// Act
			status := domain.NewStatus(expectedStatus)

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal(expectedStatus))
		})
	})

	ginkgo.Describe("HealthyStatus", func() {
		ginkgo.It("should create a status with 'healthy' status", func() {
			// Act
			status := domain.HealthyStatus()

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})
	})

	ginkgo.Describe("OkStatus", func() {
		ginkgo.It("should create a status with 'ok' status", func() {
			// Act
			status := domain.OkStatus()

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("ok"))
		})
	})
})
