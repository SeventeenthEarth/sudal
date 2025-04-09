package main

import (
	"fmt"
	"log"
	"os"

	"github.com/seventeenthearth/sudal/internal/infrastructure/server"
)

func main() {
	fmt.Println("Starting Sudal Server...")

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
		log.Printf("No PORT environment variable set, using default: %s", port)
	}

	// Create and start the server
	srv := server.NewServer(port)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
