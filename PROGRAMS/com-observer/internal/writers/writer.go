package writers

import (
	"fmt"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
)

// Writer is the interface for writing tasks to note-taking systems
type Writer interface {
	// WriteTask writes a task to the note system
	WriteTask(task messages.Task) error

	// TaskExists checks if a task with the given permalink already exists
	TaskExists(permalink string) bool
}

// NewWriter creates a writer based on the configuration
func NewWriter(cfg *config.Config) (Writer, error) {
	switch cfg.NoteWriter.Type {
	case "obsidian":
		return NewObsidianWriter(cfg.NoteWriter.Obsidian)
	case "notion":
		return NewNotionWriter(cfg.NoteWriter.Notion)
	default:
		return nil, fmt.Errorf("unsupported note writer type: %s", cfg.NoteWriter.Type)
	}
}
