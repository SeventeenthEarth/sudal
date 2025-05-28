package interfaces

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
	"github.com/seventeenthearth/sudal/internal/feature/health/domain"
)

// Handler handles health-related HTTP requests
type Handler struct {
	service application.Service
}

// NewHandler creates a new health handler
func NewHandler(service application.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers the health check routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ping", h.Ping)
	mux.HandleFunc("GET /healthz", h.Health)
	mux.HandleFunc("GET /health/database", h.DatabaseHealth)
}

// Ping handles the ping endpoint
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
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
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
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
func (h *Handler) DatabaseHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dbStatus, err := h.service.CheckDatabase(ctx)
	if err != nil {
		// Create error response with proper structure
		errorResponse := &domain.DetailedHealthStatus{
			Status:    "error",
			Message:   "Database health check failed",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Database:  domain.UnhealthyDatabaseStatus(err.Error()),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		if encErr := json.NewEncoder(w).Encode(errorResponse); encErr != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
		return
	}

	// Create successful response
	response := &domain.DetailedHealthStatus{
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
