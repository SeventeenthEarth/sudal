package entity_test

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

var _ = ginkgo.Describe("HealthStatus", func() {
	ginkgo.Describe("NewHealthStatus", func() {
		ginkgo.It("should create a new health status with the given status string", func() {
			// Arrange
			expectedStatus := "test-status"

			// Act
			status := entity.NewHealthStatus(expectedStatus)

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal(expectedStatus))
		})
	})

	ginkgo.Describe("HealthyStatus", func() {
		ginkgo.It("should create a status with 'healthy' status", func() {
			// Act
			status := entity.HealthyStatus()

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})
	})

	ginkgo.Describe("OkStatus", func() {
		ginkgo.It("should create a status with 'ok' status", func() {
			// Act
			status := entity.OkStatus()

			// Assert
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("ok"))
		})
	})
})
