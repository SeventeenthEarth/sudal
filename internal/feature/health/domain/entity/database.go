package entity

import "time"

// DatabaseStatus represents the health status of the database connection
type DatabaseStatus struct {
	Status  string           `json:"status"`
	Message string           `json:"message"`
	Stats   *ConnectionStats `json:"stats,omitempty"`
}

// ConnectionStats represents database connection pool statistics
type ConnectionStats struct {
	MaxOpenConnections int           `json:"max_open_connections"`
	OpenConnections    int           `json:"open_connections"`
	InUse              int           `json:"in_use"`
	Idle               int           `json:"idle"`
	WaitCount          int64         `json:"wait_count"`
	WaitDuration       time.Duration `json:"wait_duration"`
	MaxIdleClosed      int64         `json:"max_idle_closed"`
	MaxLifetimeClosed  int64         `json:"max_lifetime_closed"`
}

// NewDatabaseStatus creates a new DatabaseStatus with the given parameters
func NewDatabaseStatus(status, message string, stats *ConnectionStats) *DatabaseStatus {
	return &DatabaseStatus{
		Status:  status,
		Message: message,
		Stats:   stats,
	}
}

// HealthyDatabaseStatus returns a DatabaseStatus indicating the database is healthy
func HealthyDatabaseStatus(message string, stats *ConnectionStats) *DatabaseStatus {
	return NewDatabaseStatus("healthy", message, stats)
}

// UnhealthyDatabaseStatus returns a DatabaseStatus indicating the database is unhealthy
func UnhealthyDatabaseStatus(message string) *DatabaseStatus {
	return NewDatabaseStatus("unhealthy", message, nil)
}

// DetailedHealthStatus represents comprehensive health status including database information
type DetailedHealthStatus struct {
	Status    string          `json:"status"`
	Message   string          `json:"message"`
	Timestamp string          `json:"timestamp"`
	Database  *DatabaseStatus `json:"database,omitempty"`
}

// NewDetailedHealthStatus creates a new DetailedHealthStatus
func NewDetailedHealthStatus(status, message, timestamp string, database *DatabaseStatus) *DetailedHealthStatus {
	return &DetailedHealthStatus{
		Status:    status,
		Message:   message,
		Timestamp: timestamp,
		Database:  database,
	}
}
