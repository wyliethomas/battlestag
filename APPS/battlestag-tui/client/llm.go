package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"battlestag/models"
)

// ChatHistory represents a conversation message
type ChatHistory struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // Message content
}

// ChatRequest represents the request to the LLM chat endpoint
type ChatRequest struct {
	Message string        `json:"message"`
	History []ChatHistory `json:"history"`
}

// ChatResponse represents the response from the LLM chat endpoint
type ChatResponse struct {
	Message           string                 `json:"message"`
	SuggestedCommands []string               `json:"suggested_commands"`
	Model             string                 `json:"model"`
	Timestamp         time.Time              `json:"timestamp"`
	ProgramID         string                 `json:"program_id,omitempty"`
	ProgramParams     map[string]interface{} `json:"program_params,omitempty"`
}

// Chat sends a message to the LLM and returns the response
func (c *Client) Chat(message string, history []ChatHistory) (*ChatResponse, error) {
	// Build request
	req := ChatRequest{
		Message: message,
		History: history,
	}

	// Make API request
	resp, err := c.doRequest("POST", "/api/llm/chat", req)
	if err != nil {
		return nil, err
	}

	// Parse response
	var chatResp ChatResponse
	if err := parseResponse(resp, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}

// HealthLLM checks if the LLM service is available
func (c *Client) HealthLLM() (string, error) {
	resp, err := c.doRequest("GET", "/api/llm/health", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM health check failed (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("failed to decode health response: %w", err)
	}

	if !apiResp.Success {
		return "", fmt.Errorf("LLM health check failed: %s", apiResp.Error)
	}

	// Extract status and model from response data
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}

	status, _ := data["status"].(string)
	model, _ := data["model"].(string)

	return fmt.Sprintf("%s (model: %s)", status, model), nil
}
