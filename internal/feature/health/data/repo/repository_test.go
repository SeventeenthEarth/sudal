package repo_test

import (
	"context"
	"errors"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/seventeenthearth/sudal/internal/feature/health/data/repo"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

var _ = ginkgo.Describe("HealthRepository", func() {
	ginkgo.Describe("NewHealthRepository", func() {
		ginkgo.It("should create a new repository", func() {
			// Act
			localRepo := repo.NewHealthRepository(nil) // nil for test environment

			// Assert
			gomega.Expect(localRepo).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("GetStatus", func() {
		var (
			localRepo *repo.HealthRepository
			ctx       context.Context
			status    *entity.HealthStatus
			err       error
		)

		ginkgo.BeforeEach(func() {
			localRepo = repo.NewHealthRepository(nil) // nil for test environment
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			status, err = localRepo.GetStatus(ctx)
		})

		ginkgo.It("should return a healthy status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(status).NotTo(gomega.BeNil())
			gomega.Expect(status.Status).To(gomega.Equal("healthy"))
		})
	})

	ginkgo.Describe("GetDatabaseStatus", func() {
		var (
			localRepo      *repo.HealthRepository
			ctx            context.Context
			databaseStatus *entity.DatabaseStatus
			err            error
		)

		ginkgo.BeforeEach(func() {
			localRepo = repo.NewHealthRepository(nil) // nil for test environment
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			databaseStatus, err = localRepo.GetDatabaseStatus(ctx)
		})

		ginkgo.It("should return a healthy database status without error", func() {
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(databaseStatus).NotTo(gomega.BeNil())
			gomega.Expect(databaseStatus.Status).To(gomega.Equal("healthy"))
			gomega.Expect(databaseStatus.Message).To(gomega.Equal("Mock database connection is healthy"))
			gomega.Expect(databaseStatus.Stats).NotTo(gomega.BeNil())
			gomega.Expect(databaseStatus.Stats.MaxOpenConnections).To(gomega.Equal(25))
		})

		ginkgo.Context("when database manager is available", func() {
			var (
				ctrl      *gomock.Controller
				mockDB    *mocks.MockPostgresManager
				testError error
			)

			ginkgo.BeforeEach(func() {
				ctrl = gomock.NewController(ginkgo.GinkgoT())
				mockDB = mocks.NewMockPostgresManager(ctrl)
				localRepo = repo.NewHealthRepository(mockDB)
				testError = errors.New("database connection failed")
			})

			ginkgo.AfterEach(func() {
				ctrl.Finish()
			})

			ginkgo.Context("when health check succeeds", func() {
				ginkgo.BeforeEach(func() {
					healthStatus := &postgres.HealthStatus{
						Status:  "healthy",
						Message: "Database connection is healthy",
						Stats: &postgres.ConnectionStats{
							MaxOpenConnections: 50,
							OpenConnections:    10,
							InUse:              5,
							Idle:               5,
							WaitCount:          0,
							WaitDuration:       0,
							MaxIdleClosed:      0,
							MaxLifetimeClosed:  0,
						},
					}
					mockDB.EXPECT().HealthCheck(gomock.Any()).Return(healthStatus, nil)
				})

				ginkgo.It("should return database status from health check", func() {
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(databaseStatus).NotTo(gomega.BeNil())
					gomega.Expect(databaseStatus.Status).To(gomega.Equal("healthy"))
					gomega.Expect(databaseStatus.Message).To(gomega.Equal("Database connection is healthy"))
					gomega.Expect(databaseStatus.Stats).NotTo(gomega.BeNil())
					gomega.Expect(databaseStatus.Stats.MaxOpenConnections).To(gomega.Equal(50))
					gomega.Expect(databaseStatus.Stats.OpenConnections).To(gomega.Equal(10))
				})
			})

			ginkgo.Context("when health check fails", func() {
				ginkgo.BeforeEach(func() {
					mockDB.EXPECT().HealthCheck(gomock.Any()).Return(nil, testError)
				})

				ginkgo.It("should return error from health check", func() {
					gomega.Expect(err).To(gomega.HaveOccurred())
					gomega.Expect(err).To(gomega.Equal(testError))
					gomega.Expect(databaseStatus).NotTo(gomega.BeNil())
					gomega.Expect(databaseStatus.Status).To(gomega.Equal("unhealthy"))
					gomega.Expect(databaseStatus.Message).To(gomega.ContainSubstring("Database health check failed"))
				})
			})

			ginkgo.Context("when health check returns unhealthy status", func() {
				ginkgo.BeforeEach(func() {
					healthStatus := &postgres.HealthStatus{
						Status:  "unhealthy",
						Message: "Database connection is unhealthy",
						Stats: &postgres.ConnectionStats{
							MaxOpenConnections: 50,
							OpenConnections:    0,
							InUse:              0,
							Idle:               0,
							WaitCount:          10,
							WaitDuration:       5000 * time.Nanosecond,
							MaxIdleClosed:      5,
							MaxLifetimeClosed:  2,
						},
					}
					mockDB.EXPECT().HealthCheck(gomock.Any()).Return(healthStatus, nil)
				})

				ginkgo.It("should return unhealthy database status", func() {
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					gomega.Expect(databaseStatus).NotTo(gomega.BeNil())
					gomega.Expect(databaseStatus.Status).To(gomega.Equal("unhealthy"))
					gomega.Expect(databaseStatus.Message).To(gomega.Equal("Database connection is unhealthy"))
					gomega.Expect(databaseStatus.Stats).NotTo(gomega.BeNil())
					gomega.Expect(databaseStatus.Stats.WaitCount).To(gomega.Equal(int64(10)))
					gomega.Expect(databaseStatus.Stats.WaitDuration).To(gomega.Equal(5000 * time.Nanosecond))
				})
			})
		})
	})

	ginkgo.Describe("GetStatus with different scenarios", func() {
		var (
			localRepo *repo.HealthRepository
			ctx       context.Context
			status    *entity.HealthStatus
			err       error
		)

		ginkgo.BeforeEach(func() {
			ctx = context.Background()
		})

		ginkgo.JustBeforeEach(func() {
			status, err = localRepo.GetStatus(ctx)
		})

		ginkgo.Context("when repository is created with nil database manager", func() {
			ginkgo.BeforeEach(func() {
				localRepo = repo.NewHealthRepository(nil)
			})

			ginkgo.It("should return healthy status", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(status).NotTo(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("healthy"))
			})
		})

		ginkgo.Context("when repository is created with valid database manager", func() {
			var (
				ctrl   *gomock.Controller
				mockDB *mocks.MockPostgresManager
			)

			ginkgo.BeforeEach(func() {
				ctrl = gomock.NewController(ginkgo.GinkgoT())
				mockDB = mocks.NewMockPostgresManager(ctrl)
				localRepo = repo.NewHealthRepository(mockDB)
			})

			ginkgo.AfterEach(func() {
				ctrl.Finish()
			})

			ginkgo.It("should return healthy status", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(status).NotTo(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("healthy"))
			})
		})

		ginkgo.Context("when context is cancelled", func() {
			var cancelledCtx context.Context

			ginkgo.BeforeEach(func() {
				localRepo = repo.NewHealthRepository(nil)
				var cancel context.CancelFunc
				cancelledCtx, cancel = context.WithCancel(context.Background())
				cancel() // Cancel immediately
				ctx = cancelledCtx
			})

			ginkgo.It("should still return healthy status", func() {
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(status).NotTo(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("healthy"))
			})
		})
	})
})
