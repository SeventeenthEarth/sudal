package database_test

import (
	"errors"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("Database Utils Unit Tests", func() {
	var (
		ctrl        *gomock.Controller
		mockManager *mocks.MockPostgresManager
	)

	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)
		ctrl = gomock.NewController(ginkgo.GinkgoT())
		mockManager = mocks.NewMockPostgresManager(ctrl)
	})

	ginkgo.AfterEach(func() {
		ctrl.Finish()
	})

	ginkgo.Describe("GetConnectionPoolStats", func() {
		ginkgo.Context("when getting connection pool statistics", func() {
			ginkgo.It("should return connection stats when manager is healthy", func() {
				// Given
				expectedStats := &database.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    5,
					InUse:              2,
					Idle:               3,
					WaitCount:          0,
					WaitDuration:       0,
					MaxIdleClosed:      0,
					MaxLifetimeClosed:  0,
				}
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats:   expectedStats,
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When
				stats := database.GetConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
				gomega.Expect(stats.MaxOpenConnections).To(gomega.Equal(25))
				gomega.Expect(stats.OpenConnections).To(gomega.Equal(5))
			})

			ginkgo.It("should return nil when manager is nil", func() {
				// Given
				var nilManager database.PostgresManager = nil

				// When
				stats := database.GetConnectionPoolStats(nilManager)

				// Then
				gomega.Expect(stats).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when health check fails", func() {
				// Given
				expectedError := errors.New("health check failed")
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(nil, expectedError)

				// When
				stats := database.GetConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when health status has no stats", func() {
				// Given
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats:   nil, // No stats available
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When
				stats := database.GetConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when testing different connection pool scenarios", func() {
			ginkgo.It("should handle high connection usage", func() {
				// Given
				expectedStats := &database.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    24,
					InUse:              20,
					Idle:               4,
					WaitCount:          10,
					WaitDuration:       1000000000, // 1 second in nanoseconds
					MaxIdleClosed:      5,
					MaxLifetimeClosed:  2,
				}
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection pool under high load",
					Stats:   expectedStats,
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When
				stats := database.GetConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
				gomega.Expect(stats.WaitCount).To(gomega.BeNumerically(">", 0))
				gomega.Expect(stats.WaitDuration).To(gomega.BeNumerically(">", 0))
			})

			ginkgo.It("should handle pool exhaustion scenario", func() {
				// Given
				expectedStats := &database.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    25,
					InUse:              25,
					Idle:               0,
					WaitCount:          100,
					WaitDuration:       5000000000, // 5 seconds in nanoseconds
					MaxIdleClosed:      0,
					MaxLifetimeClosed:  0,
				}
				expectedHealthStatus := &database.HealthStatus{
					Status:  "unhealthy",
					Message: "Database connection pool exhausted",
					Stats:   expectedStats,
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When
				stats := database.GetConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
				gomega.Expect(stats.OpenConnections).To(gomega.Equal(stats.MaxOpenConnections))
				gomega.Expect(stats.Idle).To(gomega.Equal(0))
				gomega.Expect(stats.WaitCount).To(gomega.BeNumerically(">", 50))
			})
		})
	})

	ginkgo.Describe("LogConnectionPoolStats", func() {
		ginkgo.Context("when logging connection pool statistics", func() {
			ginkgo.It("should log stats when manager is healthy", func() {
				// Given
				expectedStats := &database.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    5,
					InUse:              2,
					Idle:               3,
					WaitCount:          0,
					WaitDuration:       0,
					MaxIdleClosed:      0,
					MaxLifetimeClosed:  0,
				}
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats:   expectedStats,
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When - This should not panic and should log the stats
				database.LogConnectionPoolStats(mockManager)

				// Then - No assertion needed as this is a logging function
				// The test passes if no panic occurs
			})

			ginkgo.It("should handle nil manager gracefully", func() {
				// Given
				var nilManager database.PostgresManager = nil

				// When - This should not panic
				database.LogConnectionPoolStats(nilManager)

				// Then - No assertion needed as this should handle nil gracefully
			})

			ginkgo.It("should handle health check failure gracefully", func() {
				// Given
				expectedError := errors.New("health check failed")
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(nil, expectedError)

				// When - This should not panic
				database.LogConnectionPoolStats(mockManager)

				// Then - No assertion needed as this should handle errors gracefully
			})

			ginkgo.It("should handle nil stats gracefully", func() {
				// Given
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats:   nil, // No stats available
				}
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil)

				// When - This should not panic
				database.LogConnectionPoolStats(mockManager)

				// Then - No assertion needed as this should handle nil stats gracefully
			})
		})
	})

	ginkgo.Describe("Integration scenarios", func() {
		ginkgo.Context("when using utils functions together", func() {
			ginkgo.It("should get stats and then log them", func() {
				// Given
				expectedStats := &database.ConnectionStats{
					MaxOpenConnections: 25,
					OpenConnections:    10,
					InUse:              5,
					Idle:               5,
					WaitCount:          2,
					WaitDuration:       100000000, // 100ms in nanoseconds
					MaxIdleClosed:      1,
					MaxLifetimeClosed:  0,
				}
				expectedHealthStatus := &database.HealthStatus{
					Status:  "healthy",
					Message: "Database connection is healthy",
					Stats:   expectedStats,
				}
				// Expect two calls since both functions will call HealthCheck
				mockManager.EXPECT().HealthCheck(gomock.Any()).Return(expectedHealthStatus, nil).Times(2)

				// When
				stats := database.GetConnectionPoolStats(mockManager)
				database.LogConnectionPoolStats(mockManager)

				// Then
				gomega.Expect(stats).To(gomega.Equal(expectedStats))
				// LogConnectionPoolStats doesn't return anything, but should not panic
			})
		})
	})
})
