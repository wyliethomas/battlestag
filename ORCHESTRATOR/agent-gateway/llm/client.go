package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client represents an LLM client for Ollama
type Client struct {
	Endpoint      string
	Model         string
	SystemPrompt  string
	Timeout       time.Duration
	HTTPClient    *http.Client
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a request to Ollama's chat API
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// ChatResponse represents a response from Ollama's chat API
type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// NewClient creates a new LLM client
func NewClient(endpoint, model, systemPrompt string, timeout time.Duration) *Client {
	return &Client{
		Endpoint:     endpoint,
		Model:        model,
		SystemPrompt: systemPrompt,
		Timeout:      timeout,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Chat sends a chat message and returns the response
func (c *Client) Chat(userMessage string, history []Message) (string, error) {
	// Build messages array
	messages := []Message{}

	// Add system prompt if configured
	if c.SystemPrompt != "" {
		messages = append(messages, Message{
			Role:    "system",
			Content: c.SystemPrompt,
		})
	}

	// Add conversation history
	messages = append(messages, history...)

	// Add current user message
	messages = append(messages, Message{
		Role:    "user",
		Content: userMessage,
	})

	// Build request
	reqBody := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Stream:   false, // Non-streaming for now
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	req, err := http.NewRequest("POST", c.Endpoint+"/api/chat", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return chatResp.Message.Content, nil
}

// Ping checks if the Ollama server is available
func (c *Client) Ping() error {
	req, err := http.NewRequest("GET", c.Endpoint+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama server unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server returned status %d", resp.StatusCode)
	}

	return nil
}
