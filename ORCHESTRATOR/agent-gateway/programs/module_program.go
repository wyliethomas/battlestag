package programs

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// ModuleProgram wraps a bash module script from SCRIPTS/modules/
type ModuleProgram struct {
	id          string
	name        string
	description string
	category    string
	parameters  []Parameter
	scriptPath  string
}

// NewModuleProgram creates a new module program
func NewModuleProgram(id, name, description, category string, params []Parameter, scriptPath string) *ModuleProgram {
	return &ModuleProgram{
		id:          id,
		name:        name,
		description: description,
		category:    category,
		parameters:  params,
		scriptPath:  scriptPath,
	}
}

// Info returns the module's metadata
func (m *ModuleProgram) Info() ProgramInfo {
	return ProgramInfo{
		ID:          m.id,
		Name:        m.name,
		Description: m.description,
		Category:    m.category,
		Parameters:  m.parameters,
	}
}

// Execute runs the module script with the given parameters
func (m *ModuleProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
	// Build command arguments from parameters
	args := make([]string, 0, len(m.parameters))

	// Add parameters in order they were defined
	for _, param := range m.parameters {
		if value, ok := params[param.Name]; ok {
			// Convert parameter value to string
			strValue := fmt.Sprintf("%v", value)
			args = append(args, strValue)
		} else if param.Required {
			// This should have been caught by Registry.Execute validation,
			// but check again just in case
			return &ExecutionResult{
				Success: false,
				Error:   fmt.Sprintf("missing required parameter: %s", param.Name),
			}, fmt.Errorf("missing required parameter: %s", param.Name)
		}
	}

	// Get absolute path to script
	absPath, err := filepath.Abs(m.scriptPath)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Error:   fmt.Sprintf("failed to resolve script path: %v", err),
		}, fmt.Errorf("failed to resolve script path: %w", err)
	}

	// Build and execute command using CommandBuilder
	cmd := NewCommand(absPath)
	for _, arg := range args {
		cmd = cmd.Arg(arg)
	}

	output, err := cmd.Execute(ctx)

	// Parse result
	if err != nil {
		// Command failed - return error details
		return &ExecutionResult{
			Success: false,
			Output:  string(output),
			Error:   fmt.Sprintf("module execution failed: %v", err),
		}, nil // Don't return error here - we have ExecutionResult with details
	}

	// Success
	return &ExecutionResult{
		Success: true,
		Output:  strings.TrimSpace(string(output)),
	}, nil
}
