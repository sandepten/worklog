package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	WorkNotesLocation string
	WorkplaceName     string   // Default workplace (for backward compatibility)
	Workplaces        []string // List of available workplaces
	OpenCodeServer    string
	AIProvider        string
	AIModel           string
}

// Load reads the configuration from ~/.config/worklog/config
func Load() (*Config, error) {
	// Load config from ~/.config/worklog/config
	configPath := getConfigPath()
	loadConfigFile(configPath)

	workplaceName := getEnv("WORKPLACE_NAME", "Work")
	workplacesStr := getEnv("WORKPLACES", "")

	// Parse workplaces list
	var workplaces []string
	if workplacesStr != "" {
		for _, wp := range strings.Split(workplacesStr, ",") {
			wp = strings.TrimSpace(wp)
			if wp != "" {
				workplaces = append(workplaces, wp)
			}
		}
	}

	// If no workplaces defined, use the default workplace name
	if len(workplaces) == 0 {
		workplaces = []string{workplaceName}
	}

	cfg := &Config{
		WorkNotesLocation: getEnv("WORK_NOTES_LOCATION", "~/Documents/obsidian-notes/Inbox/work"),
		WorkplaceName:     workplaceName,
		Workplaces:        workplaces,
		OpenCodeServer:    getEnv("OPENCODE_SERVER", "http://127.0.0.1:4096"),
		AIProvider:        getEnv("AI_PROVIDER", "github-copilot"),
		AIModel:           getEnv("AI_MODEL", "claude-sonnet-4"),
	}

	// Expand ~ in the path
	cfg.WorkNotesLocation = expandPath(cfg.WorkNotesLocation)

	return cfg, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "worklog", "config")
}

// loadConfigFile reads a key=value config file and sets environment variables
func loadConfigFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		return // Config file doesn't exist, use defaults
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set in environment
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
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

// AddWorkplace adds a new workplace to the config and saves it
func (c *Config) AddWorkplace(name string) error {
	// Check if workplace already exists
	for _, wp := range c.Workplaces {
		if strings.EqualFold(wp, name) {
			return fmt.Errorf("workplace '%s' already exists", name)
		}
	}

	// Add to the list
	c.Workplaces = append(c.Workplaces, name)

	// Save to config file
	return c.saveWorkplaces()
}

// RenameWorkplace renames an existing workplace in the config and saves it
func (c *Config) RenameWorkplace(oldName, newName string) error {
	// Check if new name already exists
	for _, wp := range c.Workplaces {
		if strings.EqualFold(wp, newName) {
			return fmt.Errorf("workplace '%s' already exists", newName)
		}
	}

	// Find and rename
	found := false
	for i, wp := range c.Workplaces {
		if wp == oldName {
			c.Workplaces[i] = newName
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("workplace '%s' not found", oldName)
	}

	// Save to config file
	return c.saveWorkplaces()
}

// saveWorkplaces writes the updated workplaces list to the config file
func (c *Config) saveWorkplaces() error {
	configPath := getConfigPath()

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing config file content
	existingContent := make(map[string]string)
	if file, err := os.Open(configPath); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				existingContent[key] = value
			}
		}
		file.Close()
	}

	// Update WORKPLACES
	existingContent["WORKPLACES"] = strings.Join(c.Workplaces, ",")

	// Write back to file
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to open config file for writing: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write all config values
	for key, value := range existingContent {
		fmt.Fprintf(writer, "%s=%s\n", key, value)
	}

	return writer.Flush()
}

// GetConfigPath returns the path to the config file (exported for use by commands)
func GetConfigPath() string {
	return getConfigPath()
}
