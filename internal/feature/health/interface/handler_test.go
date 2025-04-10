package interfaces_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	interfaces "github.com/seventeenthearth/sudal/internal/feature/health/interface"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

func TestNewHandler(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	// Act
	handler := interfaces.NewHandler(mockService)

	// Assert
	if handler == nil {
		t.Fatal("Expected handler to not be nil")
	}
}

func TestHandler_Ping(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := domain.NewStatus("test-ok")
		mockService := mocks.NewMockService(ctrl)
		mockService.EXPECT().Ping(gomock.Any()).Return(expectedStatus, nil)

		handler := interfaces.NewHandler(mockService)
		req := httptest.NewRequest("GET", "/ping", nil)
		recorder := httptest.NewRecorder()

		// Act
		handler.Ping(recorder, req)

		// Assert
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		contentType := recorder.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
		}

		var status domain.Status
		err := json.NewDecoder(recorder.Body).Decode(&status)
		if err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
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

		expectedError := fmt.Errorf("service error")
		mockService := mocks.NewMockService(ctrl)
		mockService.EXPECT().Ping(gomock.Any()).Return(nil, expectedError)

		handler := interfaces.NewHandler(mockService)
		req := httptest.NewRequest("GET", "/ping", nil)
		recorder := httptest.NewRecorder()

		// Act
		handler.Ping(recorder, req)

		// Assert
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder.Code)
		}
	})
}

func TestHandler_Health(t *testing.T) {
	// Success case
	t.Run("Success", func(t *testing.T) {
		// Arrange
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		expectedStatus := domain.NewStatus("test-healthy")
		mockService := mocks.NewMockService(ctrl)
		mockService.EXPECT().Check(gomock.Any()).Return(expectedStatus, nil)

		handler := interfaces.NewHandler(mockService)
		req := httptest.NewRequest("GET", "/healthz", nil)
		recorder := httptest.NewRecorder()

		// Act
		handler.Health(recorder, req)

		// Assert
		if recorder.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
		}

		contentType := recorder.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
		}

		var status domain.Status
		err := json.NewDecoder(recorder.Body).Decode(&status)
		if err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
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

		expectedError := fmt.Errorf("service error")
		mockService := mocks.NewMockService(ctrl)
		mockService.EXPECT().Check(gomock.Any()).Return(nil, expectedError)

		handler := interfaces.NewHandler(mockService)
		req := httptest.NewRequest("GET", "/healthz", nil)
		recorder := httptest.NewRecorder()

		// Act
		handler.Health(recorder, req)

		// Assert
		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, recorder.Code)
		}
	})
}

func TestHandler_RegisterRoutes(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	handler := interfaces.NewHandler(mockService)
	mux := http.NewServeMux()

	// Act - This should not panic
	handler.RegisterRoutes(mux)

	// Assert - We can't easily test the routes are registered correctly without making HTTP requests
	// This is more of a smoke test to ensure the method doesn't panic
}
