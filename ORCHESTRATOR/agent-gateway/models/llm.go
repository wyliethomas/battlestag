package models

import "time"

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat endpoint
type ChatRequest struct {
	Message string        `json:"message"` // Current user message
	History []ChatMessage `json:"history"` // Previous conversation history
}

// ChatResponse represents the data returned by the chat endpoint
type ChatResponse struct {
	Message            string    `json:"message"`             // AI response
	SuggestedCommands  []string  `json:"suggested_commands"`  // Commands the AI suggests
	Model              string    `json:"model"`               // Model used (e.g., "llama3.2:8b")
	Timestamp          time.Time `json:"timestamp"`           // Response timestamp
}
