package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Ollama struct {
		URL   string `yaml:"url"`
		Model string `yaml:"model"`
	} `yaml:"ollama"`
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Execution struct {
		Timeout string `yaml:"timeout"`
	} `yaml:"execution"`
	Programs struct {
		Directory string `yaml:"directory"`
	} `yaml:"programs"`
}

// LoadConfig loads configuration from file and command-line flags
func LoadConfig() (*Config, error) {
	// Define command-line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	port := flag.String("port", "", "HTTP server port (overrides config)")
	programsDir := flag.String("programs", "", "Directory containing programs (overrides config)")
	ollamaURL := flag.String("ollama-url", "", "Ollama API URL (overrides config)")
	ollamaModel := flag.String("ollama-model", "", "Ollama model name (overrides config)")
	timeout := flag.String("timeout", "", "Program execution timeout (overrides config)")
	flag.Parse()

	// Load config from file
	config := &Config{}
	configData, err := os.ReadFile(*configFile)
	if err != nil {
		// Config file doesn't exist, use defaults
		config.Ollama.URL = "http://localhost:11434"
		config.Ollama.Model = "llama3.2:latest"
		config.Server.Port = "8080"
		config.Execution.Timeout = "30s"
		config.Programs.Directory = "./programs"
	} else {
		// Parse YAML config
		if err := yaml.Unmarshal(configData, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with command-line flags if provided
	if *port != "" {
		config.Server.Port = *port
	}
	if *programsDir != "" {
		config.Programs.Directory = *programsDir
	}
	if *ollamaURL != "" {
		config.Ollama.URL = *ollamaURL
	}
	if *ollamaModel != "" {
		config.Ollama.Model = *ollamaModel
	}
	if *timeout != "" {
		config.Execution.Timeout = *timeout
	}

	return config, nil
}

// GetTimeout parses the timeout string and returns a duration
func (c *Config) GetTimeout() (time.Duration, error) {
	return time.ParseDuration(c.Execution.Timeout)
}
