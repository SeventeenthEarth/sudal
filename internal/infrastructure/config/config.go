package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Environment represents the application environment
type Environment string

const (
	// DevEnvironment represents the development environment
	DevEnvironment Environment = "dev"
	// CanaryEnvironment represents the canary environment
	CanaryEnvironment Environment = "canary"
	// ProductionEnvironment represents the production environment
	ProductionEnvironment Environment = "production"
	// TestEnvironment represents the test environment (for testing only)
	TestEnvironment Environment = "test"
)

// DBConfig holds database-specific configuration
type DBConfig struct {
	// Connection parameters
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	Name     string `mapstructure:"DB_NAME"`
	SSLMode  string `mapstructure:"DB_SSLMODE"`

	// Full connection string (DSN)
	DSN string `mapstructure:"POSTGRES_DSN"`
}

// Config holds all configuration settings for the application
type Config struct {
	// Server settings
	ServerPort string `mapstructure:"SERVER_PORT"`
	LogLevel   string `mapstructure:"LOG_LEVEL"`

	// Database settings
	DB DBConfig

	// Redis settings
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`

	// Firebase settings
	FirebaseProjectID       string `mapstructure:"FIREBASE_PROJECT_ID"`
	FirebaseCredentialsJSON string `mapstructure:"FIREBASE_CREDENTIALS_JSON"`

	// JWT settings
	JwtSecretKey string `mapstructure:"JWT_SECRET_KEY"`

	// Application settings
	AppEnv      string `mapstructure:"APP_ENV"`
	Environment string `mapstructure:"ENVIRONMENT"` // Legacy field, use AppEnv instead
}

// LoadEnvFiles loads environment variables from .env files
// It first tries to load environment-specific .env file (e.g., .env.production)
// and then falls back to the default .env file
func LoadEnvFiles() error {
	// Determine the environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		// Default to dev environment if not specified
		appEnv = string(DevEnvironment)
	}

	// Load environment-specific .env file if it exists
	if appEnv != string(DevEnvironment) {
		envFile := fmt.Sprintf(".env.%s", appEnv)
		if _, err := os.Stat(envFile); err == nil {
			// Load environment-specific .env file
			if err := loadEnvFile(envFile); err != nil {
				return fmt.Errorf("failed to load environment file %s: %w", envFile, err)
			}
		}
	}

	// Load default .env file if it exists (lowest priority)
	if _, err := os.Stat(".env"); err == nil {
		if err := loadEnvFile(".env"); err != nil {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
	}

	return nil
}

// LoadConfig loads the application configuration from environment variables
// and optionally from a configuration file if specified
func LoadConfig(configPath string) (*Config, error) {
	var config Config

	// Load environment variables from .env files
	if err := LoadEnvFiles(); err != nil {
		// Log the error but continue, as we might have environment variables set directly
		fmt.Printf("Warning: %v\n", err)
	}

	// Set up Viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Enable Viper to handle nested keys
	viper.SetEnvPrefix("")
	viper.SetTypeByDefaultValue(true)

	// Determine the environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		// Default to dev environment if not specified
		appEnv = string(DevEnvironment)
	}

	// Set default values
	setDefaults()

	// If a config file is specified, load it
	if configPath != "" {
		// Check if the file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", configPath)
		}

		viper.SetConfigFile(configPath)

		// Set Viper to handle YAML files correctly
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Unmarshal the config into our struct
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set the AppEnv field
	if config.AppEnv == "" {
		config.AppEnv = appEnv
	}

	// For backward compatibility, set Environment if it's not set
	if config.Environment == "" {
		config.Environment = config.AppEnv
	}

	// Handle nested DB configuration from YAML
	if viper.IsSet("db.dsn") {
		config.DB.DSN = viper.GetString("db.dsn")
	}
	if viper.IsSet("db.host") {
		config.DB.Host = viper.GetString("db.host")
	}
	if viper.IsSet("db.port") {
		config.DB.Port = viper.GetString("db.port")
	}
	if viper.IsSet("db.user") {
		config.DB.User = viper.GetString("db.user")
	}
	if viper.IsSet("db.password") {
		config.DB.Password = viper.GetString("db.password")
	}
	if viper.IsSet("db.name") {
		config.DB.Name = viper.GetString("db.name")
	}
	if viper.IsSet("db.sslmode") {
		config.DB.SSLMode = viper.GetString("db.sslmode")
	}

	// Process database configuration
	processDatabaseConfig(&config)

	// Validate the configuration
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(filename string) error {
	// We're using godotenv.Load() directly here to load the environment variables
	return godotenv.Load(filename)
}

// processDatabaseConfig processes database configuration
// It ensures that either DSN is set directly or constructed from individual components
func processDatabaseConfig(config *Config) {
	// Check if POSTGRES_DSN is set directly (legacy support)
	if config.DB.DSN == "" {
		postgresEnvDSN := os.Getenv("POSTGRES_DSN")
		if postgresEnvDSN != "" {
			config.DB.DSN = postgresEnvDSN
		}
	}

	// Check for individual DB components from environment variables if not already set
	if config.DB.Host == "" {
		config.DB.Host = os.Getenv("DB_HOST")
	}
	if config.DB.Port == "" {
		config.DB.Port = os.Getenv("DB_PORT")
		if config.DB.Port == "" {
			config.DB.Port = "5432" // Default PostgreSQL port
		}
	}
	if config.DB.User == "" {
		config.DB.User = os.Getenv("DB_USER")
	}
	if config.DB.Password == "" {
		config.DB.Password = os.Getenv("DB_PASSWORD")
	}
	if config.DB.Name == "" {
		config.DB.Name = os.Getenv("DB_NAME")
	}
	if config.DB.SSLMode == "" {
		config.DB.SSLMode = os.Getenv("DB_SSLMODE")
		if config.DB.SSLMode == "" {
			config.DB.SSLMode = "disable" // Default to disable SSL
		}
	}

	// If DSN is not set but individual components are, construct the DSN
	if config.DB.DSN == "" && config.DB.Host != "" {
		if config.DB.Host != "" && config.DB.User != "" && config.DB.Password != "" && config.DB.Name != "" {
			config.DB.DSN = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name, config.DB.SSLMode)
		}
	}
}

// setDefaults sets default values for configuration parameters
func setDefaults() {
	// Server defaults
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("LOG_LEVEL", "info")

	// Environment defaults
	viper.SetDefault("APP_ENV", string(DevEnvironment))
	viper.SetDefault("ENVIRONMENT", string(DevEnvironment))

	// Database defaults
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_SSLMODE", "disable")

	// Redis defaults
	viper.SetDefault("REDIS_PORT", "6379")

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

	// Validate environment
	if config.AppEnv == "" {
		missingFields = append(missingFields, "APP_ENV")
	} else {
		// Check if the environment is valid
		validEnvs := map[string]bool{
			string(DevEnvironment):        true,
			string(CanaryEnvironment):     true,
			string(ProductionEnvironment): true,
			string(TestEnvironment):       true,
		}

		if !validEnvs[config.AppEnv] {
			return fmt.Errorf("invalid APP_ENV value: %s (must be one of: dev, canary, production, test)", config.AppEnv)
		}
	}

	// Validate database configuration based on environment
	isProduction := strings.ToLower(config.AppEnv) == string(ProductionEnvironment) ||
		strings.ToLower(config.Environment) == string(ProductionEnvironment)

	if isProduction {
		// In production, database connection is required
		if config.DB.DSN == "" {
			// Check if we can construct DSN from individual components
			if config.DB.Host == "" {
				missingFields = append(missingFields, "DB_HOST or POSTGRES_DSN")
			}
			if config.DB.User == "" {
				missingFields = append(missingFields, "DB_USER")
			}
			if config.DB.Password == "" {
				missingFields = append(missingFields, "DB_PASSWORD")
			}
			if config.DB.Name == "" {
				missingFields = append(missingFields, "DB_NAME")
			}
		}

		// Other production requirements
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

	// Re-initialize Viper with default settings
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("")
	viper.SetTypeByDefaultValue(true)
}
