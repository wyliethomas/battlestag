package gateway

import "encoding/json"

// Manifest represents a program's manifest.json file
type Manifest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Executable  string          `json:"executable,omitempty"` // Optional, defaults to name
	Parameters  json.RawMessage `json:"parameters"`           // JSON Schema
}

// Program is a loaded program ready for execution
type Program struct {
	Manifest       Manifest
	ExecutablePath string
}

// ExecutionResult represents the result of executing a program
type ExecutionResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}
