// Package config provides configuration management for dotgh.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DefaultIncludes defines the default glob patterns for files to copy from templates.
// These are used when no config file exists.
var DefaultIncludes = []string{
	"AGENTS.md",
	".github/agents/*.agent.md",
	".github/copilot-chat-modes/*.chatmode.md",
	".github/copilot-instructions.md",
	".github/instructions/*.instructions.md",
	".github/prompts/*.prompt.md",
	".vscode/mcp.json",
}

// SyncConfig represents the sync configuration.
type SyncConfig struct {
	Repo       string `yaml:"repo,omitempty"`
	Branch     string `yaml:"branch,omitempty"`
	AutoCommit bool   `yaml:"auto_commit,omitempty"`
}

// Config represents the dotgh configuration.
type Config struct {
	Editor       string      `yaml:"editor,omitempty"`
	TemplatesDir string      `yaml:"templates_dir,omitempty"`
	Includes     []string    `yaml:"includes"`
	Excludes     []string    `yaml:"excludes,omitempty"`
	Sync         *SyncConfig `yaml:"sync,omitempty"`
}

// GetTemplatesDir returns the templates directory path.
// If TemplatesDir is set in the config, it returns that path (with tilde expansion).
// Otherwise, it returns the default templates directory.
func (c *Config) GetTemplatesDir() string {
	if c.TemplatesDir != "" {
		return expandTilde(c.TemplatesDir)
	}
	return GetDefaultTemplatesDir()
}

// expandTilde expands a leading ~ in the path to the user's home directory.
func expandTilde(path string) string {
	if path == "" {
		return path
	}
	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
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

// GetDefaultTemplatesDir returns the default templates directory path.
// It follows the XDG Base Directory Specification using os.UserConfigDir().
func GetDefaultTemplatesDir() string {
	return filepath.Join(GetConfigDir(), "templates")
}

// GetConfigPath returns the path to the dotgh configuration file.
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.yaml")
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
			// Return default config if file does not exist
			return &Config{Includes: DefaultIncludes}, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// GenerateDefaultConfigContent generates the default configuration file content
// with comments explaining each field.
func GenerateDefaultConfigContent() string {
	var sb strings.Builder

	// Editor section (commented out)
	sb.WriteString("# editor: Specify the editor command (e.g., \"code --wait\", \"vim\")\n")
	sb.WriteString("# If not set, VISUAL, EDITOR, GIT_EDITOR environment variables,\n")
	sb.WriteString("# or platform defaults (Linux/macOS: vi, Windows: notepad) will be used.\n")
	sb.WriteString("# editor: \"\"\n")
	sb.WriteString("\n")

	// Templates directory section (commented out)
	sb.WriteString("# templates_dir: Specify a custom templates directory location.\n")
	sb.WriteString("# If not set, the default location is used:\n")
	sb.WriteString("#   Linux/macOS: ~/.config/dotgh/templates/\n")
	sb.WriteString("#   Windows: %LOCALAPPDATA%\\dotgh\\templates\\\n")
	sb.WriteString("# Supports tilde expansion (e.g., \"~/my-templates\").\n")
	sb.WriteString("# templates_dir: \"\"\n")
	sb.WriteString("\n")

	// Includes section
	sb.WriteString("# includes: Specify file patterns to manage as templates (required)\n")
	sb.WriteString("# Supports glob patterns (*, ?, [abc]). ** (recursive) is not supported.\n")
	sb.WriteString("includes:\n")
	for _, include := range DefaultIncludes {
		sb.WriteString(fmt.Sprintf("  - \"%s\"\n", include))
	}
	sb.WriteString("\n")

	// Excludes section (commented out)
	sb.WriteString("# excludes: Specify patterns to exclude from matched includes\n")
	sb.WriteString("# Useful for excluding local configs or sensitive files.\n")
	sb.WriteString("# excludes:\n")
	sb.WriteString("#   - \".github/prompts/local.prompt.md\"\n")
	sb.WriteString("#   - \".github/prompts/secret-*.prompt.md\"\n")
	sb.WriteString("\n")

	// Sync section (commented out)
	sb.WriteString("# sync: Configuration for syncing settings across machines\n")
	sb.WriteString("# sync:\n")
	sb.WriteString("#   repo: \"git@github.com:username/dotgh-sync.git\"  # Sync repository URL\n")
	sb.WriteString("#   branch: \"main\"                                   # Branch to use\n")
	sb.WriteString("#   auto_commit: true                                # Auto-commit on push\n")

	return sb.String()
}

// CreateDefaultConfigFile creates a config file with default values at the specified path.
// It creates parent directories if they do not exist.
func CreateDefaultConfigFile(path string) error {
	// Create parent directories if they do not exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	content := GenerateDefaultConfigContent()

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}
