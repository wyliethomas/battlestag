package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaClient handles communication with Ollama LLM
type OllamaClient struct {
	host   string
	model  string
	client *http.Client
}

// OllamaRequest represents the request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}

// OllamaResponse represents the response from Ollama API
type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Response  string `json:"response"`
	Done      bool   `json:"done"`
}

// NewOllamaClient creates a new Ollama LLM client
func NewOllamaClient(host, model string) *OllamaClient {
	return &OllamaClient{
		host:  host,
		model: model,
		client: &http.Client{
			Timeout: 120 * time.Second, // LLM calls can take a while
		},
	}
}

// ParseStatementText sends extracted PDF text to LLM for structured parsing
func (c *OllamaClient) ParseStatementText(text string) (string, error) {
	prompt := buildStatementPrompt(text)

	reqBody := OllamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
		Format: "json",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", c.host)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return "", fmt.Errorf("unmarshal response: %w", err)
	}

	return ollamaResp.Response, nil
}

// HealthCheck verifies Ollama is accessible
func (c *OllamaClient) HealthCheck() error {
	url := fmt.Sprintf("%s/api/tags", c.host)
	resp, err := c.client.Get(url)
	if err != nil {
		return fmt.Errorf("ollama not accessible: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}

	return nil
}

// buildStatementPrompt creates the LLM prompt for parsing bank statements
func buildStatementPrompt(text string) string {
	return fmt.Sprintf(`You are a financial data extraction assistant. Extract transaction information from the following bank statement text and return ONLY valid JSON (no markdown, no code blocks, no explanations).

Required JSON structure:
{
  "account_name": "Account holder name or account type",
  "account_last4": "Last 4 digits of account number",
  "statement_date": "YYYY-MM-DD format",
  "transactions": [
    {
      "transaction_date": "YYYY-MM-DD",
      "post_date": "YYYY-MM-DD or null",
      "description": "Transaction description",
      "amount": -50.25,
      "transaction_type": "debit",
      "balance": 1234.56
    }
  ]
}

Rules:
- transaction_type must be "debit" or "credit"
- amount should be negative for debits (money out), positive for credits (money in)
- dates must be in YYYY-MM-DD format
- balance can be null if not shown
- post_date can be null if not shown
- Extract ALL transactions you can find

Bank statement text:
%s

Return ONLY the JSON object, nothing else.`, text)
}
