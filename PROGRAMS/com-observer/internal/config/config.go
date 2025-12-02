package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete application configuration
type Config struct {
	User       UserConfig       `yaml:"user"`
	Queue      QueueConfig      `yaml:"queue"`
	Slack      SlackConfig      `yaml:"slack"`
	Gmail      GmailConfig      `yaml:"gmail"`
	Trello     TrelloConfig     `yaml:"trello"`
	NoteWriter NoteWriterConfig `yaml:"notewriter"`
	LLM        LLMConfig        `yaml:"llm"`
}

// UserConfig contains user identity information
type UserConfig struct {
	Name    string `yaml:"name"`
	SlackID string `yaml:"slack_id"`
	Email   string `yaml:"email"`
}

// QueueConfig defines the message queue settings
type QueueConfig struct {
	Type string `yaml:"type"` // "filesystem"
	Path string `yaml:"path"` // Directory path for filesystem queue
}

// SlackConfig contains Slack API settings
type SlackConfig struct {
	Token           string   `yaml:"token"`
	ChannelsInclude []string `yaml:"channels_include"` // Empty = all channels
	ChannelsExclude []string `yaml:"channels_exclude"`
	WatchDMs        bool     `yaml:"watch_dms"`
	SyncInterval    string   `yaml:"sync_interval"` // e.g., "5m", "1h"
}

// GmailConfig contains Gmail API settings
type GmailConfig struct {
	CredentialsFile string   `yaml:"credentials_file"`
	Labels          []string `yaml:"labels"` // Gmail labels to watch
	SyncInterval    string   `yaml:"sync_interval"`
}

// TrelloConfig contains Trello API settings
type TrelloConfig struct {
	APIKey       string   `yaml:"api_key"`
	Token        string   `yaml:"token"`
	Boards       []string `yaml:"boards"` // Board IDs to watch
	SyncInterval string   `yaml:"sync_interval"`
}

// NoteWriterConfig defines note-taking system settings
type NoteWriterConfig struct {
	Type     string         `yaml:"type"` // "obsidian" | "notion"
	Obsidian ObsidianConfig `yaml:"obsidian,omitempty"`
	Notion   NotionConfig   `yaml:"notion,omitempty"`
}

// ObsidianConfig contains Obsidian vault settings
type ObsidianConfig struct {
	VaultPath string `yaml:"vault_path"`
	TasksFile string `yaml:"tasks_file"` // Relative to vault_path
}

// NotionConfig contains Notion API settings
type NotionConfig struct {
	Token      string `yaml:"token"`
	DatabaseID string `yaml:"database_id"`
}

// LLMConfig contains LLM provider settings
type LLMConfig struct {
	Provider string `yaml:"provider"` // "anthropic" | "openai" | "custom"
	APIKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint,omitempty"` // For custom providers
	Model    string `yaml:"model"`              // e.g., "claude-3-5-sonnet-20241022"
}

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required configuration fields are present
func (c *Config) Validate() error {
	if c.User.Name == "" {
		return fmt.Errorf("user.name is required")
	}

	if c.Queue.Type == "" {
		c.Queue.Type = "filesystem"
	}
	if c.Queue.Path == "" {
		return fmt.Errorf("queue.path is required")
	}

	if c.LLM.Provider == "" {
		return fmt.Errorf("llm.provider is required")
	}
	if c.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}

	if c.NoteWriter.Type == "" {
		return fmt.Errorf("notewriter.type is required")
	}

	switch c.NoteWriter.Type {
	case "obsidian":
		if c.NoteWriter.Obsidian.VaultPath == "" {
			return fmt.Errorf("notewriter.obsidian.vault_path is required")
		}
		if c.NoteWriter.Obsidian.TasksFile == "" {
			return fmt.Errorf("notewriter.obsidian.tasks_file is required")
		}
	case "notion":
		if c.NoteWriter.Notion.Token == "" {
			return fmt.Errorf("notewriter.notion.token is required")
		}
		if c.NoteWriter.Notion.DatabaseID == "" {
			return fmt.Errorf("notewriter.notion.database_id is required")
		}
	default:
		return fmt.Errorf("notewriter.type must be 'obsidian' or 'notion'")
	}

	return nil
}
