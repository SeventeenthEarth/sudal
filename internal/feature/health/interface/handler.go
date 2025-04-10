package interfaces

import (
	"encoding/json"
	"net/http"

	"github.com/seventeenthearth/sudal/internal/feature/health/application"
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
