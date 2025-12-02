package messages

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Queue represents a filesystem-based message queue
type Queue struct {
	path string
}

// NewQueue creates a new filesystem queue at the specified path
func NewQueue(path string) (*Queue, error) {
	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}

	return &Queue{path: path}, nil
}

// Push adds a message to the queue
func (q *Queue) Push(msg Message) error {
	// Generate filename: timestamp_id.json
	filename := fmt.Sprintf("%d_%s.json", time.Now().UnixNano(), msg.ID)
	filepath := filepath.Join(q.path, filename)

	// Marshal message to JSON
	data, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write message to queue: %w", err)
	}

	return nil
}

// Pop removes and returns up to N messages from the queue
func (q *Queue) Pop(batch int) ([]Message, error) {
	messages, files, err := q.readMessages(batch)
	if err != nil {
		return nil, err
	}

	// Delete the files
	for _, file := range files {
		if err := os.Remove(filepath.Join(q.path, file)); err != nil {
			return nil, fmt.Errorf("failed to remove message file %s: %w", file, err)
		}
	}

	return messages, nil
}

// Peek returns up to N messages from the queue without removing them
func (q *Queue) Peek(batch int) ([]Message, error) {
	messages, _, err := q.readMessages(batch)
	return messages, err
}

// Count returns the number of messages in the queue
func (q *Queue) Count() (int, error) {
	files, err := os.ReadDir(q.path)
	if err != nil {
		return 0, fmt.Errorf("failed to read queue directory: %w", err)
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			count++
		}
	}

	return count, nil
}

// readMessages reads up to N messages from the queue
// Returns the messages and their filenames
func (q *Queue) readMessages(batch int) ([]Message, []string, error) {
	// List all files in queue directory
	files, err := os.ReadDir(q.path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read queue directory: %w", err)
	}

	// Filter to only .json files and sort by name (timestamp-based)
	var jsonFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			jsonFiles = append(jsonFiles, file.Name())
		}
	}
	sort.Strings(jsonFiles)

	// Limit to batch size
	if len(jsonFiles) > batch {
		jsonFiles = jsonFiles[:batch]
	}

	// Read and parse messages
	messages := make([]Message, 0, len(jsonFiles))
	for _, filename := range jsonFiles {
		data, err := os.ReadFile(filepath.Join(q.path, filename))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read message file %s: %w", filename, err)
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal message from %s: %w", filename, err)
		}

		messages = append(messages, msg)
	}

	return messages, jsonFiles, nil
}
