package programs

import (
	"context"
	"fmt"
)

// Parameter defines a single parameter for a program
type Parameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "string", "int", "bool", etc.
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ProgramInfo describes a program's metadata
type ProgramInfo struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category"` // "data", "tasks", "servers", "files"
	Parameters  []Parameter `json:"parameters"`
}

// ExecutionResult is returned by a program execution
type ExecutionResult struct {
	Success bool   `json:"success"`
	Output  string `json:"output"` // Formatted text output
	Error   string `json:"error,omitempty"`
}

// Program interface that all programs must implement
type Program interface {
	// Info returns the program's metadata
	Info() ProgramInfo

	// Execute runs the program with the given parameters
	Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error)
}

// Registry holds all available programs
type Registry struct {
	programs map[string]Program
}

// NewRegistry creates a new program registry
func NewRegistry() *Registry {
	return &Registry{
		programs: make(map[string]Program),
	}
}

// Register adds a program to the registry
func (r *Registry) Register(program Program) error {
	info := program.Info()
	if _, exists := r.programs[info.ID]; exists {
		return fmt.Errorf("program %s already registered", info.ID)
	}
	r.programs[info.ID] = program
	return nil
}

// Get retrieves a program by ID
func (r *Registry) Get(id string) (Program, bool) {
	program, exists := r.programs[id]
	return program, exists
}

// List returns information about all registered programs
func (r *Registry) List() []ProgramInfo {
	infos := make([]ProgramInfo, 0, len(r.programs))
	for _, program := range r.programs {
		infos = append(infos, program.Info())
	}
	return infos
}

// Execute runs a program with the given parameters
func (r *Registry) Execute(ctx context.Context, id string, params map[string]interface{}) (*ExecutionResult, error) {
	program, exists := r.Get(id)
	if !exists {
		return nil, fmt.Errorf("program not found: %s", id)
	}

	// Validate required parameters
	info := program.Info()
	for _, param := range info.Parameters {
		if param.Required {
			if _, ok := params[param.Name]; !ok {
				return &ExecutionResult{
					Success: false,
					Error:   fmt.Sprintf("missing required parameter: %s", param.Name),
				}, fmt.Errorf("missing required parameter: %s", param.Name)
			}
		}
	}

	return program.Execute(ctx, params)
}
