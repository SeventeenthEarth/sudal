package di_test

import (
	"testing"

	"github.com/seventeenthearth/sudal/internal/infrastructure/di"
)

func TestInitializeHealthHandler(t *testing.T) {
	// Act
	handler := di.InitializeHealthHandler()

	// Assert
	if handler == nil {
		t.Fatal("Expected handler to not be nil")
	}
}
