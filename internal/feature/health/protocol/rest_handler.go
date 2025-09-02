// NOTE: Test-only REST handler
//
// This file provides lightweight net/http handlers for health endpoints that
// are primarily used by unit/integration tests and local experimentation.
// Production REST traffic is served by the OpenAPI-generated router under
// the "/api/*" path prefix. See `api/openapi.yaml` and
// `internal/infrastructure/openapi` for the production implementation.
//
// Keep this handler minimal and behaviorally equivalent to the OpenAPI routes
// so tests can validate logic without the full OpenAPI stack.
package protocol

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain/entity"
)

// HealthHandler handles health-related HTTP requests
type HealthHandler struct {
	service application.HealthService
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(service application.HealthService) *HealthHandler {
	return &HealthHandler{
		service: service,
	}
}

// RegisterRoutes registers the health check routes
func (h *HealthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", h.Ping)
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("GET /health/database", h.DatabaseHealth)
}

// Ping handles the ping endpoint
func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := h.service.Ping(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// Health handles the health check endpoint
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status, err := h.service.Check(ctx)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

// DatabaseHealth handles the database health check endpoint
func (h *HealthHandler) DatabaseHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbStatus, err := h.service.CheckDatabase(ctx)
	if err != nil {
		// Create error response with proper structure
		// Use the dbStatus if available (repository may return both status and error)
		var databaseStatus *entity.DatabaseStatus
		if dbStatus != nil {
			databaseStatus = dbStatus
		} else {
			databaseStatus = entity.UnhealthyDatabaseStatus(err.Error())
		}

		errorResponse := &entity.DetailedHealthStatus{
			Status:    "error",
			Message:   "Database health check failed",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Database:  databaseStatus,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if encErr := json.NewEncoder(w).Encode(errorResponse); encErr != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Create successful response
	response := &entity.DetailedHealthStatus{
		Status:    "healthy",
		Message:   "Database is healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
