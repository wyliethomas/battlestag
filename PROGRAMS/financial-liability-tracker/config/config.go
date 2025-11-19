package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	DBPath string
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Default to ~/.local/share/financial-liability-tracker/liabilities.db
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		dbPath = filepath.Join(home, ".local", "share", "financial-liability-tracker", "liabilities.db")
	}

	return &Config{
		DBPath: dbPath,
	}, nil
}

// DatabasePath returns the database file path
func (c *Config) DatabasePath() string {
	return c.DBPath
}
