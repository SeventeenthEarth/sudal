package domain_test

import (
	"testing"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

func TestNewStatus(t *testing.T) {
	// Arrange
	expectedStatus := "test-status"

	// Act
	status := domain.NewStatus(expectedStatus)

	// Assert
	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	if status.Status != expectedStatus {
		t.Errorf("Expected status to be %s, got %s", expectedStatus, status.Status)
	}
}

func TestHealthyStatus(t *testing.T) {
	// Act
	status := domain.HealthyStatus()

	// Assert
	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	if status.Status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %s", status.Status)
	}
}

func TestOkStatus(t *testing.T) {
	// Act
	status := domain.OkStatus()

	// Assert
	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	if status.Status != "ok" {
		t.Errorf("Expected status to be 'ok', got %s", status.Status)
	}
}
