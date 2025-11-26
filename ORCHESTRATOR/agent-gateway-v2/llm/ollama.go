package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// OllamaClient handles communication with Ollama API
type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		client:  &http.Client{},
	}
}

// ChatRequest represents a request to Ollama's chat API
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Tools    []OllamaTool  `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

// ChatResponse represents a response from Ollama's chat API
type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
}

// Chat sends a chat request to Ollama
func (c *OllamaClient) Chat(messages []Message, tools []OllamaTool) (*ChatResponse, error) {
	request := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request
	url := fmt.Sprintf("%s/api/chat", c.baseURL)
	resp, err := c.client.Post(url, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// ChatWithoutTools sends a chat request without tool support (fallback mode)
func (c *OllamaClient) ChatWithoutTools(messages []Message) (*ChatResponse, error) {
	return c.Chat(messages, nil)
}
