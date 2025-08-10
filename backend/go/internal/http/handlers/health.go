package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/provemyself/backend/internal/store"
	"github.com/provemyself/backend/internal/types"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	database *store.Database
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(database *store.Database) *HealthHandler {
	return &HealthHandler{database: database}
}

// GetHealth handles GET /api/v1/health
// @Summary Health check endpoint
// @Description Returns the health status of the API service
// @Tags System
// @Produce json
// @Success 200 {object} types.HealthResponse
// @Failure 503 {object} types.ErrorResponse
// @Router /health [get]
func (h *HealthHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check database health
	dbStatus := "healthy"
	if err := h.database.HealthCheck(ctx); err != nil {
		dbStatus = "unhealthy"
	}

	// Determine overall status
	status := "healthy"
	statusCode := http.StatusOK
	if dbStatus == "unhealthy" {
		status = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	}

	response := types.HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   "0.1.0",
		Services: &types.HealthServices{
			Database: dbStatus,
			Storage:  "healthy", // In Phase 0, we assume healthy
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}