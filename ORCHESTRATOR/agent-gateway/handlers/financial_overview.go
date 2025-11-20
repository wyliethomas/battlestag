package handlers

import (
	"net/http"

	"agent-gateway/db"
	"agent-gateway/models"
)

// FinancialOverviewHandler handles financial overview endpoints
type FinancialOverviewHandler struct {
	dbManager *db.Manager
}

// NewFinancialOverviewHandler creates a new financial overview handler
func NewFinancialOverviewHandler(dbManager *db.Manager) *FinancialOverviewHandler {
	return &FinancialOverviewHandler{
		dbManager: dbManager,
	}
}

// GetNetWorth calculates and returns net worth (assets - liabilities)
// GET /api/financial/net-worth
func (h *FinancialOverviewHandler) GetNetWorth(w http.ResponseWriter, r *http.Request) {
	overview, err := h.dbManager.GetFinancialOverview()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"net_worth":        overview.NetWorth,
		"total_assets":     overview.TotalAssets,
		"total_liabilities": overview.TotalLiabilities,
		"timestamp":        overview.Timestamp,
	})
}

// GetSummary returns a complete financial snapshot
// GET /api/financial/summary
func (h *FinancialOverviewHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	overview, err := h.dbManager.GetFinancialOverview()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, overview)
}
