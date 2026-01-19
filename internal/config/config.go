package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds the application configuration loaded from .env
type Config struct {
	WorkNotesLocation string
	WorkplaceName     string
	OpenCodeServer    string
	AIProvider        string
	AIModel           string
}

// Load reads the configuration from .env file
func Load() (*Config, error) {
	// Try to load .env from current directory or executable directory
	_ = godotenv.Load()

	cfg := &Config{
		WorkNotesLocation: getEnv("WORK_NOTES_LOCATION", "~/Documents/obsidian-notes/Inbox/work"),
		WorkplaceName:     getEnv("WORKPLACE_NAME", "Work"),
		OpenCodeServer:    getEnv("OPENCODE_SERVER", "http://127.0.0.1:4096"),
		AIProvider:        getEnv("AI_PROVIDER", "github-copilot"),
		AIModel:           getEnv("AI_MODEL", "claude-sonnet-4"),
	}

	// Expand ~ in the path
	cfg.WorkNotesLocation = expandPath(cfg.WorkNotesLocation)

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// EnsureNotesDirectory creates the notes directory if it doesn't exist
func (c *Config) EnsureNotesDirectory() error {
	return os.MkdirAll(c.WorkNotesLocation, 0755)
}
