package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	AgentGateway AgentGatewayConfig `yaml:"agent_gateway"`
	Version      string             `yaml:"version"`
}

// AgentGatewayConfig represents the Agent Gateway connection settings
type AgentGatewayConfig struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		AgentGateway: AgentGatewayConfig{
			URL:    "http://192.168.1.140:8080",
			APIKey: "test-api-key-12345",
		},
		Version: "1.0",
	}
}

// GetConfigPath returns the path to the config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "battlestag-tui")
	configFile := filepath.Join(configDir, "config.yaml")

	return configFile, nil
}

// GetConfigDir returns the config directory path
func GetConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "battlestag-tui")
	return configDir, nil
}

// ConfigExists checks if the config file exists
func ConfigExists() bool {
	configPath, err := GetConfigPath()
	if err != nil {
		return false
	}

	_, err = os.Stat(configPath)
	return err == nil
}

// Load loads the configuration from file
// If the file doesn't exist, returns default config
// Environment variables override file config
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	var cfg *Config

	// If config file doesn't exist, use defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg = DefaultConfig()
	} else {
		// Read config file
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		cfg = &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Environment variables override file config
	if url := os.Getenv("AGENT_GATEWAY_URL"); url != "" {
		cfg.AgentGateway.URL = url
	}

	if apiKey := os.Getenv("AGENT_GATEWAY_API_KEY"); apiKey != "" {
		cfg.AgentGateway.APIKey = apiKey
	}

	return cfg, nil
}

// Save saves the configuration to file
func Save(cfg *Config) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.AgentGateway.URL == "" {
		return fmt.Errorf("agent gateway URL cannot be empty")
	}

	if c.AgentGateway.APIKey == "" {
		return fmt.Errorf("agent gateway API key cannot be empty")
	}

	return nil
}
