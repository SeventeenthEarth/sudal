package connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/seventeenthearth/sudal/gen/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"go.uber.org/zap"
)

// HealthServiceHandler implements the Connect-go health service
type HealthServiceHandler struct {
	service application.Service
}

// NewHealthServiceHandler creates a new health service handler
func NewHealthServiceHandler(service application.Service) *HealthServiceHandler {
	return &HealthServiceHandler{
		service: service,
	}
}

// Check implements the Check RPC method
func (h *HealthServiceHandler) Check(
	ctx context.Context,
	req *connect.Request[healthv1.CheckRequest],
) (*connect.Response[healthv1.CheckResponse], error) {
	log.InfoContext(ctx, "Health check requested")

	// Call the application service to perform the health check
	status, err := h.service.Check(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Health check failed", zap.Error(err))
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	// Map the domain status to the proto status
	var protoStatus healthv1.ServingStatus
	switch status.Status {
	case "healthy":
		protoStatus = healthv1.ServingStatus_SERVING
	case "unhealthy":
		protoStatus = healthv1.ServingStatus_NOT_SERVING
	default:
		protoStatus = healthv1.ServingStatus_UNKNOWN
	}

	// Create and return the response
	response := &healthv1.CheckResponse{
		Status: protoStatus,
	}

	log.InfoContext(ctx, "Health check completed", zap.String("status", protoStatus.String()))
	return connect.NewResponse(response), nil
}
