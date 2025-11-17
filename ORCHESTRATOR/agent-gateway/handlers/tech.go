package handlers

import (
	"net/http"
	"strconv"

	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/models"

	"github.com/gorilla/mux"
)

// TechHandler handles tech tips endpoints
type TechHandler struct {
	db  *db.Manager
	exe *executor.Executor
}

// NewTechHandler creates a new tech handler
func NewTechHandler(dbManager *db.Manager, exec *executor.Executor) *TechHandler {
	return &TechHandler{
		db:  dbManager,
		exe: exec,
	}
}

// GetToday handles GET /api/tech/today
func (h *TechHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	tip, err := h.exe.GetTechToday()
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	models.WriteSuccess(w, tip)
}

// GetByDate handles GET /api/tech/date/:date
func (h *TechHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	tip, err := h.exe.GetTechByDate(date)
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	models.WriteSuccess(w, tip)
}

// GetRandom handles GET /api/tech/random
func (h *TechHandler) GetRandom(w http.ResponseWriter, r *http.Request) {
	tip, err := h.db.GetTechRandom()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, tip)
}

// GetLatest handles GET /api/tech/latest/:n
func (h *TechHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nStr := vars["n"]

	n, err := strconv.Atoi(nStr)
	if err != nil || n <= 0 {
		models.WriteError(w, http.StatusBadRequest, "Invalid number parameter")
		return
	}

	// Limit to reasonable number
	if n > 100 {
		n = 100
	}

	tips, err := h.db.GetTechLatest(n)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, tips)
}

// GetByCategory handles GET /api/tech/category/:category
func (h *TechHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["category"]

	tips, err := h.db.GetTechByCategory(category)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, tips)
}

// GetAll handles GET /api/tech/all
func (h *TechHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters from query string
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := r.URL.Query().Get("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 && s <= 100 {
			pageSize = s
		}
	}

	result, err := h.db.GetTechAll(page, pageSize)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, result)
}
