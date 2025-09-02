package protocol

import (
	"testing"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

func TestNormalizeStatusStr(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{entity.StatusHealthy, entity.StatusHealthy},
		{"HEALTHY", entity.StatusHealthy},
		{entity.StatusUnhealthy, entity.StatusUnhealthy},
		{"UNHEALTHY", entity.StatusUnhealthy},
		{entity.StatusUnknown, entity.StatusUnknown},
		{"", entity.StatusUnknown},
		{"custom", entity.StatusUnknown},
	}

	for _, tt := range tests {
		if got := NormalizeStatusStr(tt.in); got != tt.want {
			t.Fatalf("NormalizeStatusStr(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeStatus(t *testing.T) {
	if got := NormalizeStatus(nil); got != entity.StatusUnknown {
		t.Fatalf("NormalizeStatus(nil) = %q; want %q", got, entity.StatusUnknown)
	}
	if got := NormalizeStatus(entity.HealthyStatus()); got != entity.StatusHealthy {
		t.Fatalf("NormalizeStatus(healthy) = %q; want %q", got, entity.StatusHealthy)
	}
	if got := NormalizeStatus(entity.UnhealthyStatus()); got != entity.StatusUnhealthy {
		t.Fatalf("NormalizeStatus(unhealthy) = %q; want %q", got, entity.StatusUnhealthy)
	}
}

func TestToProtoServingStatus(t *testing.T) {
	tests := []struct {
		name string
		in   *entity.HealthStatus
		want healthv1.ServingStatus
	}{
		{"nil maps to unknown", nil, healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED},
		{"healthy maps to serving", entity.HealthyStatus(), healthv1.ServingStatus_SERVING_STATUS_SERVING},
		{"unhealthy maps to not serving", entity.UnhealthyStatus(), healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING},
		{"unknown maps to unknown", entity.UnknownStatus(), healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED},
		{"custom maps to unknown", entity.NewHealthStatus("custom"), healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToProtoServingStatus(tt.in); got != tt.want {
				t.Fatalf("ToProtoServingStatus(%v) = %v; want %v", tt.in, got, tt.want)
			}
		})
	}
}
