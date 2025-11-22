package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/battlestag/agent-gateway/llm"
	"github.com/battlestag/agent-gateway/models"
)

// LLMHandler handles LLM chat requests
type LLMHandler struct {
	client *llm.Client
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(client *llm.Client) *LLMHandler {
	return &LLMHandler{
		client: client,
	}
}

// Chat handles POST /api/llm/chat
func (h *LLMHandler) Chat(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate message
	if req.Message == "" {
		models.WriteError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Convert chat history to LLM messages
	history := make([]llm.Message, len(req.History))
	for i, msg := range req.History {
		history[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Send to LLM
	response, err := h.client.Chat(req.Message, history)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, "LLM error: "+err.Error())
		return
	}

	// Extract suggested commands from response
	suggestedCommands := extractCommands(response)

	// Build response
	chatResp := models.ChatResponse{
		Message:           response,
		SuggestedCommands: suggestedCommands,
		Model:             h.client.Model,
		Timestamp:         time.Now(),
	}

	models.WriteSuccess(w, chatResp)
}

// extractCommands extracts command suggestions from LLM response
// Looks for patterns like /command or mentions of specific commands
func extractCommands(response string) []string {
	commands := []string{}
	seen := make(map[string]bool)

	// Pattern 1: Look for /command patterns
	re := regexp.MustCompile(`/[a-z]+(?:\s+[a-z]+)*`)
	matches := re.FindAllString(response, -1)
	for _, match := range matches {
		cmd := strings.TrimSpace(match)
		if !seen[cmd] {
			commands = append(commands, cmd)
			seen[cmd] = true
		}
	}

	// Pattern 2: Look for common command mentions
	commandKeywords := map[string]string{
		"dashboard":    "/dashboard",
		"assets":       "/assets",
		"liabilities":  "/liabilities",
		"net worth":    "/dashboard",
		"overview":     "/dashboard",
		"upload":       "/upload",
		"settings":     "/settings",
	}

	lowerResponse := strings.ToLower(response)
	for keyword, command := range commandKeywords {
		if strings.Contains(lowerResponse, keyword) && !seen[command] {
			commands = append(commands, command)
			seen[command] = true
		}
	}

	return commands
}

// Health checks if the LLM service is available
func (h *LLMHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.client.Ping(); err != nil {
		models.WriteError(w, http.StatusServiceUnavailable, "LLM service unavailable: "+err.Error())
		return
	}

	models.WriteSuccess(w, map[string]string{
		"status": "healthy",
		"model":  h.client.Model,
	})
}
