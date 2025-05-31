package protocol

import (
	"connectrpc.com/connect"
	"context"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"go.uber.org/zap"
)

// HealthManager implements the Connect-go health service
type HealthManager struct {
	healthService application.HealthService
}

// NewHealthAdapter creates a new health service handler
func NewHealthAdapter(service application.HealthService) *HealthManager {
	return &HealthManager{
		healthService: service,
	}
}

// Check implements the Check RPC method
func (h *HealthManager) Check(
	ctx context.Context,
	req *connect.Request[healthv1.CheckRequest],
) (*connect.Response[healthv1.CheckResponse], error) {
	log.InfoContext(ctx, "Health check requested")

	// Call the application service to perform the health check
	status, err := h.healthService.Check(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Health check failed", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map the domain status to the proto status
	var protoStatus healthv1.ServingStatus
	switch status.Status {
	case "healthy":
		protoStatus = healthv1.ServingStatus_SERVING_STATUS_SERVING
	case "unhealthy":
		protoStatus = healthv1.ServingStatus_SERVING_STATUS_NOT_SERVING
	default:
		protoStatus = healthv1.ServingStatus_SERVING_STATUS_UNKNOWN_UNSPECIFIED
	}

	// Create and return the response
	response := &healthv1.CheckResponse{
		Status: protoStatus,
	}

	log.InfoContext(ctx, "Health check completed", zap.String("status", protoStatus.String()))
	return connect.NewResponse(response), nil
}
