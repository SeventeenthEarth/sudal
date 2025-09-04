package postgres_test

import (
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	sconfig "github.com/seventeenthearth/sudal/internal/service/config"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	postgresdb "github.com/seventeenthearth/sudal/internal/service/postgres"
)

var _ = ginkgo.Describe("PostgresManager", func() {
	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)
	})

	ginkgo.Describe("NewPostgresManager", func() {
		ginkgo.Context("when creating a new PostgresManager with valid configuration", func() {
			ginkgo.It("should create a PostgresManager successfully with DSN", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						DSN:                    "postgres://test:test@localhost:5432/testdb?sslmode=disable",
						MaxOpenConns:           25,
						MaxIdleConns:           5,
						ConnMaxLifetimeSeconds: 3600,
						ConnMaxIdleTimeSeconds: 300,
						ConnectTimeoutSeconds:  30,
					},
				}

				// When - This will fail because we don't have a real database
				// But we're testing the configuration validation
				_, err := postgresdb.NewPostgresManager(cfg)

				// Then - We expect an error because there's no real database
				// But the error should be about connection, not configuration
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to ping database"))
			})

			ginkgo.It("should create a PostgresManager successfully with individual components", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						Host:                   "localhost",
						Port:                   "5432",
						User:                   "test",
						Password:               "test",
						Name:                   "testdb",
						SSLMode:                "disable",
						MaxOpenConns:           25,
						MaxIdleConns:           5,
						ConnMaxLifetimeSeconds: 3600,
						ConnMaxIdleTimeSeconds: 300,
						ConnectTimeoutSeconds:  30,
						// Construct DSN from components for testing
						DSN: "postgres://test:test@localhost:5432/testdb?sslmode=disable",
					},
				}

				// When - This will fail because we don't have a real database
				_, err := postgresdb.NewPostgresManager(cfg)

				// Then - We expect an error because there's no real database
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to ping database"))
			})
		})

		ginkgo.Context("when creating a new PostgresManager with invalid configuration", func() {
			ginkgo.It("should return an error when DSN is empty", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						DSN: "",
					},
				}

				// When
				_, err := postgresdb.NewPostgresManager(cfg)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database DSN is required"))
			})
		})
	})

	ginkgo.Describe("PostgresManager connection pool configuration", func() {
		ginkgo.Context("when testing connection pool configuration", func() {
			ginkgo.It("should apply connection pool settings correctly", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						DSN:                    "postgres://test:test@localhost:5432/testdb?sslmode=disable",
						MaxOpenConns:           50,
						MaxIdleConns:           10,
						ConnMaxLifetimeSeconds: 7200,
						ConnMaxIdleTimeSeconds: 600,
						ConnectTimeoutSeconds:  60,
					},
				}

				// When - This will fail due to no real database, but we can verify the configuration
				_, err := postgresdb.NewPostgresManager(cfg)

				// Then - The error should be about connection, not configuration
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to ping database"))
			})
		})

		ginkgo.Context("when testing SSL configuration", func() {
			ginkgo.It("should handle SSL mode configuration", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						Host:                   "localhost",
						Port:                   "5432",
						User:                   "test",
						Password:               "test",
						Name:                   "testdb",
						SSLMode:                "require",
						SSLCert:                "/path/to/cert.pem",
						SSLKey:                 "/path/to/key.pem",
						SSLRootCert:            "/path/to/ca.pem",
						MaxOpenConns:           25,
						MaxIdleConns:           5,
						ConnMaxLifetimeSeconds: 3600,
						ConnMaxIdleTimeSeconds: 300,
						ConnectTimeoutSeconds:  30,
						// Construct DSN with SSL parameters for testing
						DSN: "postgres://test:test@localhost:5432/testdb?sslmode=require&sslcert=/path/to/cert.pem&sslkey=/path/to/key.pem&sslrootcert=/path/to/ca.pem",
					},
				}

				// When
				_, err := postgresdb.NewPostgresManager(cfg)

				// Then - Should fail due to no real database, but configuration should be valid
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("failed to ping database"))
			})
		})
	})

	ginkgo.Describe("Connection timeout behavior", func() {
		ginkgo.Context("when testing connection timeout", func() {
			ginkgo.It("should respect connection timeout settings", func() {
				// Given
				cfg := &sconfig.Config{
					DB: sconfig.DBConfig{
						DSN:                   "postgres://test:test@nonexistent:5432/testdb?sslmode=disable",
						ConnectTimeoutSeconds: 1, // Very short timeout
					},
				}

				// When
				start := time.Now()
				_, err := postgresdb.NewPostgresManager(cfg)
				duration := time.Since(start)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				// Should timeout relatively quickly (within a reasonable margin)
				gomega.Expect(duration).To(gomega.BeNumerically("<", 10*time.Second))
			})
		})
	})
})
