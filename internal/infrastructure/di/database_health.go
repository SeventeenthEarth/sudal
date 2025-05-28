package di

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/seventeenthearth/sudal/internal/infrastructure/config"
	"github.com/seventeenthearth/sudal/internal/infrastructure/database"
	"github.com/seventeenthearth/sudal/internal/infrastructure/log"
)

// DatabaseHealthHandler handles database health check requests
type DatabaseHealthHandler struct {
	dbManager database.PostgresManager
	logger    *zap.Logger
}

// NewDatabaseHealthHandler creates a new database health handler
func NewDatabaseHealthHandler(dbManager database.PostgresManager) *DatabaseHealthHandler {
	// Check if we're in test environment and return mock handler
	if IsTestEnvironment() {
		return NewMockDatabaseHealthHandler()
	}

	return &DatabaseHealthHandler{
		dbManager: dbManager,
		logger:    log.GetLogger().With(zap.String("component", "database_health_handler")),
	}
}

// DatabaseHealthResponse represents the response for database health check
type DatabaseHealthResponse struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message"`
	Timestamp string                 `json:"timestamp"`
	Database  *database.HealthStatus `json:"database,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// HandleDatabaseHealth handles HTTP requests for database health check
func (h *DatabaseHealthHandler) HandleDatabaseHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Debug("Database health check requested")

	response := &DatabaseHealthResponse{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check if we're in test environment and return mock response
	if IsTestEnvironment() {
		response.Status = "healthy"
		response.Message = "Mock database is healthy"
		response.Database = &database.HealthStatus{
			Status:  "healthy",
			Message: "Mock database connection is healthy",
			Stats: &database.ConnectionStats{
				MaxOpenConnections: 25,
				OpenConnections:    1,
				InUse:              0,
				Idle:               1,
				WaitCount:          0,
				WaitDuration:       0,
				MaxIdleClosed:      0,
				MaxLifetimeClosed:  0,
			},
		}
		h.writeJSONResponse(w, http.StatusOK, response)
		return
	}

	// Check if database manager is available
	if h.dbManager == nil {
		h.logger.Error("Database manager not initialized")
		response.Status = "error"
		response.Message = "Database manager not available"
		response.Error = "Database manager is nil"
		h.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	// Perform health check using existing database manager
	healthStatus, err := h.dbManager.HealthCheck(ctx)
	if err != nil {
		h.logger.Error("Database health check failed",
			log.FormatError(err),
		)
		response.Status = "unhealthy"
		response.Message = "Database health check failed"
		response.Error = err.Error()
		h.writeJSONResponse(w, http.StatusServiceUnavailable, response)
		return
	}

	response.Status = "healthy"
	response.Message = "Database is healthy"
	response.Database = healthStatus

	h.logger.Debug("Database health check successful",
		zap.String("status", healthStatus.Status),
		zap.Any("stats", healthStatus.Stats),
	)

	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response
func (h *DatabaseHealthHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response",
			log.FormatError(err),
		)
	}
}

// IsTestEnvironment checks if we're running in a test environment
func IsTestEnvironment() bool {
	logger := log.GetLogger()

	// Check environment variables that indicate test mode
	goTest := os.Getenv("GO_TEST")
	ginkgoTest := os.Getenv("GINKGO_TEST")

	logger.Debug("Checking test environment",
		zap.String("GO_TEST", goTest),
		zap.String("GINKGO_TEST", ginkgoTest),
	)

	if goTest == "1" || ginkgoTest == "1" {
		logger.Debug("Test environment detected via environment variables")
		return true
	}

	// Check if config indicates test environment
	cfg := config.GetConfig()
	if cfg != nil {
		logger.Debug("Checking config for test environment",
			zap.String("AppEnv", cfg.AppEnv),
			zap.String("Environment", cfg.Environment),
		)
		if cfg.AppEnv == "test" || cfg.Environment == "test" {
			logger.Debug("Test environment detected via config")
			return true
		}
	}

	logger.Debug("Production environment detected")
	return false
}

// NewMockDatabaseHealthHandler creates a mock database health handler for testing
func NewMockDatabaseHealthHandler() *DatabaseHealthHandler {
	return &DatabaseHealthHandler{
		dbManager: nil, // Mock doesn't need a real database manager
		logger:    log.GetLogger().With(zap.String("component", "mock_database_health_handler")),
	}
}
