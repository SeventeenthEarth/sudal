package protocol

import (
	"testing"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

func TestNormalizeStatusStr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"healthy lower", entity.StatusHealthy, entity.StatusHealthy},
		{"healthy upper", "HEALTHY", entity.StatusHealthy},
		{"unhealthy lower", entity.StatusUnhealthy, entity.StatusUnhealthy},
		{"unhealthy upper", "UNHEALTHY", entity.StatusUnhealthy},
		{"ok -> healthy", entity.StatusOk, entity.StatusHealthy},
		{"degraded -> unhealthy", entity.StatusDegraded, entity.StatusUnhealthy},
		{"unknown explicit", entity.StatusUnknown, entity.StatusUnknown},
		{"empty -> unknown", "", entity.StatusUnknown},
		{"custom -> unknown", "custom", entity.StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeStatusStr(tt.in); got != tt.want {
				t.Fatalf("NormalizeStatusStr(%q) = %q; want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := []struct {
		name string
		in   *entity.HealthStatus
		want string
	}{
		{"nil -> unknown", nil, entity.StatusUnknown},
		{"healthy -> healthy", entity.HealthyStatus(), entity.StatusHealthy},
		{"unhealthy -> unhealthy", entity.UnhealthyStatus(), entity.StatusUnhealthy},
		{"ok -> healthy", entity.NewHealthStatus(entity.StatusOk), entity.StatusHealthy},
		{"degraded -> unhealthy", entity.NewHealthStatus(entity.StatusDegraded), entity.StatusUnhealthy},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := NormalizeStatus(c.in); got != c.want {
				t.Fatalf("NormalizeStatus(%v) = %q; want %q", c.in, got, c.want)
			}
		})
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
