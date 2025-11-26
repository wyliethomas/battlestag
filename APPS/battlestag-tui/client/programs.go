package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// ProgramParameter defines a single parameter for a program
type ProgramParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ProgramInfo describes a program's metadata
type ProgramInfo struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Parameters  []ProgramParameter `json:"parameters"`
}

// ProgramExecutionResult is returned by a program execution
type ProgramExecutionResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

// ListPrograms fetches all available programs from the gateway
func (c *Client) ListPrograms() ([]ProgramInfo, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/api/programs/list", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Gateway v2 doesn't require authentication

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response struct {
		Success bool          `json:"success"`
		Data    []ProgramInfo `json:"data"`
		Error   string        `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Error)
	}

	return response.Data, nil
}

// ExecuteProgram executes a program with the given parameters
func (c *Client) ExecuteProgram(programID string, parameters map[string]interface{}) (*ProgramExecutionResult, error) {
	requestBody := map[string]interface{}{
		"program_id": programID,
		"parameters": parameters,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/programs/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Gateway v2 doesn't require authentication
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var response struct {
		Success bool                    `json:"success"`
		Data    *ProgramExecutionResult `json:"data"`
		Error   string                  `json:"error,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Error)
	}

	return response.Data, nil
}
