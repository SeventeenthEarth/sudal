package entity

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
	return NewHealthStatus("unknown")
}

// HealthyStatus returns a HealthStatus indicating the service is healthy
func HealthyStatus() *HealthStatus {
	return NewHealthStatus("healthy")
}

// UnhealthyStatus returns a HealthStatus indicating the service is unhealthy
func UnhealthyStatus() *HealthStatus {
	return NewHealthStatus("unhealthy")
}

// DegradedStatus returns a HealthStatus indicating the service is degraded
func DegradedStatus() *HealthStatus {
	return NewHealthStatus("degraded")
}

// OkStatus returns a HealthStatus indicating the service is ok
func OkStatus() *HealthStatus {
	return NewHealthStatus("ok")
}
