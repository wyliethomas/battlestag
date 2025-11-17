package llm

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	host := "http://localhost:11434"
	model := "dolphin3"

	client := NewClient(host, model)

	if client == nil {
		t.Fatal("Expected client to be created")
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

func TestHealthCheckUnreachableServer(t *testing.T) {
	// Use an unreachable host
	client := NewClient("http://localhost:99999", "dolphin3")

	err := client.HealthCheck()
	if err == nil {
		t.Error("Expected error for unreachable server, got nil")
	}
}

func TestGenerateStoicThoughtUnreachableServer(t *testing.T) {
	// Use an unreachable host
	client := NewClient("http://localhost:99999", "dolphin3")

	// This should fail since the server is unreachable
	_, err := client.GenerateStoicThought()
	if err == nil {
		t.Error("Expected error for unreachable server, got nil")
	}
}

// Note: To test actual LLM generation, you would need a running Ollama server
// These tests are for error handling when the server is not available
