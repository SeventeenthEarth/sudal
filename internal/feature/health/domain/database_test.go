package domain_test

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

var _ = ginkgo.Describe("Database Models", func() {
	ginkgo.Describe("ConnectionStats", func() {
		ginkgo.It("should create connection stats with correct values", func() {
			// Arrange
			stats := &domain.ConnectionStats{
				MaxOpenConnections: 25,
				OpenConnections:    5,
				InUse:              2,
				Idle:               3,
				WaitCount:          10,
				WaitDuration:       time.Second,
				MaxIdleClosed:      1,
				MaxLifetimeClosed:  2,
			}

			// Assert
			gomega.Expect(stats.MaxOpenConnections).To(gomega.Equal(25))
			gomega.Expect(stats.OpenConnections).To(gomega.Equal(5))
			gomega.Expect(stats.InUse).To(gomega.Equal(2))
			gomega.Expect(stats.Idle).To(gomega.Equal(3))
			gomega.Expect(stats.WaitCount).To(gomega.Equal(int64(10)))
			gomega.Expect(stats.WaitDuration).To(gomega.Equal(time.Second))
			gomega.Expect(stats.MaxIdleClosed).To(gomega.Equal(int64(1)))
			gomega.Expect(stats.MaxLifetimeClosed).To(gomega.Equal(int64(2)))
		})
	})

	ginkgo.Describe("DatabaseStatus", func() {
		ginkgo.Describe("NewDatabaseStatus", func() {
			ginkgo.It("should create a new database status with the given parameters", func() {
				// Arrange
				expectedStatus := "healthy"
				expectedMessage := "Database is healthy"
				expectedStats := &domain.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    1,
				}

				// Act
				dbStatus := domain.NewDatabaseStatus(expectedStatus, expectedMessage, expectedStats)

				// Assert
				gomega.Expect(dbStatus).NotTo(gomega.BeNil())
				gomega.Expect(dbStatus.Status).To(gomega.Equal(expectedStatus))
				gomega.Expect(dbStatus.Message).To(gomega.Equal(expectedMessage))
				gomega.Expect(dbStatus.Stats).To(gomega.Equal(expectedStats))
			})
		})

		ginkgo.Describe("HealthyDatabaseStatus", func() {
			ginkgo.It("should create a healthy database status", func() {
				// Arrange
				expectedMessage := "Database connection is healthy"
				expectedStats := &domain.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    1,
				}

				// Act
				dbStatus := domain.HealthyDatabaseStatus(expectedMessage, expectedStats)

				// Assert
				gomega.Expect(dbStatus).NotTo(gomega.BeNil())
				gomega.Expect(dbStatus.Status).To(gomega.Equal("healthy"))
				gomega.Expect(dbStatus.Message).To(gomega.Equal(expectedMessage))
				gomega.Expect(dbStatus.Stats).To(gomega.Equal(expectedStats))
			})
		})

		ginkgo.Describe("UnhealthyDatabaseStatus", func() {
			ginkgo.It("should create an unhealthy database status", func() {
				// Arrange
				expectedMessage := "Database connection failed"

				// Act
				dbStatus := domain.UnhealthyDatabaseStatus(expectedMessage)

				// Assert
				gomega.Expect(dbStatus).NotTo(gomega.BeNil())
				gomega.Expect(dbStatus.Status).To(gomega.Equal("unhealthy"))
				gomega.Expect(dbStatus.Message).To(gomega.Equal(expectedMessage))
				gomega.Expect(dbStatus.Stats).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("DetailedHealthStatus", func() {
		ginkgo.Describe("NewDetailedHealthStatus", func() {
			ginkgo.It("should create a new detailed health status with the given parameters", func() {
				// Arrange
				expectedStatus := "healthy"
				expectedMessage := "All systems healthy"
				expectedTimestamp := "2023-01-01T00:00:00Z"
				expectedDatabase := domain.HealthyDatabaseStatus("DB healthy", nil)

				// Act
				detailedStatus := domain.NewDetailedHealthStatus(expectedStatus, expectedMessage, expectedTimestamp, expectedDatabase)

				// Assert
				gomega.Expect(detailedStatus).NotTo(gomega.BeNil())
				gomega.Expect(detailedStatus.Status).To(gomega.Equal(expectedStatus))
				gomega.Expect(detailedStatus.Message).To(gomega.Equal(expectedMessage))
				gomega.Expect(detailedStatus.Timestamp).To(gomega.Equal(expectedTimestamp))
				gomega.Expect(detailedStatus.Database).To(gomega.Equal(expectedDatabase))
			})
		})
	})
})
