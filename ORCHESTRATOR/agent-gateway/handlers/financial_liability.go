package handlers

import (
	"net/http"

	"github.com/gorilla/mux"

	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/models"
)

// FinancialLiabilityHandler handles financial liability tracker endpoints
type FinancialLiabilityHandler struct {
	executor  *executor.Executor
	dbManager *db.Manager
}

// NewFinancialLiabilityHandler creates a new financial liability handler
func NewFinancialLiabilityHandler(exec *executor.Executor, dbManager *db.Manager) *FinancialLiabilityHandler {
	return &FinancialLiabilityHandler{
		executor:  exec,
		dbManager: dbManager,
	}
}

// AddLiability adds a new liability
// POST /api/financial-liability
func (h *FinancialLiabilityHandler) AddLiability(w http.ResponseWriter, r *http.Request) {
	var req models.AddLiabilityRequest
	if err := models.ParseJSONBody(r, &req); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call executor to add liability
	liability, err := h.executor.AddLiability(&req)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, liability)
}

// ListLiabilities lists all liabilities with optional type filter
// GET /api/financial-liability?type=credit-card
func (h *FinancialLiabilityHandler) ListLiabilities(w http.ResponseWriter, r *http.Request) {
	liabilityType := models.GetQueryParam(r, "type", "")

	liabilities, err := h.executor.ListLiabilities(liabilityType)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"liabilities": liabilities,
		"count":       len(liabilities),
	})
}

// GetLiability retrieves a specific liability
// GET /api/financial-liability/{name}?include_history=true
func (h *FinancialLiabilityHandler) GetLiability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "liability name is required")
		return
	}

	includeHistory := models.GetQueryParamBool(r, "include_history", false)

	liability, history, err := h.executor.GetLiability(name, includeHistory)
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	response := map[string]interface{}{
		"liability": liability,
	}

	if includeHistory && len(history) > 0 {
		response["balance_history"] = history
	}

	models.WriteSuccess(w, response)
}

// UpdateLiability updates a liability's balance
// PUT /api/financial-liability/{name}
func (h *FinancialLiabilityHandler) UpdateLiability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "liability name is required")
		return
	}

	var req models.UpdateLiabilityRequest
	if err := models.ParseJSONBody(r, &req); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		models.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Call executor to update liability
	if err := h.executor.UpdateLiability(name, req.Balance, req.Notes); err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"message": "Liability updated successfully",
		"name":    name,
	})
}

// DeleteLiability deletes a liability
// DELETE /api/financial-liability/{name}
func (h *FinancialLiabilityHandler) DeleteLiability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		models.WriteError(w, http.StatusBadRequest, "liability name is required")
		return
	}

	if err := h.executor.DeleteLiability(name); err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"message": "Liability deleted successfully",
		"name":    name,
	})
}

// GetTotal returns total liability balance
// GET /api/financial-liability/total
func (h *FinancialLiabilityHandler) GetTotal(w http.ResponseWriter, r *http.Request) {
	total, err := h.executor.GetTotalLiabilities()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, map[string]interface{}{
		"total_balance": total,
	})
}

// GetSummary returns liability summary grouped by type
// GET /api/financial-liability/summary
func (h *FinancialLiabilityHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.dbManager.GetLiabilitySummary()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, summary)
}
