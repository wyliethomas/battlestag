package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("API_KEY", "test-key")
	os.Setenv("PORT", "9000")
	os.Setenv("HOST", "127.0.0.1")
	os.Setenv("STOIC_DB_PATH", "/tmp/stoic.db")
	os.Setenv("TECH_DB_PATH", "/tmp/tech.db")

	defer func() {
		// Clean up
		os.Unsetenv("API_KEY")
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("STOIC_DB_PATH")
		os.Unsetenv("TECH_DB_PATH")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Auth.APIKey != "test-key" {
		t.Errorf("API key not loaded from env: got %v want test-key", cfg.Auth.APIKey)
	}

	if cfg.Server.Port != "9000" {
		t.Errorf("Port not loaded from env: got %v want 9000", cfg.Server.Port)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Host not loaded from env: got %v want 127.0.0.1", cfg.Server.Host)
	}

	if cfg.Agents.Stoic.DBPath != "/tmp/stoic.db" {
		t.Errorf("Stoic DB path not loaded from env: got %v want /tmp/stoic.db", cfg.Agents.Stoic.DBPath)
	}

	if cfg.Agents.Tech.DBPath != "/tmp/tech.db" {
		t.Errorf("Tech DB path not loaded from env: got %v want /tmp/tech.db", cfg.Agents.Tech.DBPath)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		setupEnv    func()
		cleanupEnv  func()
		expectError bool
	}{
		{
			name: "Missing API key",
			setupEnv: func() {
				os.Setenv("STOIC_DB_PATH", "/tmp/stoic.db")
				os.Setenv("TECH_DB_PATH", "/tmp/tech.db")
			},
			cleanupEnv: func() {
				os.Unsetenv("STOIC_DB_PATH")
				os.Unsetenv("TECH_DB_PATH")
			},
			expectError: true,
		},
		{
			name: "Missing Stoic DB path",
			setupEnv: func() {
				os.Setenv("API_KEY", "test-key")
				os.Setenv("TECH_DB_PATH", "/tmp/tech.db")
			},
			cleanupEnv: func() {
				os.Unsetenv("API_KEY")
				os.Unsetenv("TECH_DB_PATH")
			},
			expectError: true,
		},
		{
			name: "Valid config",
			setupEnv: func() {
				os.Setenv("API_KEY", "test-key")
				os.Setenv("STOIC_DB_PATH", "/tmp/stoic.db")
				os.Setenv("TECH_DB_PATH", "/tmp/tech.db")
			},
			cleanupEnv: func() {
				os.Unsetenv("API_KEY")
				os.Unsetenv("STOIC_DB_PATH")
				os.Unsetenv("TECH_DB_PATH")
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer tt.cleanupEnv()

			_, err := Load("")
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
