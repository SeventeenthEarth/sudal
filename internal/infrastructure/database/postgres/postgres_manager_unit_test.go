package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	postgresdb "github.com/seventeenthearth/sudal/internal/infrastructure/database/postgres"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

var _ = ginkgo.Describe("PostgresManagerImpl Unit Tests", func() {
	var (
		db         *sql.DB
		mock       sqlmock.Sqlmock
		manager    *postgresdb.PostgresManagerImpl
		ctx        context.Context
		testConfig *config.Config
	)

	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)

		var err error
		// Create sqlmock with ping monitoring enabled
		db, mock, err = sqlmock.New(sqlmock.MonitorPingsOption(true))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		ctx = context.Background()

		// Create test config
		testConfig = &config.Config{
			DB: config.DBConfig{
				DSN:                    "postgres://test:test@localhost:5432/testdb?sslmode=disable",
				MaxOpenConns:           25,
				MaxIdleConns:           5,
				ConnMaxLifetimeSeconds: 3600,
				ConnMaxIdleTimeSeconds: 300,
				ConnectTimeoutSeconds:  30,
			},
		}

		// Create PostgresManagerImpl with mocked DB
		manager = createTestPostgresManager(db, testConfig)
	})

	ginkgo.AfterEach(func() {
		if db != nil {
			db.Close() // nolint:errcheck
		}
		// Verify all expectations were met
		gomega.Expect(mock.ExpectationsWereMet()).To(gomega.BeNil())
	})

	ginkgo.Describe("GetDB", func() {
		ginkgo.Context("when getting database connection", func() {
			ginkgo.It("should return the database connection", func() {
				// When
				result := manager.GetDB()

				// Then
				gomega.Expect(result).To(gomega.Equal(db))
			})
		})
	})

	ginkgo.Describe("Ping", func() {
		ginkgo.Context("when performing health check", func() {
			ginkgo.It("should return nil when database is healthy", func() {
				// Given
				mock.ExpectPing()

				// When
				err := manager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return error when database is unhealthy", func() {
				// Given
				expectedError := errors.New("database connection failed")
				mock.ExpectPing().WillReturnError(expectedError)

				// When
				err := manager.Ping(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database health check failed"))
			})

			ginkgo.It("should handle context timeout", func() {
				// Given
				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
				defer cancel()
				expectedError := context.DeadlineExceeded
				mock.ExpectPing().WillReturnError(expectedError)

				// When
				err := manager.Ping(timeoutCtx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database health check failed"))
			})
		})
	})

	ginkgo.Describe("HealthCheck", func() {
		ginkgo.Context("when performing comprehensive health check", func() {
			ginkgo.It("should return healthy status with connection stats", func() {
				// Given
				mock.ExpectPing()

				// When
				status, err := manager.HealthCheck(ctx)

				// Then
				gomega.Expect(err).To(gomega.BeNil())
				gomega.Expect(status).ToNot(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("healthy"))
				gomega.Expect(status.Message).To(gomega.Equal("Database connection is healthy"))
				gomega.Expect(status.Stats).ToNot(gomega.BeNil())
			})

			ginkgo.It("should return error when health check fails", func() {
				// Given
				expectedError := errors.New("health check failed")
				mock.ExpectPing().WillReturnError(expectedError)

				// When
				status, err := manager.HealthCheck(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(status).ToNot(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("unhealthy"))
			})

			ginkgo.It("should return unhealthy status when ping fails", func() {
				// Given
				pingError := errors.New("connection lost")
				mock.ExpectPing().WillReturnError(pingError)

				// When
				status, err := manager.HealthCheck(ctx)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(status).ToNot(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("unhealthy"))
				gomega.Expect(status.Message).To(gomega.ContainSubstring("database health check failed"))
			})
		})
	})

	ginkgo.Describe("Close", func() {
		ginkgo.Context("when closing database connection", func() {
			ginkgo.It("should close successfully", func() {
				// Given
				mock.ExpectClose()

				// When
				err := manager.Close()

				// Then
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should return error when close fails", func() {
				// Given
				expectedError := errors.New("failed to close database connection")
				mock.ExpectClose().WillReturnError(expectedError)

				// When
				err := manager.Close()

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to close database connection pool"))
			})
		})
	})

	ginkgo.Describe("Integration scenarios", func() {
		ginkgo.Context("when performing multiple operations", func() {
			ginkgo.It("should handle ping followed by health check", func() {
				// Given
				mock.ExpectPing()
				mock.ExpectPing() // HealthCheck also calls Ping internally

				// When
				pingErr := manager.Ping(ctx)
				status, healthErr := manager.HealthCheck(ctx)

				// Then
				gomega.Expect(pingErr).To(gomega.BeNil())
				gomega.Expect(healthErr).To(gomega.BeNil())
				gomega.Expect(status.Status).To(gomega.Equal("healthy"))
			})

			ginkgo.It("should handle failed ping followed by close", func() {
				// Given
				pingError := errors.New("connection lost")
				mock.ExpectPing().WillReturnError(pingError)
				mock.ExpectClose()

				// When
				pingErr := manager.Ping(ctx)
				closeErr := manager.Close()

				// Then
				gomega.Expect(pingErr).To(gomega.HaveOccurred())
				gomega.Expect(pingErr.Error()).To(gomega.ContainSubstring("database health check failed"))
				gomega.Expect(closeErr).To(gomega.BeNil())
			})
		})
	})
})

// Helper function to create a PostgresManagerImpl with injected dependencies
func createTestPostgresManager(db *sql.DB, cfg *config.Config) *postgresdb.PostgresManagerImpl {
	logger := log.GetLogger().With(zap.String("component", "postgres_manager"))

	// Create a PostgresManagerImpl instance and manually set its fields using reflection
	manager := &postgresdb.PostgresManagerImpl{}

	// Use reflection to set private fields
	v := reflect.ValueOf(manager).Elem()

	// Set db field
	dbField := v.FieldByName("db")
	if dbField.IsValid() {
		dbField = reflect.NewAt(dbField.Type(), dbField.Addr().UnsafePointer()).Elem()
		dbField.Set(reflect.ValueOf(db))
	}

	// Set config field
	configField := v.FieldByName("config")
	if configField.IsValid() {
		configField = reflect.NewAt(configField.Type(), configField.Addr().UnsafePointer()).Elem()
		configField.Set(reflect.ValueOf(cfg))
	}

	// Set logger field
	loggerField := v.FieldByName("logger")
	if loggerField.IsValid() {
		loggerField = reflect.NewAt(loggerField.Type(), loggerField.Addr().UnsafePointer()).Elem()
		loggerField.Set(reflect.ValueOf(logger))
	}

	return manager
}
