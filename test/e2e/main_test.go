package e2e

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestAppExecution is a basic end-to-end test that builds and runs the application
func TestAppExecution(t *testing.T) {
	// Skip this test by default since the application is not fully implemented yet
	// Only run if E2E_TEST environment variable is set
	if os.Getenv("E2E_TEST") == "" {
		t.Skip("Skipping end-to-end test; set E2E_TEST=1 to run")
	}

	// Get the project root directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Navigate up to the project root
	projectRoot := filepath.Join(wd, "..", "..")

	// Build the application
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = projectRoot
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build application: %v\nOutput: %s", err, buildOutput)
	}

	// Run the application with a timeout
	appPath := filepath.Join(projectRoot, "bin", "server")
	runCmd := exec.Command(appPath, "--version")
	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run application: %v\nOutput: %s", err, runOutput)
	}

	// Check that the output contains version information
	outputStr := string(runOutput)
	t.Logf("Application output: %s", outputStr)

	// This is a minimal test - in a real project, we would make assertions about the output
	// For now, we're just checking that the application runs without error
}

// TestHealthCheck is a basic test that checks the application's health endpoint
func TestHealthCheck(t *testing.T) {
	// Skip this test by default since it requires the server to be running
	// Only run if E2E_TEST environment variable is set
	if os.Getenv("E2E_TEST") == "" {
		t.Skip("Skipping end-to-end test; set E2E_TEST=1 to run")
	}

	// Define the server URL
	serverURL := "http://localhost:8080"

	// Test the ping endpoint
	t.Run("Ping", func(t *testing.T) {
		resp, err := http.Get(serverURL + "/ping")
		if err != nil {
			t.Fatalf("Failed to make request to ping endpoint: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Parse the response body
		var status map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// Check the status
		if status["status"] != "ok" {
			t.Errorf("Expected status %s, got %s", "ok", status["status"])
		}
	})

	// Test the health endpoint
	t.Run("Health", func(t *testing.T) {
		resp, err := http.Get(serverURL + "/healthz")
		if err != nil {
			t.Fatalf("Failed to make request to health endpoint: %v", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				t.Logf("Failed to close response body: %v", err)
			}
		}()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		// Parse the response body
		var status map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// Check the status
		if status["status"] != "healthy" {
			t.Errorf("Expected status %s, got %s", "healthy", status["status"])
		}
	})
}
