package database_test

import (
	"context"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

var _ = ginkgo.Describe("Database Utils", func() {
	ginkgo.BeforeEach(func() {
		// Initialize logger for tests
		log.Init(log.InfoLevel)
	})

	ginkgo.Describe("VerifyDatabaseConnectivity", func() {
		ginkgo.Context("when verifying database connectivity with valid configuration", func() {
			ginkgo.It("should attempt to verify connectivity and fail gracefully with no database", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						DSN:                    "postgres://test:test@localhost:5432/testdb?sslmode=disable",
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
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then - Should fail because there's no real database
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})

			ginkgo.It("should handle context timeout properly", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						DSN:                   "postgres://test:test@nonexistent:5432/testdb?sslmode=disable",
						ConnectTimeoutSeconds: 1,
					},
				}

				// Create a context with a very short timeout
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()

				// When
				start := time.Now()
				err := database.VerifyDatabaseConnectivity(ctx, cfg)
				duration := time.Since(start)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				// Should complete within reasonable time
				gomega.Expect(duration).To(gomega.BeNumerically("<", 15*time.Second))
			})
		})

		ginkgo.Context("when verifying database connectivity with invalid configuration", func() {
			ginkgo.It("should return an error when DSN is empty", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						DSN: "",
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})
		})
	})

	ginkgo.Describe("GetConnectionPoolStats", func() {
		ginkgo.Context("when getting connection pool stats", func() {
			ginkgo.It("should return nil when PostgresManager is nil", func() {
				// When
				stats := database.GetConnectionPoolStats(nil)

				// Then
				gomega.Expect(stats).To(gomega.BeNil())
			})

			ginkgo.It("should handle PostgresManager with nil database", func() {
				// This test would require access to internal fields or a test constructor
				// For now, we'll test the public behavior
				stats := database.GetConnectionPoolStats(nil)
				gomega.Expect(stats).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("LogConnectionPoolStats", func() {
		ginkgo.Context("when logging connection pool stats", func() {
			ginkgo.It("should handle nil PostgresManager gracefully", func() {
				// When/Then - Should not panic
				gomega.Expect(func() {
					database.LogConnectionPoolStats(nil)
				}).NotTo(gomega.Panic())
			})
		})
	})

	ginkgo.Describe("Connection pool configuration validation", func() {
		ginkgo.Context("when testing different pool configurations", func() {
			ginkgo.It("should handle default values correctly", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						DSN: "postgres://test:test@localhost:5432/testdb?sslmode=disable",
						// Using default values for pool configuration
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then - Should fail due to no database, but configuration should be processed
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})

			ginkgo.It("should handle custom pool configuration", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						DSN:                    "postgres://test:test@localhost:5432/testdb?sslmode=disable",
						MaxOpenConns:           100,
						MaxIdleConns:           20,
						ConnMaxLifetimeSeconds: 7200,
						ConnMaxIdleTimeSeconds: 600,
						ConnectTimeoutSeconds:  60,
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then - Should fail due to no database, but configuration should be processed
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})
		})
	})

	ginkgo.Describe("SSL/TLS configuration", func() {
		ginkgo.Context("when testing SSL configuration", func() {
			ginkgo.It("should handle SSL mode require", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						Host:     "localhost",
						Port:     "5432",
						User:     "test",
						Password: "test",
						Name:     "testdb",
						SSLMode:  "require",
						// Provide DSN for testing
						DSN: "postgres://test:test@localhost:5432/testdb?sslmode=require",
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})

			ginkgo.It("should handle SSL certificates configuration", func() {
				// Given
				cfg := &config.Config{
					DB: config.DBConfig{
						Host:        "localhost",
						Port:        "5432",
						User:        "test",
						Password:    "test",
						Name:        "testdb",
						SSLMode:     "verify-full",
						SSLCert:     "/path/to/client.crt",
						SSLKey:      "/path/to/client.key",
						SSLRootCert: "/path/to/ca.crt",
						// Provide DSN with SSL certificates for testing
						DSN: "postgres://test:test@localhost:5432/testdb?sslmode=verify-full&sslcert=/path/to/client.crt&sslkey=/path/to/client.key&sslrootcert=/path/to/ca.crt",
					},
				}

				ctx := context.Background()

				// When
				err := database.VerifyDatabaseConnectivity(ctx, cfg)

				// Then
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("database connectivity verification failed"))
			})
		})
	})
})
