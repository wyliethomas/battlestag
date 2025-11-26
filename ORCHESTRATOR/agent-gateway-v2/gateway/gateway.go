package gateway

import (
	"agent-gateway/llm"
	"fmt"
)

// Gateway orchestrates the interaction between LLM and programs
type Gateway struct {
	registry *Registry
	executor *Executor
	llmClient *llm.OllamaClient
}

// NewGateway creates a new gateway instance
func NewGateway(registry *Registry, executor *Executor, llmClient *llm.OllamaClient) *Gateway {
	return &Gateway{
		registry:  registry,
		executor:  executor,
		llmClient: llmClient,
	}
}

// ProcessChat handles a chat interaction with tool execution loop
func (g *Gateway) ProcessChat(messages []llm.Message) (string, error) {
	// Get available tools from registry
	tools := g.registry.ToOllamaTools()
	fmt.Printf("DEBUG: Sending %d tools to LLM\n", len(tools))

	// Conversation loop
	conversationMessages := make([]llm.Message, len(messages))
	copy(conversationMessages, messages)

	maxIterations := 10 // Prevent infinite loops
	for i := 0; i < maxIterations; i++ {
		// Send to LLM with tools
		response, err := g.llmClient.Chat(conversationMessages, tools)
		if err != nil {
			return "", fmt.Errorf("llm error: %w", err)
		}

		fmt.Printf("DEBUG: Received response with %d tool_calls\n", len(response.Message.ToolCalls))
		fmt.Printf("DEBUG: Response content: %s\n", response.Message.Content)

		// Check if there are tool calls
		if len(response.Message.ToolCalls) > 0 {
			fmt.Printf("DEBUG: Processing %d tool calls\n", len(response.Message.ToolCalls))
			// Validate and execute tool calls
			allValid := true
			for _, toolCall := range response.Message.ToolCalls {
				fmt.Printf("DEBUG: Validating tool call: %s\n", toolCall.Function.Name)
				// Validate tool exists
				if err := g.executor.ValidateToolCall(g.registry, toolCall); err != nil {
					// Tool doesn't exist - this is a hallucination
					fmt.Printf("DEBUG: Tool validation FAILED: %v\n", err)
					allValid = false
					break
				}
				fmt.Printf("DEBUG: Tool validation OK\n")
			}

			// If any tool is invalid, retry without tools (graceful fallback)
			if !allValid {
				fmt.Printf("DEBUG: Tool validation failed, falling back without tools\n")
				response, err = g.llmClient.ChatWithoutTools(conversationMessages)
				if err != nil {
					return "", fmt.Errorf("llm error on fallback: %w", err)
				}
				return response.Message.Content, nil
			}
			fmt.Printf("DEBUG: All tools valid, proceeding with execution\n")

			// All tools valid - execute them
			// Add assistant message with tool calls to conversation
			conversationMessages = append(conversationMessages, response.Message)

				// Execute each tool and add results
			for _, toolCall := range response.Message.ToolCalls {
				program, _ := g.registry.GetProgram(toolCall.Function.Name)
				fmt.Printf("DEBUG: Executing tool: %s with args: %s\n", toolCall.Function.Name, string(toolCall.Function.Arguments))
				result := g.executor.Execute(program, toolCall.Function.Arguments)
				fmt.Printf("DEBUG: Tool result - Success: %v, Output: %s, Error: %s\n", result.Success, result.Output, result.Error)

				// Add tool result as a message
				toolResultMsg := llm.Message{
					Role:    "tool",
					Content: g.formatToolResult(result),
				}
				fmt.Printf("DEBUG: Adding tool result message: %s\n", toolResultMsg.Content)
				conversationMessages = append(conversationMessages, toolResultMsg)
			}

			// Continue loop to get LLM's response to tool results
			continue
		}

		// No tool calls - return the assistant's message
		return response.Message.Content, nil
	}

	return "", fmt.Errorf("exceeded maximum iterations (%d)", maxIterations)
}

// formatToolResult converts an ExecutionResult to a string for the LLM
func (g *Gateway) formatToolResult(result ExecutionResult) string {
	if result.Success {
		return result.Output
	}
	return fmt.Sprintf("Error: %s", result.Error)
}

// GetRegistry returns the gateway's registry
func (g *Gateway) GetRegistry() *Registry {
	return g.registry
}
