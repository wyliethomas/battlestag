package programs

import (
	"context"
	"fmt"
)

// EchoProgram is a simple test program that echoes back input
type EchoProgram struct{}

// NewEchoProgram creates a new echo program
func NewEchoProgram() *EchoProgram {
	return &EchoProgram{}
}

// Info returns the program's metadata
func (p *EchoProgram) Info() ProgramInfo {
	return ProgramInfo{
		ID:          "echo",
		Name:        "Echo Test",
		Description: "A simple test program that echoes back the provided message",
		Category:    "test",
		Parameters: []Parameter{
			{
				Name:        "message",
				Type:        "string",
				Description: "The message to echo back",
				Required:    true,
			},
			{
				Name:        "repeat",
				Type:        "int",
				Description: "Number of times to repeat the message (default: 1)",
				Required:    false,
			},
		},
	}
}

// Execute runs the echo program
func (p *EchoProgram) Execute(ctx context.Context, params map[string]interface{}) (*ExecutionResult, error) {
	// Get message parameter
	message, ok := params["message"].(string)
	if !ok {
		return &ExecutionResult{
			Success: false,
			Error:   "message parameter must be a string",
		}, fmt.Errorf("invalid message parameter")
	}

	// Get optional repeat parameter
	repeat := 1
	if repeatVal, ok := params["repeat"]; ok {
		switch v := repeatVal.(type) {
		case float64:
			repeat = int(v)
		case int:
			repeat = v
		}
	}

	// Build output
	output := ""
	for i := 0; i < repeat; i++ {
		output += fmt.Sprintf("%d. %s\n", i+1, message)
	}

	return &ExecutionResult{
		Success: true,
		Output:  output,
	}, nil
}
