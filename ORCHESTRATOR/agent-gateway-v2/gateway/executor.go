package gateway

import (
	"agent-gateway/llm"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Executor handles program execution
type Executor struct {
	timeout time.Duration
}

// NewExecutor creates a new executor with the specified timeout
func NewExecutor(timeout time.Duration) *Executor {
	return &Executor{
		timeout: timeout,
	}
}

// Execute runs a program with the provided arguments
func (e *Executor) Execute(program Program, arguments json.RawMessage) ExecutionResult {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	// Create command
	cmd := exec.CommandContext(ctx, program.ExecutablePath)

	// Prepare stdin with JSON arguments
	var stdin bytes.Buffer
	stdin.Write(arguments)
	cmd.Stdin = &stdin

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the program
	err := cmd.Run()

	// Build result
	result := ExecutionResult{
		Success: err == nil,
		Output:  stdout.String(),
	}

	if err != nil {
		// Include both error and stderr in error message
		errorMsg := err.Error()
		if stderr.Len() > 0 {
			errorMsg = fmt.Sprintf("%s: %s", errorMsg, stderr.String())
		}
		result.Error = errorMsg
	}

	return result
}

// ValidateToolCall checks if a tool call is valid (tool exists in registry)
func (e *Executor) ValidateToolCall(registry *Registry, toolCall llm.OllamaToolCall) error {
	_, exists := registry.GetProgram(toolCall.Function.Name)
	if !exists {
		return fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}
	return nil
}
