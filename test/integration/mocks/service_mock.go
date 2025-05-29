package mocks

import (
	"context"
	"fmt"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

// MockService is a mock implementation of the health.Service interface
type MockService struct {
	PingFunc          func(ctx context.Context) (*entity.HealthStatus, error)
	CheckFunc         func(ctx context.Context) (*entity.HealthStatus, error)
	CheckDatabaseFunc func(ctx context.Context) (*entity.DatabaseStatus, error)
}

// Ping calls the mocked PingFunc
func (m *MockService) Ping(ctx context.Context) (*entity.HealthStatus, error) {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return entity.OkStatus(), nil
}

// Check calls the mocked CheckFunc
func (m *MockService) Check(ctx context.Context) (*entity.HealthStatus, error) {
	if m.CheckFunc != nil {
		return m.CheckFunc(ctx)
	}
	return entity.HealthyStatus(), nil
}

// CheckDatabase calls the mocked CheckDatabaseFunc
func (m *MockService) CheckDatabase(ctx context.Context) (*entity.DatabaseStatus, error) {
	if m.CheckDatabaseFunc != nil {
		return m.CheckDatabaseFunc(ctx)
	}
	// Return a default healthy database status for tests
	stats := &entity.ConnectionStats{
		MaxOpenConnections: 25,
		OpenConnections:    1,
		InUse:              0,
		Idle:               1,
		WaitCount:          0,
		WaitDuration:       0,
		MaxIdleClosed:      0,
		MaxLifetimeClosed:  0,
	}
	return entity.HealthyDatabaseStatus("Mock database connection is healthy", stats), nil
}

// NewMockServiceWithError returns a mock service that returns an error for all methods
func NewMockServiceWithError() *MockService {
	return &MockService{
		PingFunc: func(ctx context.Context) (*entity.HealthStatus, error) {
			return nil, fmt.Errorf("mock ping error")
		},
		CheckFunc: func(ctx context.Context) (*entity.HealthStatus, error) {
			return nil, fmt.Errorf("mock check error")
		},
		CheckDatabaseFunc: func(ctx context.Context) (*entity.DatabaseStatus, error) {
			return nil, fmt.Errorf("mock database check error")
		},
	}
}

// NewMockService returns a mock service with default implementations
func NewMockService() *MockService {
	return &MockService{}
}
