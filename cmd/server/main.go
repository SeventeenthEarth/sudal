package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
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
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set the global config instance
	config.SetConfig(cfg)

	// Log configuration details
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Server port: %s", cfg.ServerPort)

	// Create and start the server
	srv := server.NewServer(cfg.ServerPort)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
