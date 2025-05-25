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
			"SERVER_PORT":                   os.Getenv("SERVER_PORT"),
			"LOG_LEVEL":                     os.Getenv("LOG_LEVEL"),
			"APP_ENV":                       os.Getenv("APP_ENV"),
			"ENVIRONMENT":                   os.Getenv("ENVIRONMENT"),
			"POSTGRES_DSN":                  os.Getenv("POSTGRES_DSN"),
			"DB_HOST":                       os.Getenv("DB_HOST"),
			"DB_PORT":                       os.Getenv("DB_PORT"),
			"DB_USER":                       os.Getenv("DB_USER"),
			"DB_PASSWORD":                   os.Getenv("DB_PASSWORD"),
			"DB_NAME":                       os.Getenv("DB_NAME"),
			"DB_SSLMODE":                    os.Getenv("DB_SSLMODE"),
			"DB_MAX_OPEN_CONNS":             os.Getenv("DB_MAX_OPEN_CONNS"),
			"DB_MAX_IDLE_CONNS":             os.Getenv("DB_MAX_IDLE_CONNS"),
			"DB_CONN_MAX_LIFETIME_SECONDS":  os.Getenv("DB_CONN_MAX_LIFETIME_SECONDS"),
			"DB_CONN_MAX_IDLE_TIME_SECONDS": os.Getenv("DB_CONN_MAX_IDLE_TIME_SECONDS"),
			"DB_CONNECT_TIMEOUT_SECONDS":    os.Getenv("DB_CONNECT_TIMEOUT_SECONDS"),
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
				_ = os.Setenv("APP_ENV", "test")     // 오류 무시
				_ = os.Setenv("ENVIRONMENT", "test") // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify config values
				Expect(cfg.ServerPort).To(Equal("9090"))
				Expect(cfg.LogLevel).To(Equal("debug"))
				Expect(cfg.AppEnv).To(Equal("test"))
				Expect(cfg.Environment).To(Equal("test"))
			})
		})

		Context("when loading from a config file", func() {
			It("should load configuration from a file", func() {
				// Create a config file
				configContent := `
server_port: "8888"
log_level: "info"
app_env: "dev"
environment: "dev"
db:
  dsn: "postgres://user:pass@host:5432/db?sslmode=disable"
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
				Expect(cfg.AppEnv).To(Equal("dev"))
				Expect(cfg.Environment).To(Equal("dev"))
				Expect(cfg.DB.DSN).To(Equal("postgres://user:pass@host:5432/db?sslmode=disable"))
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
				Expect(cfg.AppEnv).To(Equal("dev"))
				Expect(cfg.Environment).To(Equal("dev"))
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
				_ = os.Setenv("DB_SSLMODE", "disable")   // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify PostgresDSN
				expectedDSN := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
				Expect(cfg.DB.DSN).To(Equal(expectedDSN))
			})

			It("should load database connection pool configuration correctly", func() {
				// Set environment variables for database pool configuration
				_ = os.Setenv("DB_HOST", "localhost")                 // 오류 무시
				_ = os.Setenv("DB_PORT", "5432")                      // 오류 무시
				_ = os.Setenv("DB_USER", "testuser")                  // 오류 무시
				_ = os.Setenv("DB_PASSWORD", "testpass")              // 오류 무시
				_ = os.Setenv("DB_NAME", "testdb")                    // 오류 무시
				_ = os.Setenv("DB_SSLMODE", "disable")                // 오류 무시
				_ = os.Setenv("DB_MAX_OPEN_CONNS", "50")              // 오류 무시
				_ = os.Setenv("DB_MAX_IDLE_CONNS", "10")              // 오류 무시
				_ = os.Setenv("DB_CONN_MAX_LIFETIME_SECONDS", "7200") // 오류 무시
				_ = os.Setenv("DB_CONN_MAX_IDLE_TIME_SECONDS", "600") // 오류 무시
				_ = os.Setenv("DB_CONNECT_TIMEOUT_SECONDS", "60")     // 오류 무시

				// Load config
				cfg, err := config.LoadConfig("")
				Expect(err).NotTo(HaveOccurred())
				Expect(cfg).NotTo(BeNil())

				// Verify database pool configuration
				Expect(cfg.DB.MaxOpenConns).To(Equal(50))
				Expect(cfg.DB.MaxIdleConns).To(Equal(10))
				Expect(cfg.DB.ConnMaxLifetimeSeconds).To(Equal(7200))
				Expect(cfg.DB.ConnMaxIdleTimeSeconds).To(Equal(600))
				Expect(cfg.DB.ConnectTimeoutSeconds).To(Equal(60))
			})
		})
	})

	Describe("ValidateConfig", func() {
		Context("when validating a config with missing required fields", func() {
			It("should return an error for missing ServerPort", func() {
				// Create a config with missing ServerPort
				cfg := &config.Config{
					LogLevel:    "info",
					AppEnv:      "dev",
					Environment: "dev",
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("SERVER_PORT"))
			})

			It("should return an error for missing AppEnv", func() {
				// Create a config with missing AppEnv
				cfg := &config.Config{
					ServerPort: "8080",
					LogLevel:   "info",
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("APP_ENV"))
			})
		})

		Context("when validating in production environment", func() {
			It("should fail validation when required database fields are missing in production", func() {
				// Create a production config with missing database configuration
				cfg := &config.Config{
					ServerPort:  "8080",
					LogLevel:    "info",
					AppEnv:      "production",
					Environment: "production",
					// Missing DB.DSN
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("DB_HOST or POSTGRES_DSN"))
			})

			It("should pass validation when database configuration is provided in production", func() {
				// Create a production config with database configuration
				cfg := &config.Config{
					ServerPort:  "8080",
					LogLevel:    "info",
					AppEnv:      "production",
					Environment: "production",
					DB: config.DBConfig{
						DSN:                    "postgres://user:pass@host:5432/db",
						MaxOpenConns:           25,
						MaxIdleConns:           5,
						ConnMaxLifetimeSeconds: 3600,
						ConnMaxIdleTimeSeconds: 300,
						ConnectTimeoutSeconds:  30,
					},
				}

				// Validate config
				err := config.ValidateConfig(cfg)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should pass validation when database components are provided in production", func() {
				// Create a production config with database components
				cfg := &config.Config{
					ServerPort:  "8080",
					LogLevel:    "info",
					AppEnv:      "production",
					Environment: "production",
					DB: config.DBConfig{
						Host:                   "localhost",
						Port:                   "5432",
						User:                   "user",
						Password:               "password",
						Name:                   "testdb",
						SSLMode:                "require",
						MaxOpenConns:           25,
						MaxIdleConns:           5,
						ConnMaxLifetimeSeconds: 3600,
						ConnMaxIdleTimeSeconds: 300,
						ConnectTimeoutSeconds:  30,
						DSN:                    "postgres://user:password@localhost:5432/testdb?sslmode=require",
					},
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
				AppEnv:      "dev",
				Environment: "dev",
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
