package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChatHistory represents a conversation message
type ChatHistory struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // Message content
}

// Message represents a chat message for gateway v2
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequestV2 represents the request to the gateway v2 chat endpoint
type ChatRequestV2 struct {
	Messages []Message `json:"messages"`
}

// ChatResponseV2 represents the response from gateway v2
type ChatResponseV2 struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// ChatResponse represents the response from the LLM chat endpoint (backward compatible)
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
	// Convert history + new message to v2 format
	messages := make([]Message, 0, len(history)+1)
	for _, h := range history {
		messages = append(messages, Message{
			Role:    h.Role,
			Content: h.Content,
		})
	}
	messages = append(messages, Message{
		Role:    "user",
		Content: message,
	})

	// Build v2 request
	req := ChatRequestV2{
		Messages: messages,
	}

	// Make API request to v2 endpoint
	resp, err := c.doRequest("POST", "/chat", req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse v2 response
	var v2Resp ChatResponseV2
	if err := json.NewDecoder(resp.Body).Decode(&v2Resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for errors
	if v2Resp.Error != "" {
		return nil, fmt.Errorf("gateway error: %s", v2Resp.Error)
	}

	// Convert to old format for backward compatibility with TUI
	return &ChatResponse{
		Message:           v2Resp.Response,
		SuggestedCommands: []string{}, // v2 doesn't provide suggestions
		Model:             "",          // v2 doesn't expose model in response
		Timestamp:         time.Now(),
		ProgramID:         "",
		ProgramParams:     nil,
	}, nil
}

// HealthLLM checks if the LLM service is available (gateway v2)
func (c *Client) HealthLLM() (string, error) {
	resp, err := c.doRequest("GET", "/health", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("health check failed (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse v2 health response
	var health struct {
		Status         string `json:"status"`
		ProgramsLoaded int    `json:"programs_loaded"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return "", fmt.Errorf("failed to decode health response: %w", err)
	}

	return fmt.Sprintf("%s (%d programs loaded)", health.Status, health.ProgramsLoaded), nil
}
