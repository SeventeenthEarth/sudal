package domain

// Status represents the health status of the service
type Status struct {
	Status string `json:"status"`
}

// NewStatus creates a new Status with the given status string
func NewStatus(status string) *Status {
	return &Status{
		Status: status,
	}
}

// HealthyStatus returns a Status indicating the service is healthy
func HealthyStatus() *Status {
	return NewStatus("healthy")
}

// OkStatus returns a Status indicating the service is ok
func OkStatus() *Status {
	return NewStatus("ok")
}
