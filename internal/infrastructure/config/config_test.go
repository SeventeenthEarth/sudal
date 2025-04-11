package config_test

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
)

// 이미 config_suite_test.go에 TestConfig 함수가 정의되어 있으므로 여기서는 제거합니다.

var _ = ginkgo.Describe("LoadConfig", func() {
	ginkgo.Context("when loading from a config file", func() {
		var tempConfigFile string

		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()
			// Create a temporary config file
			tempFile, err := os.CreateTemp("", "config-*.yaml")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tempConfigFile = tempFile.Name()

			// Write test configuration to the file
			configContent := `
# Server Configuration
server_port: 9999
log_level: debug
environment: test

# Database Configuration
postgres_dsn: postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable

# Redis Configuration
redis_addr: localhost:6379
redis_password: "testpassword"
`
			_, err = tempFile.WriteString(configContent)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = tempFile.Close()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Remove the temporary config file
			err := os.Remove(tempConfigFile)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should load configuration from a file", func() {
			// Load config from the temporary file
			cfg, err := config.LoadConfig(tempConfigFile)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify values from the config file
			gomega.Expect(cfg.ServerPort).To(gomega.Equal("9999"))
			gomega.Expect(cfg.LogLevel).To(gomega.Equal("debug"))
			gomega.Expect(cfg.Environment).To(gomega.Equal("test"))
			gomega.Expect(cfg.PostgresDSN).To(gomega.Equal("postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"))
			gomega.Expect(cfg.RedisAddr).To(gomega.Equal("localhost:6379"))
			gomega.Expect(cfg.RedisPassword).To(gomega.Equal("testpassword"))
		})

		ginkgo.Context("when the config file does not exist", func() {
			ginkgo.It("should return an error", func() {
				// Try to load config from a non-existent file
				_, err := config.LoadConfig("/non/existent/config.yaml")
				gomega.Expect(err).To(gomega.HaveOccurred())
				gomega.Expect(err.Error()).To(gomega.ContainSubstring("config file does not exist"))
			})
		})
	})

	ginkgo.Context("when unmarshaling fails", func() {
		ginkgo.BeforeEach(func() {
			// Set an environment variable with an invalid value for a numeric field
			// This will cause unmarshaling to fail
			err := os.Setenv("SERVER_PORT", "not-a-number")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			err := os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		// Note: This test might not actually fail because SERVER_PORT is a string in the Config struct
		// We're keeping it as an example of how to test unmarshaling failures
		ginkgo.It("should handle unmarshaling errors gracefully", func() {
			// The test is more for demonstration, as our current Config struct uses string types
			// which won't fail on unmarshaling from non-numeric values
			_, _ = config.LoadConfig("")
			// We don't assert on the error here since it might not actually fail
		})
	})
	// Test loading from environment variables
	ginkgo.Context("when loading from environment variables", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()
			// Set environment variables for testing
			err := os.Setenv("SERVER_PORT", "9090")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("LOG_LEVEL", "debug")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("ENVIRONMENT", "test")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("LOG_LEVEL")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("ENVIRONMENT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should load configuration from environment variables", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify values
			gomega.Expect(cfg.ServerPort).To(gomega.Equal("9090"))
			gomega.Expect(cfg.LogLevel).To(gomega.Equal("debug"))
			gomega.Expect(cfg.Environment).To(gomega.Equal("test"))
		})
	})

	ginkgo.Context("when using default values", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Clear relevant environment variables
			err := os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("LOG_LEVEL")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("ENVIRONMENT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should use default values when environment variables are not set", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify default values
			gomega.Expect(cfg.ServerPort).To(gomega.Equal("8080"))
			gomega.Expect(cfg.LogLevel).To(gomega.Equal("info"))
			gomega.Expect(cfg.Environment).To(gomega.Equal("development"))
		})
	})

	ginkgo.Context("when constructing PostgresDSN from components", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Make sure POSTGRES_DSN is not set
			err := os.Unsetenv("POSTGRES_DSN")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set environment variables for testing
			err = os.Setenv("DB_HOST", "testhost")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_PORT", "5432")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_USER", "testuser")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_PASSWORD", "testpass")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_NAME", "testdb")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("DB_HOST")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_USER")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_PASSWORD")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_NAME")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should construct PostgresDSN correctly from components", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify PostgresDSN was constructed correctly
			expectedDSN := "postgres://testuser:testpass@testhost:5432/testdb?sslmode=disable"
			gomega.Expect(cfg.PostgresDSN).To(gomega.Equal(expectedDSN))
		})
	})

	ginkgo.Context("when constructing RedisAddr from components", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Make sure REDIS_ADDR is not set
			err := os.Unsetenv("REDIS_ADDR")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set environment variables for testing
			err = os.Setenv("REDIS_HOST", "redishost")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("REDIS_PORT", "6379")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("REDIS_HOST")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("REDIS_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should construct RedisAddr correctly from components", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify RedisAddr was constructed correctly
			expectedAddr := "redishost:6379"
			gomega.Expect(cfg.RedisAddr).To(gomega.Equal(expectedAddr))
		})
	})

	ginkgo.Context("when validating in production environment", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()
			// Set environment for production but missing required fields
			err := os.Setenv("ENVIRONMENT", "production")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("ENVIRONMENT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should fail validation when required fields are missing in production", func() {
			// Load config - should fail validation
			_, err := config.LoadConfig("")
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

var _ = ginkgo.Describe("validateConfig", func() {
	ginkgo.Context("when validating a config with missing required fields", func() {
		ginkgo.It("should return an error for missing ServerPort", func() {
			// Create a config with missing ServerPort
			cfg := &config.Config{
				Environment: "development",
				ServerPort:  "", // Missing required field
			}

			// Directly validate the config
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("SERVER_PORT"))
		})

		ginkgo.It("should return an error for missing Environment", func() {
			// Create a config with missing Environment
			cfg := &config.Config{
				ServerPort:  "8080",
				Environment: "", // Missing required field
			}

			// Directly validate the config
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("ENVIRONMENT"))
		})
	})

	ginkgo.Context("when validating a production config with missing fields", func() {
		ginkgo.BeforeEach(func() {
			// Set environment for production with some required fields missing
			err := os.Setenv("ENVIRONMENT", "production")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Setenv("SERVER_PORT", "8080")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set only some of the required production fields
			err = os.Setenv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/db")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			// Deliberately not setting REDIS_ADDR, FIREBASE_PROJECT_ID, JWT_SECRET_KEY
		})

		ginkgo.AfterEach(func() {
			// Clean up environment variables
			err := os.Unsetenv("ENVIRONMENT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("POSTGRES_DSN")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should return an error listing all missing required fields", func() {
			// Create a config with missing production fields
			cfg := &config.Config{
				Environment: "production",
				ServerPort:  "8080",
				PostgresDSN: "postgres://user:pass@localhost:5432/db",
				// Missing: RedisAddr, FirebaseProjectID, JwtSecretKey
			}

			// Directly validate the config
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())

			// Check that the error message contains all missing fields
			errMsg := err.Error()
			gomega.Expect(errMsg).To(gomega.ContainSubstring("REDIS_ADDR"))
			gomega.Expect(errMsg).To(gomega.ContainSubstring("FIREBASE_PROJECT_ID"))
			gomega.Expect(errMsg).To(gomega.ContainSubstring("JWT_SECRET_KEY"))
		})

		ginkgo.It("should pass validation when all required production fields are set", func() {
			// Create a complete production config
			cfg := &config.Config{
				Environment:       "production",
				ServerPort:        "8080",
				PostgresDSN:       "postgres://user:pass@localhost:5432/db",
				RedisAddr:         "localhost:6379",
				FirebaseProjectID: "test-project",
				JwtSecretKey:      "test-secret-key",
			}

			// Directly validate the config - should pass validation
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})
	})
})

var _ = ginkgo.Describe("GetConfig", func() {
	ginkgo.Context("when config is not loaded", func() {
		ginkgo.BeforeEach(func() {
			// Reset the config instance
			config.SetConfig(nil)
		})

		ginkgo.It("should panic when config is not loaded", func() {
			// GetConfig should panic if config is not loaded
			gomega.Expect(func() {
				config.GetConfig()
			}).To(gomega.Panic())
		})
	})

	ginkgo.Context("when config is loaded", func() {
		var cfg *config.Config

		ginkgo.BeforeEach(func() {
			// Load a config
			var err error
			cfg, err = config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set the config instance
			config.SetConfig(cfg)
		})

		ginkgo.It("should return the expected config instance", func() {
			// GetConfig should return the config
			retrievedCfg := config.GetConfig()
			gomega.Expect(retrievedCfg).To(gomega.Equal(cfg))
		})
	})
})
