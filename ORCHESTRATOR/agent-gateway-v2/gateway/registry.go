package gateway

import (
	"agent-gateway/llm"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Registry manages the collection of available programs
type Registry struct {
	programs map[string]Program
}

// NewRegistry creates a new registry by scanning the programs directory
func NewRegistry(programsDir string) (*Registry, error) {
	registry := &Registry{
		programs: make(map[string]Program),
	}

	// Walk the programs directory
	entries, err := os.ReadDir(programsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read programs directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		programDir := filepath.Join(programsDir, entry.Name())
		manifestPath := filepath.Join(programDir, "manifest.json")

		// Check if manifest.json exists
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			continue // Skip directories without manifest
		}

		// Load and parse manifest
		manifestData, err := os.ReadFile(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read manifest for %s: %w", entry.Name(), err)
		}

		var manifest Manifest
		if err := json.Unmarshal(manifestData, &manifest); err != nil {
			return nil, fmt.Errorf("failed to parse manifest for %s: %w", entry.Name(), err)
		}

		// Determine executable path (use manifest.executable or default to manifest.name)
		executableName := manifest.Executable
		if executableName == "" {
			executableName = manifest.Name
		}
		executablePath := filepath.Join(programDir, executableName)

		// Verify executable exists
		if _, err := os.Stat(executablePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("executable not found for %s: %s", manifest.Name, executablePath)
		}

		// Register the program
		registry.programs[manifest.Name] = Program{
			Manifest:       manifest,
			ExecutablePath: executablePath,
		}
	}

	return registry, nil
}

// GetProgram retrieves a program by name
func (r *Registry) GetProgram(name string) (Program, bool) {
	prog, ok := r.programs[name]
	return prog, ok
}

// GetAllPrograms returns all registered programs
func (r *Registry) GetAllPrograms() map[string]Program {
	return r.programs
}

// ToOllamaTools converts all registered programs to Ollama tool format
func (r *Registry) ToOllamaTools() []llm.OllamaTool {
	tools := make([]llm.OllamaTool, 0, len(r.programs))

	for _, program := range r.programs {
		tool := llm.OllamaTool{
			Type: "function",
		}
		tool.Function.Name = program.Manifest.Name
		tool.Function.Description = program.Manifest.Description
		tool.Function.Parameters = program.Manifest.Parameters

		tools = append(tools, tool)
	}

	return tools
}

// Count returns the number of registered programs
func (r *Registry) Count() int {
	return len(r.programs)
}
