package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/models"
)

// FinancialAssetHandler handles financial asset tracker endpoints
type FinancialAssetHandler struct {
	executor  *executor.Executor
	dbManager *db.Manager
}

// NewFinancialAssetHandler creates a new financial asset handler
func NewFinancialAssetHandler(exec *executor.Executor, dbManager *db.Manager) *FinancialAssetHandler {
	return &FinancialAssetHandler{
		executor:  exec,
		dbManager: dbManager,
	}
}

// AddAsset adds a new asset
// POST /api/financial-asset
func (h *FinancialAssetHandler) AddAsset(w http.ResponseWriter, r *http.Request) {
	var req models.AddAssetRequest
	if err := models.ParseJSONBody(r, &req); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call executor to add asset
	var purchaseDate string
	if req.PurchaseDate != nil {
		purchaseDate = *req.PurchaseDate
	}

	asset, err := h.executor.AddAsset(
		req.Name,
		req.Category,
		req.CurrentValue,
		req.PurchasePrice,
		purchaseDate,
		req.Notes,
	)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, asset)
}

// ListAssets lists all assets with optional filters
// GET /api/financial-asset?include_removed=false&category=stocks
func (h *FinancialAssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	includeRemoved := models.GetQueryParamBool(r, "include_removed", false)
	category := models.GetQueryParam(r, "category", "")

	assets, err := h.executor.ListAssets(includeRemoved, category)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"assets": assets,
		"count":  len(assets),
	})
}

// GetAsset retrieves a specific asset
// GET /api/financial-asset/{name}
func (h *FinancialAssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "asset name is required")
		return
	}

	asset, err := h.executor.GetAsset(name)
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	models.WriteSuccess(w, asset)
}

// UpdateAsset updates an asset's value
// PUT /api/financial-asset/{name}
func (h *FinancialAssetHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "asset name is required")
		return
	}

	var req models.UpdateAssetRequest
	if err := models.ParseJSONBody(r, &req); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call executor to update asset
	if err := h.executor.UpdateAssetValue(name, req.CurrentValue, req.Notes); err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"message": "Asset updated successfully",
		"name":    name,
	})
}

// RemoveAsset soft-deletes an asset
// DELETE /api/financial-asset/{name}
func (h *FinancialAssetHandler) RemoveAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "asset name is required")
		return
	}

	if err := h.executor.RemoveAsset(name); err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"message": "Asset removed successfully",
		"name":    name,
	})
}

// RestoreAsset restores a removed asset
// POST /api/financial-asset/{name}/restore
func (h *FinancialAssetHandler) RestoreAsset(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "asset name is required")
		return
	}

	if err := h.executor.RestoreAsset(name); err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"message": "Asset restored successfully",
		"name":    name,
	})
}

// GetSummary returns asset portfolio summary
// GET /api/financial-asset/summary
func (h *FinancialAssetHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dbManager.GetAssetSummary()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, summary)
}
