package writers

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
)

// ObsidianWriter writes tasks to an Obsidian vault
type ObsidianWriter struct {
	vaultPath string
	tasksFile string
}

// NewObsidianWriter creates a new Obsidian writer
func NewObsidianWriter(cfg config.ObsidianConfig) (*ObsidianWriter, error) {
	// Validate paths
	if _, err := os.Stat(cfg.VaultPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("vault path does not exist: %s", cfg.VaultPath)
	}

	tasksPath := filepath.Join(cfg.VaultPath, cfg.TasksFile)

	// Ensure tasks file directory exists
	tasksDir := filepath.Dir(tasksPath)
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	// Create tasks file if it doesn't exist
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		f, err := os.Create(tasksPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create tasks file: %w", err)
		}
		f.WriteString("# Action Items\n\n")
		f.Close()
	}

	return &ObsidianWriter{
		vaultPath: cfg.VaultPath,
		tasksFile: tasksPath,
	}, nil
}

// WriteTask appends a task to the Obsidian tasks file
func (w *ObsidianWriter) WriteTask(task messages.Task) error {
	// Check if task already exists
	if w.TaskExists(task.Permalink) {
		return nil // Skip duplicate
	}

	// Format task entry
	entry := w.formatTask(task)

	// Append to file
	f, err := os.OpenFile(w.tasksFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open tasks file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry + "\n"); err != nil {
		return fmt.Errorf("failed to write task: %w", err)
	}

	return nil
}

// TaskExists checks if a task with the given permalink already exists
func (w *ObsidianWriter) TaskExists(permalink string) bool {
	f, err := os.Open(w.tasksFile)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, permalink) {
			return true
		}
	}

	return false
}

// formatTask formats a task for Obsidian markdown
func (w *ObsidianWriter) formatTask(task messages.Task) string {
	var sb strings.Builder

	// Task checkbox with description
	sb.WriteString(fmt.Sprintf("- [ ] %s\n", task.Description))

	// Metadata line
	date := task.Timestamp.Format("2006-01-02")
	sb.WriteString(fmt.Sprintf("  - **%s**: %s | @%s | %s\n",
		strings.Title(task.Source),
		task.Channel,
		task.Author,
		date))

	// Context (if present)
	if task.Context != "" {
		// Indent context for readability
		contextLines := strings.Split(task.Context, "\n")
		for _, line := range contextLines {
			if strings.TrimSpace(line) != "" {
				sb.WriteString(fmt.Sprintf("  - %s\n", line))
			}
		}
	}

	// Permalink
	sb.WriteString(fmt.Sprintf("  - [View](%s)\n", task.Permalink))

	return sb.String()
}
