package di_test

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
)

var _ = ginkgo.Describe("DI", func() {
	var originalConfig *config.Config

	ginkgo.BeforeEach(func() {
		// Set test environment variables
		os.Setenv("GINKGO_TEST", "1")
		os.Setenv("APP_ENV", "test")

		// Load test configuration with required database DSN
		var err error
		originalConfig, err = config.LoadConfig("")
		gomega.Expect(err).NotTo(gomega.HaveOccurred())

		// Set test-specific configuration values
		originalConfig.AppEnv = "test"
		originalConfig.Environment = "test"
		originalConfig.DB.DSN = "postgres://test:test@localhost:5432/testdb?sslmode=disable"
		originalConfig.DB.Host = "localhost"
		originalConfig.DB.Port = "5432"
		originalConfig.DB.User = "test"
		originalConfig.DB.Password = "test"
		originalConfig.DB.Name = "testdb"
		originalConfig.DB.SSLMode = "disable"
		originalConfig.DB.MaxOpenConns = 25
		originalConfig.DB.MaxIdleConns = 5
		originalConfig.DB.ConnMaxLifetimeSeconds = 3600
		originalConfig.DB.ConnMaxIdleTimeSeconds = 300
		originalConfig.DB.ConnectTimeoutSeconds = 30

		// Set the global configuration instance
		config.SetConfig(originalConfig)
	})

	ginkgo.AfterEach(func() {
		// Clean up configuration
		config.SetConfig(nil)
		config.ResetViper()

		// Clean up environment variables
		os.Unsetenv("GINKGO_TEST")
		os.Unsetenv("APP_ENV")
	})

	ginkgo.Describe("InitializeHealthHandler", func() {
		ginkgo.It("should return a non-nil handler", func() {
			// Act
			handler, err := di.InitializeHealthHandler()

			// Assert
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("InitializeHealthConnectHandler", func() {
		ginkgo.It("should return a non-nil Connect handler", func() {
			// Act
			handler, err := di.InitializeHealthConnectHandler()

			// Assert
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("ProvideConfig", func() {
		ginkgo.Context("when config is loaded", func() {
			ginkgo.It("should return the config instance", func() {
				// Act
				providedCfg := di.ProvideConfig()

				// Assert
				gomega.Expect(providedCfg).To(gomega.Equal(originalConfig))
				gomega.Expect(providedCfg).To(gomega.Equal(config.GetConfig()))
				gomega.Expect(providedCfg.AppEnv).To(gomega.Equal("test"))
				gomega.Expect(providedCfg.DB.DSN).To(gomega.Equal("postgres://test:test@localhost:5432/testdb?sslmode=disable"))
			})
		})

		ginkgo.Context("when config is not loaded", func() {
			ginkgo.BeforeEach(func() {
				// Reset the config instance for this specific test
				config.SetConfig(nil)
			})

			ginkgo.AfterEach(func() {
				// Restore the config for other tests
				config.SetConfig(originalConfig)
			})

			ginkgo.It("should panic when trying to provide config", func() {
				// Act & Assert
				gomega.Expect(func() {
					di.ProvideConfig()
				}).To(gomega.Panic())
			})
		})
	})
})
