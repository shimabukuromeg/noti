package config

import (
	"fmt"
	"os"
)

// Config holds the application configuration
type Config struct {
	NotionToken string
	DatabaseID  string
}

// Load reads configuration from environment variables.
// NOTION_TOKEN (required for all commands)
// NOTI_DATABASE_ID (required for list, push with new page creation)
func Load() *Config {
	return &Config{
		NotionToken: os.Getenv("NOTION_TOKEN"),
		DatabaseID:  os.Getenv("NOTI_DATABASE_ID"),
	}
}

// ValidateToken checks if the Notion token is set, returns error if not
func (c *Config) ValidateToken() error {
	if c.NotionToken == "" {
		return fmt.Errorf("NOTION_TOKEN environment variable is not set")
	}
	return nil
}

// ValidateDatabase checks if the database ID is set, returns error if not
func (c *Config) ValidateDatabase() error {
	if c.DatabaseID == "" {
		return fmt.Errorf("NOTI_DATABASE_ID environment variable is not set")
	}
	return nil
}
