package messages

import "time"

// Message represents a standardized message from any communication platform
type Message struct {
	ID        string            `json:"id"`
	Source    string            `json:"source"`    // "slack" | "gmail" | "trello"
	Channel   string            `json:"channel"`   // Channel name, email subject, or board name
	Author    string            `json:"author"`    // Author name or email
	Content   string            `json:"content"`   // Message text content
	Timestamp time.Time         `json:"timestamp"` // When message was sent
	ThreadID  string            `json:"thread_id,omitempty"`
	Permalink string            `json:"permalink,omitempty"` // Link back to original message
	Metadata  map[string]string `json:"metadata,omitempty"`  // Additional source-specific data
}

// Task represents an extracted action item from messages
type Task struct {
	Description string    `json:"description"` // What needs to be done
	Context     string    `json:"context"`     // Additional context from the conversation
	Assignee    string    `json:"assignee"`    // Who is being asked to do this task
	Source      string    `json:"source"`      // "slack" | "gmail" | "trello"
	Channel     string    `json:"channel"`     // Where it came from
	Author      string    `json:"author"`      // Who mentioned it
	Timestamp   time.Time `json:"timestamp"`   // When it was mentioned
	Permalink   string    `json:"permalink"`   // Link back to source
}
