package di_test

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

var _ = ginkgo.Describe("DI", func() {
	var originalConfig *config.Config

	ginkgo.BeforeEach(func() {
		// Set test environment variables
		os.Setenv("GINKGO_TEST", "1") // nolint:errcheck
		os.Setenv("APP_ENV", "test")  // nolint:errcheck

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
		os.Unsetenv("GINKGO_TEST") // nolint:errcheck
		os.Unsetenv("APP_ENV")     // nolint:errcheck
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

	ginkgo.Describe("ProvidePostgresManager", func() {
		ginkgo.Context("when in test environment", func() {
			ginkgo.It("should return nil for test environment", func() {
				// Arrange
				os.Setenv("GO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GO_TEST") // nolint:errcheck

				// Act
				manager, err := di.ProvidePostgresManager(originalConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when GINKGO_TEST is set", func() {
				// Arrange
				os.Setenv("GINKGO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GINKGO_TEST") // nolint:errcheck

				// Act
				manager, err := di.ProvidePostgresManager(originalConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config AppEnv is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.AppEnv = "test"

				// Act
				manager, err := di.ProvidePostgresManager(&testConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config Environment is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.Environment = "test"

				// Act
				manager, err := di.ProvidePostgresManager(&testConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when not in test environment", func() {
			ginkgo.BeforeEach(func() {
				// Clear test environment variables
				os.Unsetenv("GO_TEST")     // nolint:errcheck
				os.Unsetenv("GINKGO_TEST") // nolint:errcheck
				// Set production config
				prodConfig := *originalConfig
				prodConfig.AppEnv = "production"
				prodConfig.Environment = "production"
				config.SetConfig(&prodConfig)
			})

			ginkgo.It("should return error for invalid configuration", func() {
				// Arrange
				invalidConfig := &config.Config{
					AppEnv:      "production",
					Environment: "production",
					DB: config.DBConfig{
						DSN: "", // Invalid empty DSN
					},
				}

				// Act
				manager, err := di.ProvidePostgresManager(invalidConfig)

				// Assert
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("ProvideRedisManager", func() {
		ginkgo.Context("when in test environment", func() {
			ginkgo.It("should return nil for test environment", func() {
				// Arrange
				os.Setenv("GO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GO_TEST") // nolint:errcheck

				// Act
				manager, err := di.ProvideRedisManager(originalConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when GINKGO_TEST is set", func() {
				// Arrange
				os.Setenv("GINKGO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GINKGO_TEST") // nolint:errcheck

				// Act
				manager, err := di.ProvideRedisManager(originalConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config AppEnv is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.AppEnv = "test"

				// Act
				manager, err := di.ProvideRedisManager(&testConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config Environment is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.Environment = "test"

				// Act
				manager, err := di.ProvideRedisManager(&testConfig)

				// Assert
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when not in test environment", func() {
			ginkgo.BeforeEach(func() {
				// Clear test environment variables
				os.Unsetenv("GO_TEST")     // nolint:errcheck
				os.Unsetenv("GINKGO_TEST") // nolint:errcheck
				// Set production config
				prodConfig := *originalConfig
				prodConfig.AppEnv = "production"
				prodConfig.Environment = "production"
				config.SetConfig(&prodConfig)
			})

			ginkgo.It("should return error for invalid configuration", func() {
				// Arrange
				invalidConfig := &config.Config{
					AppEnv:      "production",
					Environment: "production",
					Redis: config.RedisConfig{
						Addr: "", // Invalid empty address
					},
				}

				// Act
				manager, err := di.ProvideRedisManager(invalidConfig)

				// Assert
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(manager).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("ProvideCacheUtil", func() {
		ginkgo.Context("when in test environment", func() {
			ginkgo.It("should return nil for test environment", func() {
				// Arrange
				os.Setenv("GO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GO_TEST") // nolint:errcheck

				// Act
				cacheUtil := di.ProvideCacheUtil(nil)

				// Assert
				gomega.Expect(cacheUtil).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when GINKGO_TEST is set", func() {
				// Arrange
				os.Setenv("GINKGO_TEST", "1")    // nolint:errcheck
				defer os.Unsetenv("GINKGO_TEST") // nolint:errcheck

				// Act
				cacheUtil := di.ProvideCacheUtil(nil)

				// Assert
				gomega.Expect(cacheUtil).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config AppEnv is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.AppEnv = "test"
				config.SetConfig(&testConfig)

				// Act
				cacheUtil := di.ProvideCacheUtil(nil)

				// Assert
				gomega.Expect(cacheUtil).To(gomega.BeNil())
			})

			ginkgo.It("should return nil when config Environment is test", func() {
				// Arrange
				testConfig := *originalConfig
				testConfig.Environment = "test"
				config.SetConfig(&testConfig)

				// Act
				cacheUtil := di.ProvideCacheUtil(nil)

				// Assert
				gomega.Expect(cacheUtil).To(gomega.BeNil())
			})
		})

		ginkgo.Context("when not in test environment", func() {
			ginkgo.BeforeEach(func() {
				// Clear test environment variables
				os.Unsetenv("GO_TEST")     // nolint:errcheck
				os.Unsetenv("GINKGO_TEST") // nolint:errcheck
				// Set production config
				prodConfig := *originalConfig
				prodConfig.AppEnv = "production"
				prodConfig.Environment = "production"
				config.SetConfig(&prodConfig)
			})

			ginkgo.It("should return cache util when redis manager is nil", func() {
				// Act
				cacheUtil := di.ProvideCacheUtil(nil)

				// Assert
				gomega.Expect(cacheUtil).NotTo(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("NewOpenAPIHandler", func() {
		ginkgo.It("should create a new OpenAPI handler", func() {
			// Arrange - create a mock service
			ctrl := gomock.NewController(ginkgo.GinkgoT())
			defer ctrl.Finish()

			mockService := mocks.NewMockHealthService(ctrl)

			// Act
			handler := di.NewOpenAPIHandler(mockService)

			// Assert
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("InitializeSwaggerHandler", func() {
		ginkgo.It("should create a new Swagger handler", func() {
			// Act
			handler := di.InitializeSwaggerHandler()

			// Assert
			gomega.Expect(handler).NotTo(gomega.BeNil())
		})
	})
})
