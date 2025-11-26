package api

import (
	"agent-gateway/gateway"
	"agent-gateway/llm"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handler manages HTTP requests for the gateway
type Handler struct {
	gateway *gateway.Gateway
}

// NewHandler creates a new API handler
func NewHandler(gw *gateway.Gateway) *Handler {
	return &Handler{
		gateway: gw,
	}
}

// ChatRequest represents the incoming chat request
type ChatRequest struct {
	Messages []llm.Message `json:"messages"`
}

// ChatResponse represents the outgoing chat response
type ChatResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// HandleChat processes chat requests
func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		respondWithError(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Messages) == 0 {
		respondWithError(w, "No messages provided", http.StatusBadRequest)
		return
	}

	// Process chat through gateway
	log.Printf("Processing chat with %d messages", len(req.Messages))
	response, err := h.gateway.ProcessChat(req.Messages)
	if err != nil {
		log.Printf("Error processing chat: %v", err)
		respondWithError(w, fmt.Sprintf("Error processing chat: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	respondWithJSON(w, ChatResponse{
		Response: response,
	})
}

// HandleHealth returns the health status of the gateway
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	registry := h.gateway.GetRegistry()

	health := map[string]interface{}{
		"status":          "healthy",
		"programs_loaded": registry.Count(),
	}

	respondWithJSON(w, health)
}

// HandlePrograms lists all available programs
func (h *Handler) HandlePrograms(w http.ResponseWriter, r *http.Request) {
	registry := h.gateway.GetRegistry()
	programs := make([]map[string]string, 0)

	for name, program := range registry.GetAllPrograms() {
		programs = append(programs, map[string]string{
			"name":        name,
			"description": program.Manifest.Description,
			"executable":  program.ExecutablePath,
		})
	}

	respondWithJSON(w, map[string]interface{}{
		"programs": programs,
		"count":    len(programs),
	})
}

// Helper functions

func respondWithJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ChatResponse{
		Error: message,
	})
}
