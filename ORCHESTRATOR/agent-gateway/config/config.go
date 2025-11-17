package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Agents AgentsConfig `yaml:"agents"`
	Server ServerConfig `yaml:"server"`
	Auth   AuthConfig   `yaml:"auth"`
	Logging LoggingConfig `yaml:"logging"`
}

// AgentsConfig contains database paths for all agents
type AgentsConfig struct {
	Stoic StoicConfig `yaml:"stoic"`
	Tech  TechConfig  `yaml:"tech"`
}

// StoicConfig contains configuration for stoic thoughts agent
type StoicConfig struct {
	ExecutablePath string `yaml:"executable_path"`
	DBPath         string `yaml:"db_path"` // Optional: for advanced features
}

// TechConfig contains configuration for tech tips agent
type TechConfig struct {
	ExecutablePath string `yaml:"executable_path"`
	DBPath         string `yaml:"db_path"` // Optional: for advanced features
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port string `yaml:"port"`
	Host string `yaml:"host"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	APIKey string `yaml:"api_key"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	AccessLog string `yaml:"access_log"`
	ErrorLog  string `yaml:"error_log"`
}

// Load reads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Logging: LoggingConfig{
			AccessLog: "logs/access.log",
			ErrorLog:  "logs/error.log",
		},
	}

	// Read config file if it exists
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = port
	}
	if host := os.Getenv("HOST"); host != "" {
		cfg.Server.Host = host
	}
	if apiKey := os.Getenv("API_KEY"); apiKey != "" {
		cfg.Auth.APIKey = apiKey
	}
	if stoicExec := os.Getenv("STOIC_EXECUTABLE"); stoicExec != "" {
		cfg.Agents.Stoic.ExecutablePath = stoicExec
	}
	if stoicDB := os.Getenv("STOIC_DB_PATH"); stoicDB != "" {
		cfg.Agents.Stoic.DBPath = stoicDB
	}
	if techExec := os.Getenv("TECH_EXECUTABLE"); techExec != "" {
		cfg.Agents.Tech.ExecutablePath = techExec
	}
	if techDB := os.Getenv("TECH_DB_PATH"); techDB != "" {
		cfg.Agents.Tech.DBPath = techDB
	}

	// Validate required fields
	if cfg.Auth.APIKey == "" {
		return nil, fmt.Errorf("API key is required (set via config file or API_KEY env var)")
	}
	if cfg.Agents.Stoic.ExecutablePath == "" {
		return nil, fmt.Errorf("Stoic executable path is required")
	}
	if cfg.Agents.Tech.ExecutablePath == "" {
		return nil, fmt.Errorf("Tech executable path is required")
	}

	return cfg, nil
}
