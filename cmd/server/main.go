package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
	"go.uber.org/zap"
)

func main() {
	// Initialize application
	cfg := initializeApplication()

	// Verify external dependencies
	verifyDependencies(cfg)

	// Start the server
	startServer(cfg)

	// Cleanup
	cleanup()
}

// initializeApplication handles configuration loading and logger initialization
func initializeApplication() *config.Config {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Set the global config instance
	config.SetConfig(cfg)

	// Initialize the logger with the configured log level
	logLevel := log.ParseLevel(cfg.LogLevel)
	log.Init(logLevel)

	// Log configuration details
	log.Info("Application starting",
		zap.String("app_env", cfg.AppEnv),
		zap.String("server_port", cfg.ServerPort),
		zap.String("log_level", string(logLevel)),
	)

	return cfg
}

// verifyDependencies checks connectivity to external dependencies
func verifyDependencies(cfg *config.Config) {
	verifyDatabaseConnectivity(cfg)
	verifyRedisConnectivity(cfg)
}

// verifyDatabaseConnectivity verifies database connection
func verifyDatabaseConnectivity(cfg *config.Config) {
	log.Info("Verifying database connectivity...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.VerifyDatabaseConnectivity(ctx, cfg); err != nil {
		log.Error("Database connectivity verification failed", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Database connectivity verified successfully")
}

// verifyRedisConnectivity verifies Redis connection
func verifyRedisConnectivity(cfg *config.Config) {
	log.Info("Verifying Redis connectivity...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := database.VerifyRedisConnectivity(ctx, cfg); err != nil {
		log.Error("Redis connectivity verification failed", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Redis connectivity verified successfully")
}

// startServer creates and starts the HTTP server
func startServer(cfg *config.Config) {
	srv := server.NewServer(cfg.ServerPort)
	if err := srv.Start(); err != nil {
		log.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}
}

// cleanup performs cleanup operations before application exit
func cleanup() {
	// Ensure logs are flushed before exiting
	if err := log.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to sync logger: %v\n", err)
	}
}
