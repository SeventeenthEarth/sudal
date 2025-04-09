package health_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	healthApp "github.com/seventeenthearth/sudal/internal/feature/health/application"
	healthData "github.com/seventeenthearth/sudal/internal/feature/health/data"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	healthHandler "github.com/seventeenthearth/sudal/internal/feature/health/interface"
)

func TestPingEndpoint(t *testing.T) {
	// Create a new health repository
	repo := healthData.NewRepository()

	// Create a new health service
	service := healthApp.NewService(repo)

	// Create a new health handler
	handler := healthHandler.NewHandler(service)

	// Create a new HTTP request
	req := httptest.NewRequest("GET", "/ping", nil)

	// Create a new recorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the ping handler
	handler.Ping(recorder, req)

	// Check the status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Check the content type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Parse the response body
	var status domain.Status
	err := json.NewDecoder(recorder.Body).Decode(&status)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Check the status
	if status.Status != "ok" {
		t.Errorf("Expected status %s, got %s", "ok", status.Status)
	}
}

func TestHealthEndpoint(t *testing.T) {
	// Create a new health repository
	repo := healthData.NewRepository()

	// Create a new health service
	service := healthApp.NewService(repo)

	// Create a new health handler
	handler := healthHandler.NewHandler(service)

	// Create a new HTTP request
	req := httptest.NewRequest("GET", "/healthz", nil)

	// Create a new recorder to capture the response
	recorder := httptest.NewRecorder()

	// Call the health handler
	handler.Health(recorder, req)

	// Check the status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Check the content type
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Parse the response body
	var status domain.Status
	err := json.NewDecoder(recorder.Body).Decode(&status)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Check the status
	if status.Status != "healthy" {
		t.Errorf("Expected status %s, got %s", "healthy", status.Status)
	}
}
