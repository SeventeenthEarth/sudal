package integration_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

var _ = Describe("Domain Function Integration Tests", func() {
	Describe("DetailedHealthStatus Functions", func() {
		Context("when creating detailed health status", func() {
			Describe("NewDetailedHealthStatus", func() {
				It("should create a detailed health status with all parameters", func() {
					// Given: Valid parameters for detailed health status
					status := "healthy"
					message := "All systems operational"
					timestamp := time.Now().UTC().Format(time.RFC3339)

					stats := &entity.ConnectionStats{
						MaxOpenConnections: 25,
						OpenConnections:    10,
						InUse:              5,
						Idle:               5,
						WaitCount:          0,
						WaitDuration:       0,
						MaxIdleClosed:      0,
						MaxLifetimeClosed:  0,
					}
					databaseStatus := entity.HealthyDatabaseStatus("Database is healthy", stats)

					// When: Creating detailed health status using NewDetailedHealthStatus
					detailedStatus := entity.NewDetailedHealthStatus(status, message, timestamp, databaseStatus)

					// Then: Should create proper detailed health status
					Expect(detailedStatus).NotTo(BeNil())
					Expect(detailedStatus.Status).To(Equal(status))
					Expect(detailedStatus.Message).To(Equal(message))
					Expect(detailedStatus.Timestamp).To(Equal(timestamp))
					Expect(detailedStatus.Database).To(Equal(databaseStatus))
				})

				It("should create detailed health status with nil database", func() {
					// Given: Parameters with nil database status
					status := "error"
					message := "Database unavailable"
					timestamp := time.Now().UTC().Format(time.RFC3339)

					// When: Creating detailed health status with nil database
					detailedStatus := entity.NewDetailedHealthStatus(status, message, timestamp, nil)

					// Then: Should create proper detailed health status with nil database
					Expect(detailedStatus).NotTo(BeNil())
					Expect(detailedStatus.Status).To(Equal(status))
					Expect(detailedStatus.Message).To(Equal(message))
					Expect(detailedStatus.Timestamp).To(Equal(timestamp))
					Expect(detailedStatus.Database).To(BeNil())
				})

				It("should create detailed health status with unhealthy database", func() {
					// Given: Parameters with unhealthy database status
					status := "degraded"
					message := "Database connection issues"
					timestamp := time.Now().UTC().Format(time.RFC3339)
					databaseStatus := entity.UnhealthyDatabaseStatus("Connection timeout")

					// When: Creating detailed health status with unhealthy database
					detailedStatus := entity.NewDetailedHealthStatus(status, message, timestamp, databaseStatus)

					// Then: Should create proper detailed health status
					Expect(detailedStatus).NotTo(BeNil())
					Expect(detailedStatus.Status).To(Equal(status))
					Expect(detailedStatus.Message).To(Equal(message))
					Expect(detailedStatus.Timestamp).To(Equal(timestamp))
					Expect(detailedStatus.Database).To(Equal(databaseStatus))
					Expect(detailedStatus.Database.Status).To(Equal("unhealthy"))
					Expect(detailedStatus.Database.Message).To(Equal("Connection timeout"))
					Expect(detailedStatus.Database.Stats).To(BeNil())
				})

				It("should handle empty string parameters", func() {
					// Given: Empty string parameters
					status := ""
					message := ""
					timestamp := ""

					// When: Creating detailed health status with empty strings
					detailedStatus := entity.NewDetailedHealthStatus(status, message, timestamp, nil)

					// Then: Should create detailed health status with empty values
					Expect(detailedStatus).NotTo(BeNil())
					Expect(detailedStatus.Status).To(Equal(""))
					Expect(detailedStatus.Message).To(Equal(""))
					Expect(detailedStatus.Timestamp).To(Equal(""))
					Expect(detailedStatus.Database).To(BeNil())
				})

				It("should preserve complex connection statistics", func() {
					// Given: Complex connection statistics
					status := "healthy"
					message := "Database performance optimal"
					timestamp := time.Now().UTC().Format(time.RFC3339)

					stats := &entity.ConnectionStats{
						MaxOpenConnections: 100,
						OpenConnections:    75,
						InUse:              45,
						Idle:               30,
						WaitCount:          1234,
						WaitDuration:       time.Duration(5 * time.Second),
						MaxIdleClosed:      567,
						MaxLifetimeClosed:  89,
					}
					databaseStatus := entity.HealthyDatabaseStatus("High performance database", stats)

					// When: Creating detailed health status with complex stats
					detailedStatus := entity.NewDetailedHealthStatus(status, message, timestamp, databaseStatus)

					// Then: Should preserve all connection statistics
					Expect(detailedStatus).NotTo(BeNil())
					Expect(detailedStatus.Database.Stats).NotTo(BeNil())
					Expect(detailedStatus.Database.Stats.MaxOpenConnections).To(Equal(100))
					Expect(detailedStatus.Database.Stats.OpenConnections).To(Equal(75))
					Expect(detailedStatus.Database.Stats.InUse).To(Equal(45))
					Expect(detailedStatus.Database.Stats.Idle).To(Equal(30))
					Expect(detailedStatus.Database.Stats.WaitCount).To(Equal(int64(1234)))
					Expect(detailedStatus.Database.Stats.WaitDuration).To(Equal(time.Duration(5 * time.Second)))
					Expect(detailedStatus.Database.Stats.MaxIdleClosed).To(Equal(int64(567)))
					Expect(detailedStatus.Database.Stats.MaxLifetimeClosed).To(Equal(int64(89)))
				})
			})
		})
	})
})
