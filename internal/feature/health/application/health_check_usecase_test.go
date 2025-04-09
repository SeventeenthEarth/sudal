package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	"github.com/seventeenthearth/sudal/internal/mocks"
)

func TestNewHealthCheckUseCase(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	// Act
	useCase := application.NewHealthCheckUseCase(mockRepo)

	// Assert
	if useCase == nil {
		t.Fatal("Expected use case to not be nil")
	}
}

func TestHealthCheckUseCase_Execute(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := domain.NewStatus("test-healthy")
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().GetStatus(gomock.Any()).Return(expectedStatus, nil)

		useCase := application.NewHealthCheckUseCase(mockRepo)
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

		if status.Status != expectedStatus.Status {
			t.Errorf("Expected status to be '%s', got '%s'", expectedStatus.Status, status.Status)
		}
	})

	// Error case
	t.Run("Error", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedError := errors.New("repository error")
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, expectedError)

		useCase := application.NewHealthCheckUseCase(mockRepo)
		ctx := context.Background()

		// Act
		status, err := useCase.Execute(ctx)

		// Assert
		if err != expectedError {
			t.Fatalf("Expected error %v, got %v", expectedError, err)
		}

		if status != nil {
			t.Errorf("Expected status to be nil, got %v", status)
		}
	})
}
