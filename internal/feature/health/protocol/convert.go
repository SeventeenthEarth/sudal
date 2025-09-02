package protocol

import (
	"strings"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

// NormalizeStatusStr normalizes a raw status string to one of: "healthy", "unhealthy", "unknown".
// It is case-insensitive and defaults to "unknown" for unrecognized values.
func NormalizeStatusStr(status string) string {
	switch strings.ToLower(status) {
	case entity.StatusHealthy:
		return entity.StatusHealthy
	case entity.StatusUnhealthy:
		return entity.StatusUnhealthy
	case entity.StatusOk:
		// Treat "ok" as healthy for normalization
		return entity.StatusHealthy
	case entity.StatusDegraded:
		// Treat "degraded" as unhealthy for normalization
		return entity.StatusUnhealthy
	case entity.StatusUnknown:
		return entity.StatusUnknown
	default:
		return entity.StatusUnknown
	}
}

// NormalizeStatus normalizes an entity.HealthStatus pointer safely, treating nil as "unknown".
func NormalizeStatus(s *entity.HealthStatus) string {
	if s == nil {
		return entity.StatusUnknown
	}
	return NormalizeStatusStr(s.Status)
}

// ToProtoServingStatus converts a domain HealthStatus to the proto ServingStatus enum.
// Nil input or unknown values map to UNKNOWN_UNSPECIFIED.
func ToProtoServingStatus(s *entity.HealthStatus) healthv1.ServingStatus {
    switch NormalizeStatus(s) {
    case entity.StatusHealthy:
        return healthv1.ServingStatus_SERVING_STATUS_SERVING
    case entity.StatusUnhealthy:
        return healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING
    default:
        return healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED
    }
}
