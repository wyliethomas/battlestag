package handlers

import (
	"net/http"
	"strconv"

	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/models"

	"github.com/gorilla/mux"
)

// StoicHandler handles stoic thoughts endpoints
type StoicHandler struct {
	db  *db.Manager
	exe *executor.Executor
}

// NewStoicHandler creates a new stoic handler
func NewStoicHandler(dbManager *db.Manager, exec *executor.Executor) *StoicHandler {
	return &StoicHandler{
		db:  dbManager,
		exe: exec,
	}
}

// GetToday handles GET /api/stoic/today
func (h *StoicHandler) GetToday(w http.ResponseWriter, r *http.Request) {
	thought, err := h.exe.GetStoicToday()
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	models.WriteSuccess(w, thought)
}

// GetByDate handles GET /api/stoic/date/:date
func (h *StoicHandler) GetByDate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	thought, err := h.exe.GetStoicByDate(date)
	if err != nil {
		models.WriteError(w, http.StatusNotFound, err.Error())
		return
	}

	models.WriteSuccess(w, thought)
}

// GetRandom handles GET /api/stoic/random
func (h *StoicHandler) GetRandom(w http.ResponseWriter, r *http.Request) {
	thought, err := h.db.GetStoicRandom()
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, thought)
}

// GetLatest handles GET /api/stoic/latest/:n
func (h *StoicHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
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

	thoughts, err := h.db.GetStoicLatest(n)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, thoughts)
}

// GetAll handles GET /api/stoic/all
func (h *StoicHandler) GetAll(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.db.GetStoicAll(page, pageSize)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	models.WriteSuccess(w, result)
}
