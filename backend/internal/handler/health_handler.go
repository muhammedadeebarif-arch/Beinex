package handler

import (
	"context"
	"net/http"
	"time"

	"ems-backend/internal/database"
	"ems-backend/internal/response"
)

// HealthHandler provides health check endpoints.
type HealthHandler struct {
	db *database.DB
}

// NewHealthHandler creates a HealthHandler.
func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// HealthCheck reports the health of the application and PostgreSQL database.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "UP"
	if err := h.db.Pool.PingContext(ctx); err != nil {
		dbStatus = "DOWN"
		response.JSON(w, http.StatusServiceUnavailable, map[string]any{
			"status":   "DEGRADED",
			"database": dbStatus,
			"error":    "Database ping failed",
		})
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"status":   "UP",
		"database": dbStatus,
		"time":     time.Now().UTC().Format(time.RFC3339),
	})
}
