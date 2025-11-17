package handlers

import (
	"net/http"
	"time"

	"agent-gateway/db"
	"agent-gateway/models"
)

// MetaHandler handles meta/health endpoints
type MetaHandler struct {
	db        *db.Manager
	startTime time.Time
}

// NewMetaHandler creates a new meta handler
func NewMetaHandler(dbManager *db.Manager, startTime time.Time) *MetaHandler {
	return &MetaHandler{
		db:        dbManager,
		startTime: startTime,
	}
}

// GetHealth handles GET /api/health
func (h *MetaHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := h.db.CheckHealth()

	// Return 503 if any critical service is down
	if health.Status != "healthy" {
		models.WriteJSON(w, http.StatusServiceUnavailable, models.Response{
			Success: true,
			Data:    health,
		})
		return
	}

	models.WriteSuccess(w, health)
}

// GetStats handles GET /api/stats
func (h *MetaHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.db.GetStats(h.startTime)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, stats)
}
