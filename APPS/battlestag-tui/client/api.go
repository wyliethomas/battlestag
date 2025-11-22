package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/battlestag/battlestag-tui/models"
)

// Client represents the API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new API client from environment variables
// Deprecated: Use NewClientWithConfig instead
func NewClient() *Client {
	baseURL := os.Getenv("AGENT_GATEWAY_URL")
	if baseURL == "" {
		baseURL = "http://192.168.1.140:8080"
	}

	apiKey := os.Getenv("AGENT_GATEWAY_API_KEY")
	if apiKey == "" {
		apiKey = "test-api-key-12345" // Default matching server config at /etc/agent-gateway/config.yaml
	}

	return NewClientWithConfig(baseURL, apiKey)
}

// NewClientWithConfig creates a new API client with explicit URL and API key
func NewClientWithConfig(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.HTTPClient.Do(req)
}

// parseResponse parses the API response
func parseResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("API error: %s", apiResp.Error)
	}

	// Re-marshal and unmarshal to convert data to target type
	dataBytes, err := json.Marshal(apiResp.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	if err := json.Unmarshal(dataBytes, target); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// Health checks the API health
func (c *Client) Health() (*models.HealthResponse, error) {
	resp, err := c.doRequest("GET", "/api/health", nil)
	if err != nil {
		return nil, err
	}

	var health models.HealthResponse
	if err := parseResponse(resp, &health); err != nil {
		return nil, err
	}

	return &health, nil
}

// Stats retrieves system statistics
func (c *Client) Stats() (*models.StatsResponse, error) {
	resp, err := c.doRequest("GET", "/api/stats", nil)
	if err != nil {
		return nil, err
	}

	var stats models.StatsResponse
	if err := parseResponse(resp, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetFinancialOverview retrieves the complete financial overview
func (c *Client) GetFinancialOverview() (*models.FinancialOverview, error) {
	resp, err := c.doRequest("GET", "/api/financial/net-worth", nil)
	if err != nil {
		return nil, err
	}

	var overview models.FinancialOverview
	if err := parseResponse(resp, &overview); err != nil {
		return nil, err
	}

	return &overview, nil
}

// GetTodayStoicThought retrieves today's stoic thought
func (c *Client) GetTodayStoicThought() (*models.StoicThought, error) {
	resp, err := c.doRequest("GET", "/api/stoic/today", nil)
	if err != nil {
		return nil, err
	}

	var thought models.StoicThought
	if err := parseResponse(resp, &thought); err != nil {
		return nil, err
	}

	return &thought, nil
}

// ListAssets retrieves all assets
func (c *Client) ListAssets() ([]models.Asset, error) {
	resp, err := c.doRequest("GET", "/api/financial-asset", nil)
	if err != nil {
		return nil, err
	}

	var assets []models.Asset
	if err := parseResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}

// CreateAsset creates a new asset
func (c *Client) CreateAsset(asset models.Asset) error {
	resp, err := c.doRequest("POST", "/api/financial-asset", asset)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create asset (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateAsset updates an existing asset
func (c *Client) UpdateAsset(name string, asset models.Asset) error {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/financial-asset/%s", name), asset)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update asset (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteAsset deletes an asset
func (c *Client) DeleteAsset(name string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/financial-asset/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete asset (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// ListLiabilities retrieves all liabilities
func (c *Client) ListLiabilities() ([]models.Liability, error) {
	resp, err := c.doRequest("GET", "/api/financial-liability", nil)
	if err != nil {
		return nil, err
	}

	var liabilities []models.Liability
	if err := parseResponse(resp, &liabilities); err != nil {
		return nil, err
	}

	return liabilities, nil
}

// CreateLiability creates a new liability
func (c *Client) CreateLiability(liability models.Liability) error {
	resp, err := c.doRequest("POST", "/api/financial-liability", liability)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create liability (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateLiability updates an existing liability
func (c *Client) UpdateLiability(name string, liability models.Liability) error {
	resp, err := c.doRequest("PUT", fmt.Sprintf("/api/financial-liability/%s", name), liability)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update liability (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteLiability deletes a liability
func (c *Client) DeleteLiability(name string) error {
	resp, err := c.doRequest("DELETE", fmt.Sprintf("/api/financial-liability/%s", name), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete liability (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UploadStatement uploads a PDF bank statement for processing
func (c *Client) UploadStatement(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/financial-statement/process", body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
