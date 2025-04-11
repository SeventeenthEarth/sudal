package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Starting Sudal Server...")

	// Load .env file if it exists (for local development)
	_ = godotenv.Load()

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
		zap.String("environment", cfg.Environment),
		zap.String("server_port", cfg.ServerPort),
		zap.String("log_level", string(logLevel)),
	)

	// Create and start the server
	srv := server.NewServer(cfg.ServerPort)
	if err := srv.Start(); err != nil {
		log.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}

	// Ensure logs are flushed before exiting
	if err := log.Sync(); err != nil {
		fmt.Printf("Failed to sync logger: %v\n", err)
	}
}
