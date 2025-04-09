package data_test

import (
	"context"
	"testing"

	"github.com/seventeenthearth/sudal/internal/feature/health/data"
)

func TestNewRepository(t *testing.T) {
	// Act
	repo := data.NewRepository()

	// Assert
	if repo == nil {
		t.Fatal("Expected repository to not be nil")
	}
}

func TestRepository_GetStatus(t *testing.T) {
	// Arrange
	repo := data.NewRepository()
	ctx := context.Background()

	// Act
	status, err := repo.GetStatus(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	if status.Status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %s", status.Status)
	}
}
