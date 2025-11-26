package llm

import "encoding/json"

// OllamaTool represents a tool in Ollama's function calling format
type OllamaTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

// OllamaToolCall represents a tool call from Ollama
type OllamaToolCall struct {
	Function struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	} `json:"function"`
}

// Message represents a chat message
type Message struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}
