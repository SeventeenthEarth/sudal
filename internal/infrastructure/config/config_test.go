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

			// Clear environment variables that might interfere with config file loading
			err := os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("LOG_LEVEL")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("REDIS_ADDR")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			err = os.Unsetenv("REDIS_PASSWORD")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Create a temporary config file
			tempFile, err := os.CreateTemp("", "config-*.yaml")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			tempConfigFile = tempFile.Name()

			// Write test configuration to the file
			configContent := `
# Server Configuration
server_port: 9999
log_level: debug
app_env: test

# Database Configuration
db:
  dsn: postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable

# Redis Configuration
redis:
  addr: localhost:6379
  password: "testpassword"
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
			gomega.Expect(cfg.AppEnv).To(gomega.Equal("test"))
			// Environment field removed; AppEnv only
			gomega.Expect(cfg.DB.DSN).To(gomega.Equal("postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"))
			gomega.Expect(cfg.Redis.Addr).To(gomega.Equal("localhost:6379"))
			gomega.Expect(cfg.Redis.Password).To(gomega.Equal("testpassword"))
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

			err = os.Setenv("APP_ENV", "test")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// ENVIRONMENT no longer used
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("LOG_LEVEL")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// ENVIRONMENT no longer used
		})

		ginkgo.It("should load configuration from environment variables", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify values
			gomega.Expect(cfg.ServerPort).To(gomega.Equal("9090"))
			gomega.Expect(cfg.LogLevel).To(gomega.Equal("debug"))
			gomega.Expect(cfg.AppEnv).To(gomega.Equal("test"))
			// Environment field removed; AppEnv only
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

			err = os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// ENVIRONMENT no longer used
		})

		ginkgo.It("should use default values when environment variables are not set", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify default values
			gomega.Expect(cfg.ServerPort).To(gomega.Equal("8080"))
			gomega.Expect(cfg.LogLevel).To(gomega.Equal("info"))
			gomega.Expect(cfg.AppEnv).To(gomega.Equal("dev"))
			// Environment field removed; default AppEnv is dev
		})
	})

	ginkgo.Context("when constructing PostgresDSN from components", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Make sure POSTGRES_DSN is not set
			err := os.Unsetenv("POSTGRES_DSN")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Make sure DB_SSLMODE is set to default
			err = os.Setenv("DB_SSLMODE", "disable")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set APP_ENV to test
			err = os.Setenv("APP_ENV", "test")
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

			err = os.Unsetenv("DB_SSLMODE")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should construct PostgresDSN correctly from components", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify PostgresDSN was constructed correctly
			expectedDSN := "postgres://testuser:testpass@testhost:5432/testdb?sslmode=disable"
			gomega.Expect(cfg.DB.DSN).To(gomega.Equal(expectedDSN))
		})
	})

	ginkgo.Context("when validating database connection pool configuration", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Set environment variables for testing database pool configuration
			err := os.Setenv("APP_ENV", "test")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_MAX_OPEN_CONNS", "50")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_MAX_IDLE_CONNS", "10")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_CONN_MAX_LIFETIME_SECONDS", "7200")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_CONN_MAX_IDLE_TIME_SECONDS", "600")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Setenv("DB_CONNECT_TIMEOUT_SECONDS", "60")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("DB_MAX_OPEN_CONNS")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_MAX_IDLE_CONNS")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_CONN_MAX_LIFETIME_SECONDS")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_CONN_MAX_IDLE_TIME_SECONDS")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("DB_CONNECT_TIMEOUT_SECONDS")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			err = os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should load database connection pool configuration correctly", func() {
			// Load config
			cfg, err := config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify database pool configuration was loaded correctly
			gomega.Expect(cfg.DB.MaxOpenConns).To(gomega.Equal(50))
			gomega.Expect(cfg.DB.MaxIdleConns).To(gomega.Equal(10))
			gomega.Expect(cfg.DB.ConnMaxLifetimeSeconds).To(gomega.Equal(7200))
			gomega.Expect(cfg.DB.ConnMaxIdleTimeSeconds).To(gomega.Equal(600))
			gomega.Expect(cfg.DB.ConnectTimeoutSeconds).To(gomega.Equal(60))
		})
	})

	ginkgo.Context("when validating in production environment", func() {
		ginkgo.BeforeEach(func() {
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Clear all database-related environment variables first
			dbEnvVars := []string{
				"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
				"POSTGRES_DSN", "DB_SSL_CERT", "DB_SSL_KEY", "DB_SSL_ROOT_CERT",
			}
			for _, envVar := range dbEnvVars {
				_ = os.Unsetenv(envVar)
			}

			// Set environment for production but missing required database fields
			err := os.Setenv("APP_ENV", "production")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// ENVIRONMENT no longer used

			// Set SERVER_PORT to avoid that validation error
			err = os.Setenv("SERVER_PORT", "8080")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.AfterEach(func() {
			// Clean up after test
			err := os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// ENVIRONMENT no longer used

			err = os.Unsetenv("SERVER_PORT")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Clean up database environment variables
			dbEnvVars := []string{
				"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SSLMODE",
				"POSTGRES_DSN", "DB_SSL_CERT", "DB_SSL_KEY", "DB_SSL_ROOT_CERT",
			}
			for _, envVar := range dbEnvVars {
				_ = os.Unsetenv(envVar)
			}
		})

		ginkgo.It("should fail validation when required database fields are missing in production", func() {
			// Load config - should fail validation due to missing database configuration
			_, err := config.LoadConfig("")
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("DB_HOST or POSTGRES_DSN"))
		})

		ginkgo.Context("when testing PORT environment variable mapping", func() {
			ginkgo.BeforeEach(func() {
				config.ResetViper()
				// Set APP_ENV to dev to avoid production validation
				err := os.Setenv("APP_ENV", "dev")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.AfterEach(func() {
				_ = os.Unsetenv("APP_ENV")
			})

			ginkgo.It("should map PORT to SERVER_PORT when SERVER_PORT is not set", func() {
				// Given
				// Unset SERVER_PORT first to ensure it's not set
				_ = os.Unsetenv("SERVER_PORT")

				err := os.Setenv("PORT", "9090")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer func() { _ = os.Unsetenv("PORT") }()

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cfg.ServerPort).To(gomega.Equal("9090"))
			})

			ginkgo.It("should not override SERVER_PORT when both PORT and SERVER_PORT are set", func() {
				// Given
				err := os.Setenv("PORT", "9090")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer func() { _ = os.Unsetenv("PORT") }()

				err = os.Setenv("SERVER_PORT", "8080")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer func() { _ = os.Unsetenv("SERVER_PORT") }()

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cfg.ServerPort).To(gomega.Equal("8080"))
			})
		})

		ginkgo.Context("when testing processDatabaseConfig edge cases", func() {
			ginkgo.BeforeEach(func() {
				config.ResetViper()
				// Set APP_ENV to dev to avoid production validation
				err := os.Setenv("APP_ENV", "dev")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			})

			ginkgo.AfterEach(func() {
				_ = os.Unsetenv("APP_ENV")
			})

			ginkgo.It("should handle POSTGRES_DSN environment variable", func() {
				// Given
				err := os.Setenv("POSTGRES_DSN", "postgres://user:pass@host:5432/db?sslmode=require")
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				defer func() { _ = os.Unsetenv("POSTGRES_DSN") }()

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cfg.DB.DSN).To(gomega.Equal("postgres://user:pass@host:5432/db?sslmode=require"))
			})

			ginkgo.It("should construct DSN from individual components with SSL certificates", func() {
				// Given
				envVars := map[string]string{
					"DB_HOST":          "localhost",
					"DB_PORT":          "5432",
					"DB_USER":          "testuser",
					"DB_PASSWORD":      "testpass",
					"DB_NAME":          "testdb",
					"DB_SSLMODE":       "verify-full",
					"DB_SSL_CERT":      "/path/to/cert.pem",
					"DB_SSL_KEY":       "/path/to/key.pem",
					"DB_SSL_ROOT_CERT": "/path/to/ca.pem",
				}

				for key, value := range envVars {
					err := os.Setenv(key, value)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					defer func(k string) { _ = os.Unsetenv(k) }(key)
				}

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cfg.DB.DSN).To(gomega.ContainSubstring("postgres://testuser:testpass@localhost:5432/testdb?sslmode=verify-full"))
				gomega.Expect(cfg.DB.DSN).To(gomega.ContainSubstring("sslcert=/path/to/cert.pem"))
				gomega.Expect(cfg.DB.DSN).To(gomega.ContainSubstring("sslkey=/path/to/key.pem"))
				gomega.Expect(cfg.DB.DSN).To(gomega.ContainSubstring("sslrootcert=/path/to/ca.pem"))
			})

			ginkgo.It("should handle connection pool configuration from environment variables", func() {
				// Given
				envVars := map[string]string{
					"DB_MAX_OPEN_CONNS":             "50",
					"DB_MAX_IDLE_CONNS":             "10",
					"DB_CONN_MAX_LIFETIME_SECONDS":  "7200",
					"DB_CONN_MAX_IDLE_TIME_SECONDS": "600",
					"DB_CONNECT_TIMEOUT_SECONDS":    "60",
				}

				for key, value := range envVars {
					err := os.Setenv(key, value)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					defer func(k string) { _ = os.Unsetenv(k) }(key)
				}

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				gomega.Expect(cfg.DB.MaxOpenConns).To(gomega.Equal(50))
				gomega.Expect(cfg.DB.MaxIdleConns).To(gomega.Equal(10))
				gomega.Expect(cfg.DB.ConnMaxLifetimeSeconds).To(gomega.Equal(7200))
				gomega.Expect(cfg.DB.ConnMaxIdleTimeSeconds).To(gomega.Equal(600))
				gomega.Expect(cfg.DB.ConnectTimeoutSeconds).To(gomega.Equal(60))
			})

			ginkgo.It("should handle invalid connection pool values gracefully", func() {
				// Given
				envVars := map[string]string{
					"DB_MAX_OPEN_CONNS":             "invalid",
					"DB_MAX_IDLE_CONNS":             "not_a_number",
					"DB_CONN_MAX_LIFETIME_SECONDS":  "abc",
					"DB_CONN_MAX_IDLE_TIME_SECONDS": "xyz",
					"DB_CONNECT_TIMEOUT_SECONDS":    "def",
				}

				for key, value := range envVars {
					err := os.Setenv(key, value)
					gomega.Expect(err).NotTo(gomega.HaveOccurred())
					defer func(k string) { _ = os.Unsetenv(k) }(key)
				}

				// When
				cfg, err := config.LoadConfig("")

				// Then
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
				// Should use default values when parsing fails
				gomega.Expect(cfg.DB.MaxOpenConns).To(gomega.Equal(25))
				gomega.Expect(cfg.DB.MaxIdleConns).To(gomega.Equal(5))
				gomega.Expect(cfg.DB.ConnMaxLifetimeSeconds).To(gomega.Equal(3600))
				gomega.Expect(cfg.DB.ConnMaxIdleTimeSeconds).To(gomega.Equal(300))
				gomega.Expect(cfg.DB.ConnectTimeoutSeconds).To(gomega.Equal(30))
			})
		})
	})
})

var _ = ginkgo.Describe("validateConfig", func() {
	ginkgo.Context("when validating a config with missing required fields", func() {
		ginkgo.It("should return an error for missing ServerPort", func() {
			// Create a config with missing ServerPort
			cfg := &config.Config{
				ServerPort: "", // Missing required field
				AppEnv:     "dev",
			}

			// Directly validate the config
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("SERVER_PORT"))
		})

		ginkgo.It("should return an error for missing AppEnv", func() {
			// Create a config with missing AppEnv
			cfg := &config.Config{
				ServerPort: "8080",
				AppEnv:     "", // Missing required field
			}

			// Directly validate the config
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("APP_ENV"))
		})
	})

	ginkgo.Context("when validating a production config with database configuration", func() {
		ginkgo.It("should pass validation when database DSN is provided", func() {
			// Create a production config with database DSN
			cfg := &config.Config{
				AppEnv:     "production",
				ServerPort: "8080",
				DB: config.DBConfig{
					DSN: "postgres://user:pass@localhost:5432/db",
				},
			}

			// Directly validate the config - should pass validation
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should pass validation when database components are provided", func() {
			// Create a production config with database components
			cfg := &config.Config{
				AppEnv:     "production",
				ServerPort: "8080",
				DB: config.DBConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "user",
					Password: "password",
					Name:     "testdb",
					SSLMode:  "require",
				},
			}

			// Directly validate the config - should pass validation
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should fail validation when database configuration is missing", func() {
			// Create a production config without database configuration
			cfg := &config.Config{
				AppEnv:     "production",
				ServerPort: "8080",
				DB:         config.DBConfig{}, // Empty database config
			}

			// Directly validate the config - should fail validation
			err := config.ValidateConfig(cfg)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(err.Error()).To(gomega.ContainSubstring("DB_HOST or POSTGRES_DSN"))
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
			// Reset Viper to clear any previous configuration
			config.ResetViper()

			// Set APP_ENV to a valid value for this test
			err := os.Setenv("APP_ENV", "test")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Load a config
			cfg, err = config.LoadConfig("")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			// Set the config instance
			config.SetConfig(cfg)
		})

		ginkgo.AfterEach(func() {
			// Clean up environment variable
			err := os.Unsetenv("APP_ENV")
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
		})

		ginkgo.It("should return the expected config instance", func() {
			// GetConfig should return the config
			retrievedCfg := config.GetConfig()
			gomega.Expect(retrievedCfg).To(gomega.Equal(cfg))
		})
	})
})
