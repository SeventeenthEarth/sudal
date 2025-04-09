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

func TestNewService(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)

	// Act
	service := application.NewService(mockRepo)

	// Assert
	if service == nil {
		t.Fatal("Expected service to not be nil")
	}
}

func TestService_Ping(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mocks.NewMockRepository(ctrl)

		// We need to create a service with our mocks, but the NewService function
		// creates its own use cases. For this test, we'll just verify the behavior
		// of the real service with the real PingUseCase.
		service := application.NewService(mockRepo)
		ctx := context.Background()

		// Act
		status, err := service.Ping(ctx)

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
	})
}

func TestService_Check(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := domain.NewStatus("test-healthy")
		mockRepo := mocks.NewMockRepository(ctrl)
		mockRepo.EXPECT().GetStatus(gomock.Any()).Return(expectedStatus, nil)

		service := application.NewService(mockRepo)
		ctx := context.Background()

		// Act
		status, err := service.Check(ctx)

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

		service := application.NewService(mockRepo)
		ctx := context.Background()

		// Act
		status, err := service.Check(ctx)

		// Assert
		if err != expectedError {
			t.Fatalf("Expected error %v, got %v", expectedError, err)
		}

		if status != nil {
			t.Errorf("Expected status to be nil, got %v", status)
		}
	})
}
