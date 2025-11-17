package llm

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	host := "http://localhost:11434"
	model := "dolphin3"

	client := NewClient(host, model)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.host != host {
		t.Errorf("Expected host %s, got %s", host, client.host)
	}

	if client.model != model {
		t.Errorf("Expected model %s, got %s", model, client.model)
	}

	if client.maxRetries != defaultMaxRetries {
		t.Errorf("Expected maxRetries %d, got %d", defaultMaxRetries, client.maxRetries)
	}
}

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		tip      string
		expected string
	}{
		{
			tip:      "Use git log --oneline to see a compact commit history.",
			expected: "git",
		},
		{
			tip:      "Run docker ps -a to see all containers including stopped ones.",
			expected: "docker",
		},
		{
			tip:      "Use bash aliases to create shortcuts for common commands.",
			expected: "shell",
		},
		{
			tip:      "In vim, use :set number to show line numbers.",
			expected: "vim",
		},
		{
			tip:      "Connect to a remote server with ssh user@hostname.",
			expected: "ssh",
		},
		{
			tip:      "Use python -m venv env to create a virtual environment.",
			expected: "python",
		},
		{
			tip:      "Build your Go project with go build -o myapp .",
			expected: "go",
		},
		{
			tip:      "Check systemd service status with systemctl status servicename.",
			expected: "linux",
		},
		{
			tip:      "Use the terminal command to do something useful.",
			expected: "cli",
		},
		{
			tip:      "This is a tip about something unrelated to coding.",
			expected: "general",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := extractCategory(tt.tip)
			if result != tt.expected {
				t.Errorf("Expected category %s for tip '%s', got %s", tt.expected, tt.tip, result)
			}
		})
	}
}

func TestExtractCategory_CaseInsensitive(t *testing.T) {
	tip := "USE GIT LOG --ONELINE TO SEE COMMITS"
	result := extractCategory(tip)

	if result != "git" {
		t.Errorf("Expected category 'git' for uppercase tip, got %s", result)
	}
}

func TestExtractCategory_MultipleKeywords(t *testing.T) {
	// When multiple categories match, first match should win
	tip := "Use git to commit your Docker configuration files"
	result := extractCategory(tip)

	// Should match git since it appears first in the pattern order
	if result != "git" {
		t.Errorf("Expected 'git' for multi-keyword tip, got %s", result)
	}
}

// Note: HealthCheck and GenerateTechTip tests require a running Ollama server
// These are integration tests and would typically be run separately or skipped in CI

func TestGenerateTechTip_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := NewClient("http://localhost:11434", "dolphin3")

	// Only run if Ollama is available
	if err := client.HealthCheck(); err != nil {
		t.Skipf("Ollama server not available: %v", err)
	}

	tip, category, err := client.GenerateTechTip()
	if err != nil {
		t.Fatalf("Failed to generate tech tip: %v", err)
	}

	if tip == "" {
		t.Error("Expected non-empty tip")
	}

	if category == "" {
		t.Error("Expected non-empty category")
	}

	t.Logf("Generated tip: %s (category: %s)", tip, category)
}

func TestHealthCheck_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	client := NewClient("http://localhost:11434", "dolphin3")

	err := client.HealthCheck()
	if err != nil {
		t.Skipf("Ollama server not available: %v", err)
	}
}

func TestHealthCheck_InvalidHost(t *testing.T) {
	client := NewClient("http://localhost:99999", "dolphin3")

	err := client.HealthCheck()
	if err == nil {
		t.Error("Expected error for invalid host, got nil")
	}
}
