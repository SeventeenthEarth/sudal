package mocks

import (
	"context"
	"fmt"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

// MockService is a mock implementation of the health.Service interface
type MockService struct {
	PingFunc  func(ctx context.Context) (*domain.Status, error)
	CheckFunc func(ctx context.Context) (*domain.Status, error)
}

// Ping calls the mocked PingFunc
func (m *MockService) Ping(ctx context.Context) (*domain.Status, error) {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return domain.OkStatus(), nil
}

// Check calls the mocked CheckFunc
func (m *MockService) Check(ctx context.Context) (*domain.Status, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx)
	}
	return domain.HealthyStatus(), nil
}

// NewMockServiceWithError returns a mock service that returns an error for all methods
func NewMockServiceWithError() *MockService {
	return &MockService{
		PingFunc: func(ctx context.Context) (*domain.Status, error) {
			return nil, fmt.Errorf("mock ping error")
		},
		CheckFunc: func(ctx context.Context) (*domain.Status, error) {
			return nil, fmt.Errorf("mock check error")
		},
	}
}

// NewMockService returns a mock service with default implementations
func NewMockService() *MockService {
	return &MockService{}
}
