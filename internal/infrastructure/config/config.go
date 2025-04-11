package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration settings for the application
type Config struct {
	// Server settings
	ServerPort string `mapstructure:"SERVER_PORT"`
	LogLevel   string `mapstructure:"LOG_LEVEL"`

	// Database settings
	PostgresDSN string `mapstructure:"POSTGRES_DSN"`

	// Redis settings
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	// Firebase settings
	FirebaseProjectID       string `mapstructure:"FIREBASE_PROJECT_ID"`
	FirebaseCredentialsJSON string `mapstructure:"FIREBASE_CREDENTIALS_JSON"`

	// JWT settings
	JwtSecretKey string `mapstructure:"JWT_SECRET_KEY"`

	// Application settings
	Environment string `mapstructure:"ENVIRONMENT"`
}

// LoadConfig loads the application configuration from environment variables
// and optionally from a configuration file if specified
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// Set up Viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set default values
	setDefaults()

	// If a config file is specified, load it
	if configPath != "" {
		// Check if the file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", configPath)
		}

		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal the config into our struct
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate the configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for configuration parameters
func setDefaults() {
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("ENVIRONMENT", "development")

	// Construct PostgresDSN from individual components if not provided directly
	if os.Getenv("POSTGRES_DSN") == "" && os.Getenv("DB_HOST") != "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432" // Default PostgreSQL port
		}
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		if host != "" && user != "" && password != "" && dbname != "" {
			dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
				user, password, host, port, dbname)
			viper.SetDefault("POSTGRES_DSN", dsn)
		}
	}

	// Construct RedisAddr from individual components if not provided directly
	if os.Getenv("REDIS_ADDR") == "" && os.Getenv("REDIS_HOST") != "" {
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if port == "" {
			port = "6379" // Default Redis port
		}

		if host != "" {
			addr := fmt.Sprintf("%s:%s", host, port)
			viper.SetDefault("REDIS_ADDR", addr)
		}
	}

	// Map PORT to SERVER_PORT for compatibility with Cloud Run and other platforms
	if os.Getenv("PORT") != "" && os.Getenv("SERVER_PORT") == "" {
		viper.SetDefault("SERVER_PORT", os.Getenv("PORT"))
	}
}

// ValidateConfig validates the configuration values
func ValidateConfig(config *Config) error {
	var missingFields []string

	// Check required fields
	if config.ServerPort == "" {
		missingFields = append(missingFields, "SERVER_PORT")
	}

	if config.Environment == "" {
		missingFields = append(missingFields, "ENVIRONMENT")
	}

	// In production, we should require more fields
	if strings.ToLower(config.Environment) == "production" {
		if config.PostgresDSN == "" {
			missingFields = append(missingFields, "POSTGRES_DSN")
		}

		if config.RedisAddr == "" {
			missingFields = append(missingFields, "REDIS_ADDR")
		}

		if config.FirebaseProjectID == "" {
			missingFields = append(missingFields, "FIREBASE_PROJECT_ID")
		}

		if config.JwtSecretKey == "" {
			missingFields = append(missingFields, "JWT_SECRET_KEY")
		}
	}

	if len(missingFields) > 0 {
		return errors.New("missing required configuration fields: " + strings.Join(missingFields, ", "))
	}

	return nil
}

// GetConfig is a singleton function that returns the application configuration
// It should be called after LoadConfig has been called once
var configInstance *Config

// GetConfig returns the current configuration instance
// It panics if the configuration has not been loaded yet
func GetConfig() *Config {
	if configInstance == nil {
		panic("configuration not loaded, call LoadConfig first")
	}
	return configInstance
}

// SetConfig sets the global configuration instance
// This is primarily used for testing and should not be called directly in application code
func SetConfig(config *Config) {
	configInstance = config
}

// ResetViper resets the Viper instance to its default state
// This is primarily used for testing to ensure a clean state between tests
func ResetViper() {
	viper.Reset()
}
