package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the database configuration
type Config struct {
	DBPath string
}

const defaultDBPath = "./transactions.db"

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	dbPath := os.Getenv("DB_PATH")

	// If not set, use default
	if dbPath == "" {
		dbPath = defaultDBPath
	}

	// Expand ~ to home directory if present
	if len(dbPath) > 0 && dbPath[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("get home directory: %w", err)
		}
		dbPath = filepath.Join(home, dbPath[1:])
	}

	// Ensure directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	cfg := &Config{
		DBPath: dbPath,
	}

	return cfg, nil
}

// DatabasePath returns the database file path
func (c *Config) DatabasePath() string {
	return c.DBPath
}
