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
	userName  string
}

// NewObsidianWriter creates a new Obsidian writer
func NewObsidianWriter(cfg config.ObsidianConfig, userCfg config.UserConfig) (*ObsidianWriter, error) {
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
		f.WriteString("# Action Items\n\n## My Tasks\n\n## Following Up On\n\n")
		f.Close()
	}

	return &ObsidianWriter{
		vaultPath: cfg.VaultPath,
		tasksFile: tasksPath,
		userName:  userCfg.Name,
	}, nil
}

// WriteTask adds a task to the appropriate section based on assignee
func (w *ObsidianWriter) WriteTask(task messages.Task) error {
	// Check if task already exists
	if w.TaskExists(task.Permalink) {
		return nil // Skip duplicate
	}

	// Read entire file
	content, err := os.ReadFile(w.tasksFile)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Determine which section this task belongs to
	isMyTask := w.isMyTask(task.Assignee)
	targetSection := "## My Tasks"
	if !isMyTask {
		targetSection = "## Following Up On"
	}

	// Find the target section and insert the task
	newLines := []string{}
	inTargetSection := false
	taskInserted := false

	for i, line := range lines {
		newLines = append(newLines, line)

		// Check if we're entering the target section
		if strings.HasPrefix(line, targetSection) {
			inTargetSection = true
			continue
		}

		// Check if we're leaving the target section (next section header or end of file)
		if inTargetSection && (strings.HasPrefix(line, "##") || i == len(lines)-1) && !taskInserted {
			// Insert task before this line
			taskEntry := w.formatTask(task)

			// Insert before the current line if it's a section header
			if strings.HasPrefix(line, "##") {
				newLines = newLines[:len(newLines)-1] // Remove the just-added section header
				newLines = append(newLines, taskEntry)
				newLines = append(newLines, line) // Re-add the section header
			} else {
				newLines = append(newLines, taskEntry)
			}

			taskInserted = true
			inTargetSection = false
		}
	}

	// If we didn't insert the task (shouldn't happen with proper section headers), append at end
	if !taskInserted {
		newLines = append(newLines, w.formatTask(task))
	}

	// Write back to file
	newContent := strings.Join(newLines, "\n")
	if err := os.WriteFile(w.tasksFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// isMyTask determines if a task is assigned to the user
func (w *ObsidianWriter) isMyTask(assignee string) bool {
	// Normalize for comparison
	assignee = strings.ToLower(strings.TrimSpace(assignee))
	userName := strings.ToLower(strings.TrimSpace(w.userName))

	// Check if assignee matches user name
	if assignee == userName {
		return true
	}

	// Check if assignee is empty or "unassigned" - treat as user's task by default
	if assignee == "" || assignee == "unassigned" {
		return true
	}

	return false
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
	// If not user's task, prefix with assignee
	if !w.isMyTask(task.Assignee) && task.Assignee != "" {
		sb.WriteString(fmt.Sprintf("- [ ] **@%s** - %s\n", task.Assignee, task.Description))
	} else {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", task.Description))
	}

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
