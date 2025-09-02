package entity

// Canonical health status string constants
const (
	StatusUnknown   = "unknown"
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
	StatusDegraded  = "degraded"
	StatusOk        = "ok"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status string `json:"status"`
}

// NewHealthStatus creates a new HealthStatus with the given status string
func NewHealthStatus(status string) *HealthStatus {
	return &HealthStatus{
		Status: status,
	}
}

// UnknownStatus returns a HealthStatus indicating the service status is unknown
func UnknownStatus() *HealthStatus {
	return NewHealthStatus(StatusUnknown)
}

// HealthyStatus returns a HealthStatus indicating the service is healthy
func HealthyStatus() *HealthStatus {
	return NewHealthStatus(StatusHealthy)
}

// UnhealthyStatus returns a HealthStatus indicating the service is unhealthy
func UnhealthyStatus() *HealthStatus {
	return NewHealthStatus(StatusUnhealthy)
}

// DegradedStatus returns a HealthStatus indicating the service is degraded
func DegradedStatus() *HealthStatus {
	return NewHealthStatus(StatusDegraded)
}

// OkStatus returns a HealthStatus indicating the service is ok
func OkStatus() *HealthStatus {
	return NewHealthStatus(StatusOk)
}
