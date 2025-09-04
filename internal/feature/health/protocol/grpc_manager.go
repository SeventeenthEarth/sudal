package protocol

import (
	"context"

	"connectrpc.com/connect"

	healthv1 "github.com/seventeenthearth/sudal/gen/go/health/v1"
	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	log "github.com/seventeenthearth/sudal/internal/service/logger"
	"go.uber.org/zap"
)

// HealthManager implements the Connect-go health service
type HealthManager struct {
	healthService application.HealthService
}

// NewHealthManager creates a new health service handler
func NewHealthManager(service application.HealthService) *HealthManager {
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

	// Map the domain status to the proto status via converter
	protoStatus := ToProtoServingStatus(status)

	// Create and return the response
	response := &healthv1.CheckResponse{
		Status: protoStatus,
	}

	log.InfoContext(ctx, "Health check completed", zap.String("status", protoStatus.String()))
	return connect.NewResponse(response), nil
}
