package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	DBPath      string
	OllamaHost  string
	OllamaModel string
}

const (
	defaultDBPath      = "./transactions.db"
	defaultOllamaHost  = "http://localhost:11434"
	defaultOllamaModel = "dolphin3"
)

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	dbPath := getEnv("DB_PATH", defaultDBPath)

	// Expand ~ to home directory if present
	if len(dbPath) > 0 && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	cfg := &Config{
		DBPath:      dbPath,
		OllamaHost:  getEnv("OLLAMA_HOST", defaultOllamaHost),
		OllamaModel: getEnv("OLLAMA_MODEL", defaultOllamaModel),
	}

	return cfg, nil
}

// DatabasePath returns the database file path
func (c *Config) DatabasePath() string {
	return c.DBPath
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
