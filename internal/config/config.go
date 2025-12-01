// Package config provides configuration management for dotgh.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultTargets defines the default files/directories to copy from templates.
// These are used when no config file exists.
var DefaultTargets = []string{
	"AGENTS.md",
	".github/copilot-instructions.md",
	".github/instructions/*.instructions.md",
	".github/prompts/*.prompt.md",
	".vscode/mcp.json",
}

// Config represents the dotgh configuration.
type Config struct {
	Targets []string `yaml:"targets"`
}

// GetConfigDir returns the path to the dotgh configuration directory.
// It follows the XDG Base Directory Specification.
func GetConfigDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "dotgh")
}

// Load loads the configuration from the default config directory.
// If no config file exists, it returns the default configuration.
func Load() (*Config, error) {
	return LoadFromDir(GetConfigDir())
}

// LoadFromDir loads the configuration from the specified directory.
// If no config file exists, it returns the default configuration.
func LoadFromDir(dir string) (*Config, error) {
	configPath := filepath.Join(dir, "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return &Config{Targets: DefaultTargets}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}
