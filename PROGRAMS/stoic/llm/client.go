package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultMaxRetries = 3
	defaultTimeout    = 30 * time.Second
	retryDelay        = 20 * time.Second // Wait 20s for Ollama server to wake up
)

type Client struct {
	host       string
	model      string
	httpClient *http.Client
	maxRetries int
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
	CreatedAt string `json:"created_at"`
}

// NewClient creates a new Ollama LLM client
func NewClient(host, model string) *Client {
	return &Client{
		host:  host,
		model: model,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
		maxRetries: defaultMaxRetries,
	}
}

// GenerateStoicThought generates a stoic thought using the LLM with retry logic
func (c *Client) GenerateStoicThought() (string, error) {
	prompt := "Generate a single stoic thought for the day. Keep it under 3 sentences. Focus on practical wisdom from Marcus Aurelius, Epictetus, or Seneca. Be direct and actionable."

	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		thought, err := c.generate(prompt)
		if err == nil {
			return thought, nil
		}

		lastErr = err
		if attempt < c.maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastErr)
}

// generate makes a request to the Ollama API to generate text
func (c *Client) generate(prompt string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/generate", c.host)

	reqBody := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var genResp generateResponse
	if err := json.Unmarshal(body, &genResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if genResp.Response == "" {
		return "", fmt.Errorf("empty response from LLM")
	}

	return genResp.Response, nil
}

// HealthCheck verifies that the Ollama server is reachable with retry logic
func (c *Client) HealthCheck() error {
	var lastErr error
	for attempt := 1; attempt <= c.maxRetries; attempt++ {
		if attempt > 1 {
			fmt.Printf("Waiting %v for Ollama server to wake up (attempt %d/%d)...\n", retryDelay, attempt, c.maxRetries)
		}

		err := c.healthCheckOnce()
		if err == nil {
			return nil
		}

		lastErr = err
		if attempt < c.maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return fmt.Errorf("health check failed after %d attempts: %w", c.maxRetries, lastErr)
}

// healthCheckOnce performs a single health check attempt
func (c *Client) healthCheckOnce() error {
	endpoint := fmt.Sprintf("%s/api/tags", c.host)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama server unreachable at %s: %w", c.host, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama server returned status %d", resp.StatusCode)
	}

	return nil
}
