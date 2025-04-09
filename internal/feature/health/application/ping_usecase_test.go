package application_test

import (
	"context"
	"testing"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
)

func TestNewPingUseCase(t *testing.T) {
	// Act
	useCase := application.NewPingUseCase()

	// Assert
	if useCase == nil {
		t.Fatal("Expected use case to not be nil")
	}
}

func TestPingUseCase_Execute(t *testing.T) {
	// Arrange
	useCase := application.NewPingUseCase()
	ctx := context.Background()

	// Act
	status, err := useCase.Execute(ctx)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if status == nil {
		t.Fatal("Expected status to not be nil")
	}

	if status.Status != "ok" {
		t.Errorf("Expected status to be 'ok', got %s", status.Status)
	}
}
