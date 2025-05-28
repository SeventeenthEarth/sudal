package mocks

import (
	"fmt"
	"time"

	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
	"github.com/seventeenthearth/sudal/internal/mocks"
	"go.uber.org/mock/gomock"
)

// SetHealthyStatus configures the mock repository to return healthy status
func SetHealthyStatus(mockRepo *mocks.MockHealthRepository) {
	// Set up expectations for GetStatus
	mockRepo.EXPECT().GetStatus(gomock.Any()).Return(domain.HealthyStatus(), nil).AnyTimes()

	// Set up expectations for GetDatabaseStatus
	healthyDBStatus := &domain.DatabaseStatus{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Stats: &domain.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    5,
			InUse:              2,
			Idle:               3,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		},
	}
	mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(healthyDBStatus, nil).AnyTimes()
}

// SetUnhealthyStatus configures the mock repository to return unhealthy status with error
func SetUnhealthyStatus(mockRepo *mocks.MockHealthRepository, err error) {
	// Set up expectations for GetStatus to return error
	mockRepo.EXPECT().GetStatus(gomock.Any()).Return(nil, err).AnyTimes()

	// Set up expectations for GetDatabaseStatus to return unhealthy status with error
	// This matches the actual repository behavior which returns both DatabaseStatus and error
	unhealthyDBStatus := domain.UnhealthyDatabaseStatus(fmt.Sprintf("Database health check failed: %v", err))
	mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(unhealthyDBStatus, err).AnyTimes()
}

// SetCustomStatus configures the mock repository to return a custom status
func SetCustomStatus(mockRepo *mocks.MockHealthRepository, status *domain.Status) {
	// Set up expectations for GetStatus
	mockRepo.EXPECT().GetStatus(gomock.Any()).Return(status, nil).AnyTimes()

	// Set up expectations for GetDatabaseStatus with default healthy database
	healthyDBStatus := &domain.DatabaseStatus{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Stats: &domain.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    5,
			InUse:              2,
			Idle:               3,
			WaitCount:          0,
			WaitDuration:       0,
			MaxIdleClosed:      0,
			MaxLifetimeClosed:  0,
		},
	}
	mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(healthyDBStatus, nil).AnyTimes()
}

// SetDegradedStatus configures the mock repository to return degraded status
func SetDegradedStatus(mockRepo *mocks.MockHealthRepository) {
	// Set up expectations for GetStatus
	degradedStatus := domain.NewStatus("degraded")
	mockRepo.EXPECT().GetStatus(gomock.Any()).Return(degradedStatus, nil).AnyTimes()

	// Set up expectations for GetDatabaseStatus
	// Configure degraded state with all connections in use (high connection usage)
	degradedDBStatus := &domain.DatabaseStatus{
		Status:  "degraded",
		Message: "Database connection is degraded",
		Stats: &domain.ConnectionStats{
			MaxOpenConnections: 25,
			OpenConnections:    25, // All connections in use
			InUse:              25, // All connections in use
			Idle:               0,  // No idle connections
			WaitCount:          10, // Requests waiting
			WaitDuration:       500 * time.Millisecond,
			MaxIdleClosed:      2,
			MaxLifetimeClosed:  1,
		},
	}
	mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(degradedDBStatus, nil).AnyTimes()
}

// SetDatabaseStatus configures the mock repository to return a custom database status
func SetDatabaseStatus(mockRepo *mocks.MockHealthRepository, dbStatus *domain.DatabaseStatus) {
	// Set up expectations for GetStatus with default healthy status
	mockRepo.EXPECT().GetStatus(gomock.Any()).Return(domain.HealthyStatus(), nil).AnyTimes()

	// Set up expectations for GetDatabaseStatus
	mockRepo.EXPECT().GetDatabaseStatus(gomock.Any()).Return(dbStatus, nil).AnyTimes()
}

// ConfigureForHealthyState is an alias for SetHealthyStatus for consistency with test_helpers.go
func ConfigureForHealthyState(mockRepo *mocks.MockHealthRepository) {
	SetHealthyStatus(mockRepo)
}

// ConfigureForUnhealthyState is an alias for SetUnhealthyStatus for consistency with test_helpers.go
func ConfigureForUnhealthyState(mockRepo *mocks.MockHealthRepository, err error) {
	SetUnhealthyStatus(mockRepo, err)
}

// ConfigureForDegradedState is an alias for SetDegradedStatus for consistency with test_helpers.go
func ConfigureForDegradedState(mockRepo *mocks.MockHealthRepository) {
	SetDegradedStatus(mockRepo)
}
