package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/zavieruka/video-platform/backend/internal/config"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	config *config.Config
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config: cfg,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
}

// HandleHealth returns the health status of the application
// This is used by Cloud Run for health checks
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:      "healthy",
		Timestamp:   time.Now().UTC(),
		Version:     "0.1.0", // TODO: Get from build info
		Environment: h.config.Environment,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error but don't fail the health check
		// In production, we'd use structured logging here
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// HandleReady checks if the application is ready to serve traffic
// This checks GCP client connections
func (h *HealthHandler) HandleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if GCP clients are initialized
	if h.config.FirestoreClient == nil || h.config.StorageClient == nil {
		response := map[string]string{
			"status":  "not ready",
			"message": "GCP clients not initialized",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(response)
		return
	}

	// TODO: Add actual health checks for Firestore and Storage
	// For now, just check if clients exist

	response := map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
		"checks": map[string]string{
			"firestore": "ok",
			"storage":   "ok",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
