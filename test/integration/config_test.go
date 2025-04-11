package integration_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
)

var _ = Describe("Configuration System Integration", func() {
	var (
		tempDir string
		envVars map[string]string
	)

	BeforeEach(func() {
		// Create a temporary directory for config files
		var err error
		tempDir, err = os.MkdirTemp("", "config-test")
		Expect(err).NotTo(HaveOccurred())

		// Save original environment variables
		envVars = map[string]string{
			"SERVER_PORT":  os.Getenv("SERVER_PORT"),
			"LOG_LEVEL":    os.Getenv("LOG_LEVEL"),
			"ENVIRONMENT":  os.Getenv("ENVIRONMENT"),
			"POSTGRES_DSN": os.Getenv("POSTGRES_DSN"),
			"REDIS_ADDR":   os.Getenv("REDIS_ADDR"),
		}

		// Reset config for each test
		config.ResetViper()
	})

	AfterEach(func() {
		// Clean up temporary directory
		_ = os.RemoveAll(tempDir) // 오류 무시

		// Restore original environment variables
		for key, value := range envVars {
			if value == "" {
				_ = os.Unsetenv(key) // 오류 무시
			} else {
				_ = os.Setenv(key, value) // 오류 무시
			}
		}

		// Reset config
		config.SetConfig(nil)
		config.ResetViper()
	})

	Describe("LoadConfig", func() {
		Context("when loading from environment variables", func() {
			It("should load configuration from environment variables", func() {
				// Set environment variables
				_ = os.Setenv("SERVER_PORT", "9090") // 오류 무시
				_ = os.Setenv("LOG_LEVEL", "debug")  // 오류 무시
				_ = os.Setenv("ENVIRONMENT", "test") // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify config values
				Expect(cfg.ServerPort).To(Equal("9090"))
				Expect(cfg.LogLevel).To(Equal("debug"))
				Expect(cfg.Environment).To(Equal("test"))
			})
		})

		Context("when loading from a config file", func() {
			It("should load configuration from a file", func() {
				// Create a config file
				configContent := `
server_port: "8888"
log_level: "info"
environment: "development"
`
				configFile := filepath.Join(tempDir, "config.yaml")
				err := os.WriteFile(configFile, []byte(configContent), 0644)
				Expect(err).NotTo(HaveOccurred())

				// Load config
				cfg, err := config.LoadConfig(configFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify config values
				Expect(cfg.ServerPort).To(Equal("8888"))
				Expect(cfg.LogLevel).To(Equal("info"))
				Expect(cfg.Environment).To(Equal("development"))
			})

			It("should return an error when the config file does not exist", func() {
				// Load config with non-existent file
				_, err := config.LoadConfig(filepath.Join(tempDir, "nonexistent.yaml"))
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when using default values", func() {
			It("should use default values when environment variables are not set", func() {
				// Load config without setting environment variables
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify default values
				Expect(cfg.ServerPort).To(Equal("8080"))
				Expect(cfg.LogLevel).To(Equal("info"))
				Expect(cfg.Environment).To(Equal("development"))
			})
		})

		Context("when constructing database connection strings", func() {
			It("should construct PostgresDSN correctly from components", func() {
				// Set environment variables for PostgreSQL components
				_ = os.Setenv("DB_HOST", "localhost")    // 오류 무시
				_ = os.Setenv("DB_PORT", "5432")         // 오류 무시
				_ = os.Setenv("DB_USER", "testuser")     // 오류 무시
				_ = os.Setenv("DB_PASSWORD", "testpass") // 오류 무시
				_ = os.Setenv("DB_NAME", "testdb")       // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify PostgresDSN
				expectedDSN := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
				Expect(cfg.PostgresDSN).To(Equal(expectedDSN))
			})

			It("should construct RedisAddr correctly from components", func() {
				// Set environment variables for Redis components
				_ = os.Setenv("REDIS_HOST", "localhost") // 오류 무시
				_ = os.Setenv("REDIS_PORT", "6379")      // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify RedisAddr
				expectedAddr := "localhost:6379"
				Expect(cfg.RedisAddr).To(Equal(expectedAddr))
			})
		})
	})

	Describe("ValidateConfig", func() {
		Context("when validating a config with missing required fields", func() {
			It("should return an error for missing ServerPort", func() {
				// Create a config with missing ServerPort
				cfg := &config.Config{
					LogLevel:    "info",
					Environment: "development",
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("SERVER_PORT"))
			})

			It("should return an error for missing Environment", func() {
				// Create a config with missing Environment
				cfg := &config.Config{
					ServerPort: "8080",
					LogLevel:   "info",
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("ENVIRONMENT"))
			})
		})

		Context("when validating in production environment", func() {
			It("should fail validation when required fields are missing in production", func() {
				// Create a production config with missing required fields
				cfg := &config.Config{
					ServerPort:  "8080",
					LogLevel:    "info",
					Environment: "production",
					// Missing PostgresDSN and RedisAddr
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("POSTGRES_DSN"))
				Expect(err.Error()).To(ContainSubstring("REDIS_ADDR"))
			})

			It("should pass validation when all required production fields are set", func() {
				// Create a complete production config
				cfg := &config.Config{
					ServerPort:        "8080",
					LogLevel:          "info",
					Environment:       "production",
					PostgresDSN:       "postgres://user:pass@host:5432/db",
					RedisAddr:         "host:6379",
					FirebaseProjectID: "test-project",
					JwtSecretKey:      "test-secret",
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetConfig and SetConfig", func() {
		It("should return the expected config instance", func() {
			// Create a config
			cfg := &config.Config{
				ServerPort:  "8080",
				LogLevel:    "info",
				Environment: "development",
			}

			// Set the config
			config.SetConfig(cfg)

			// Get the config
			retrievedCfg := config.GetConfig()

			// Verify it's the same config
			Expect(retrievedCfg).To(Equal(cfg))
		})

		It("should panic when config is not loaded", func() {
			// Reset config
			config.SetConfig(nil)

			// Expect panic when getting config
			Expect(func() {
				config.GetConfig()
			}).To(Panic())
		})
	})
})
