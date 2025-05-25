package openapi

import (
	"context"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
	"go.uber.org/zap"
)

// OpenAPIHandler implements the ogen-generated Handler interface
type OpenAPIHandler struct {
	healthService application.Service
}

// NewOpenAPIHandler creates a new OpenAPI handler
func NewOpenAPIHandler(healthService application.Service) *OpenAPIHandler {
	return &OpenAPIHandler{
		healthService: healthService,
	}
}

// Ping implements the ping operation
func (h *OpenAPIHandler) Ping(ctx context.Context) (PingRes, error) {
	log.InfoContext(ctx, "OpenAPI ping requested")
	
	// Simple ping response
	response := &PingResponse{
		Status: "ok",
	}
	
	log.InfoContext(ctx, "OpenAPI ping completed", zap.String("status", response.Status))
	return response, nil
}

// Health implements the health operation
func (h *OpenAPIHandler) Health(ctx context.Context) (HealthRes, error) {
	log.InfoContext(ctx, "OpenAPI health check requested")
	
	// Call the application service to perform the health check
	status, err := h.healthService.Check(ctx)
	if err != nil {
		log.ErrorContext(ctx, "OpenAPI health check failed", zap.Error(err))
		
		// Return service unavailable response
		response := &HealthServiceUnavailable{
			Status: HealthResponseStatusUnhealthy,
		}
		return response, nil
	}
	
	// Map the domain status to the OpenAPI status
	var apiStatus HealthResponseStatus
	switch status.Status {
	case "healthy":
		apiStatus = HealthResponseStatusHealthy
	case "unhealthy":
		apiStatus = HealthResponseStatusUnhealthy
	default:
		apiStatus = HealthResponseStatusUnhealthy
	}
	
	// Determine response type based on status
	if apiStatus == HealthResponseStatusHealthy {
		response := &HealthOK{
			Status: apiStatus,
		}
		log.InfoContext(ctx, "OpenAPI health check completed", zap.String("status", string(apiStatus)))
		return response, nil
	} else {
		response := &HealthServiceUnavailable{
			Status: apiStatus,
		}
		log.InfoContext(ctx, "OpenAPI health check completed", zap.String("status", string(apiStatus)))
		return response, nil
	}
}

// DatabaseHealth implements the database health operation
func (h *OpenAPIHandler) DatabaseHealth(ctx context.Context) (DatabaseHealthRes, error) {
	log.InfoContext(ctx, "OpenAPI database health check requested")
	
	// Call the application service to perform the health check
	status, err := h.healthService.Check(ctx)
	if err != nil {
		log.ErrorContext(ctx, "OpenAPI database health check failed", zap.Error(err))
		
		// Return service unavailable response
		response := &DatabaseHealthServiceUnavailable{
			Status:   DatabaseHealthResponseStatusUnhealthy,
			Database: DatabaseHealthResponseDatabaseDisconnected,
		}
		return response, nil
	}
	
	// Map the domain status to the OpenAPI status
	var apiStatus DatabaseHealthResponseStatus
	var dbStatus DatabaseHealthResponseDatabase
	
	switch status.Status {
	case "healthy":
		apiStatus = DatabaseHealthResponseStatusHealthy
		dbStatus = DatabaseHealthResponseDatabaseConnected
	case "unhealthy":
		apiStatus = DatabaseHealthResponseStatusUnhealthy
		dbStatus = DatabaseHealthResponseDatabaseDisconnected
	default:
		apiStatus = DatabaseHealthResponseStatusUnhealthy
		dbStatus = DatabaseHealthResponseDatabaseDisconnected
	}
	
	// Determine response type based on status
	if apiStatus == DatabaseHealthResponseStatusHealthy {
		response := &DatabaseHealthOK{
			Status:   apiStatus,
			Database: dbStatus,
		}
		log.InfoContext(ctx, "OpenAPI database health check completed", 
			zap.String("status", string(apiStatus)),
			zap.String("database", string(dbStatus)))
		return response, nil
	} else {
		response := &DatabaseHealthServiceUnavailable{
			Status:   apiStatus,
			Database: dbStatus,
		}
		log.InfoContext(ctx, "OpenAPI database health check completed", 
			zap.String("status", string(apiStatus)),
			zap.String("database", string(dbStatus)))
		return response, nil
	}
}
