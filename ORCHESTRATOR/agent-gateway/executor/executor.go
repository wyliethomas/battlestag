package executor

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"agent-gateway/models"
)

// Executor runs agent programs and parses their output
type Executor struct {
	stoicPath string
	techPath  string
}

// NewExecutor creates a new executor
func NewExecutor(stoicPath, techPath string) *Executor {
	return &Executor{
		stoicPath: stoicPath,
		techPath:  techPath,
	}
}

// StoicOutput represents the JSON output from stoic program
type StoicOutput struct {
	Date      string `json:"date"`
	Thought   string `json:"thought"`
	CreatedAt string `json:"created_at"`
}

// TechOutput represents the JSON output from tech program
type TechOutput struct {
	Date      string `json:"date"`
	Title     string `json:"title"`
	Tip       string `json:"tip"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// GetStoicToday gets today's stoic thought
func (e *Executor) GetStoicToday() (*models.StoicThought, error) {
	return e.GetStoicByDate("")
}

// GetStoicByDate gets a stoic thought for a specific date (empty string for today)
func (e *Executor) GetStoicByDate(date string) (*models.StoicThought, error) {
	args := []string{"--json"}
	if date != "" {
		args = append(args, "--date", date)
	}

	cmd := exec.Command(e.stoicPath, args...)
	// Set working directory to the program's directory so it can find its database
	cmd.Dir = filepath.Dir(e.stoicPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute stoic program: %w (output: %s)", err, string(output))
	}

	var stoicOut StoicOutput
	if err := json.Unmarshal(output, &stoicOut); err != nil {
		return nil, fmt.Errorf("failed to parse stoic output: %w (output: %s)", err, string(output))
	}

	createdAt, err := time.Parse(time.RFC3339, stoicOut.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
	}

	// Clean up thought - remove surrounding quotes if present
	thought := strings.Trim(stoicOut.Thought, "\"")

	return &models.StoicThought{
		Date:      stoicOut.Date,
		Thought:   thought,
		CreatedAt: createdAt,
		Source:    "Stoic Agent", // Default source since program doesn't provide it
	}, nil
}

// GetTechToday gets today's tech tip
func (e *Executor) GetTechToday() (*models.TechTip, error) {
	return e.GetTechByDate("")
}

// GetTechByDate gets a tech tip for a specific date (empty string for today)
func (e *Executor) GetTechByDate(date string) (*models.TechTip, error) {
	args := []string{"--json"}
	if date != "" {
		args = append(args, "--date", date)
	}

	cmd := exec.Command(e.techPath, args...)
	// Set working directory to the program's directory so it can find its database
	cmd.Dir = filepath.Dir(e.techPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute tech program: %w (output: %s)", err, string(output))
	}

	var techOut TechOutput
	if err := json.Unmarshal(output, &techOut); err != nil {
		return nil, fmt.Errorf("failed to parse tech output: %w (output: %s)", err, string(output))
	}

	createdAt, err := time.Parse(time.RFC3339, techOut.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at timestamp: %w", err)
	}

	return &models.TechTip{
		Date:      techOut.Date,
		Title:     techOut.Title,
		Tip:       techOut.Tip,
		Category:  techOut.Category,
		CreatedAt: createdAt,
	}, nil
}
