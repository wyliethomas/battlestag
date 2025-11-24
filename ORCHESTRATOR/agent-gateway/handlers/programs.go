package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"agent-gateway/models"
	"agent-gateway/programs"
)

// ProgramsHandler handles program-related requests
type ProgramsHandler struct {
	registry *programs.Registry
}

// NewProgramsHandler creates a new programs handler
func NewProgramsHandler(registry *programs.Registry) *ProgramsHandler {
	return &ProgramsHandler{
		registry: registry,
	}
}

// ListPrograms returns all available programs
// GET /api/programs/list
func (h *ProgramsHandler) ListPrograms(w http.ResponseWriter, r *http.Request) {
	programsList := h.registry.List()

	response := models.Response{
		Success: true,
		Data:    programsList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExecuteProgram executes a program with provided parameters
// POST /api/programs/execute
// Body: {"program_id": "echo", "parameters": {"message": "hello"}}
func (h *ProgramsHandler) ExecuteProgram(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ProgramID  string                 `json:"program_id"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response := models.Response{
			Success: false,
			Error:   "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if request.ProgramID == "" {
		response := models.Response{
			Success: false,
			Error:   "program_id is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Log program execution for debugging
	paramsJSON, _ := json.Marshal(request.Parameters)
	log.Printf("[PROGRAM] Executing: %s with params: %s", request.ProgramID, string(paramsJSON))

	// Execute program with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := h.registry.Execute(ctx, request.ProgramID, request.Parameters)

	// Log result
	if err != nil {
		log.Printf("[PROGRAM] Error executing %s: %v", request.ProgramID, err)
	} else if !result.Success {
		log.Printf("[PROGRAM] Failed %s: %s", request.ProgramID, result.Error)
	} else {
		log.Printf("[PROGRAM] Success %s", request.ProgramID)
	}
	if err != nil {
		response := models.Response{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.Response{
		Success: result.Success,
		Data:    result,
	}

	if !result.Success {
		response.Error = result.Error
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
