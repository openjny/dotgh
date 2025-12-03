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

// Config represents the dotgh configuration.
type Config struct {
	Editor   string   `yaml:"editor,omitempty"`
	Includes []string `yaml:"includes"`
	Excludes []string `yaml:"excludes,omitempty"`
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
	sb.WriteString("# editor: エディタコマンドを指定します（例: \"code --wait\", \"vim\"）\n")
	sb.WriteString("# 未設定の場合、VISUAL, EDITOR, GIT_EDITOR 環境変数、\n")
	sb.WriteString("# またはプラットフォームデフォルト（Linux/macOS: vi, Windows: notepad）が使用されます\n")
	sb.WriteString("# editor: \"\"\n")
	sb.WriteString("\n")

	// Includes section
	sb.WriteString("# includes: テンプレートとして管理するファイルパターンを指定します（必須）\n")
	sb.WriteString("# glob形式（*, ?, [abc]）をサポート。**（再帰パターン）は未サポート。\n")
	sb.WriteString("includes:\n")
	for _, include := range DefaultIncludes {
		sb.WriteString(fmt.Sprintf("  - \"%s\"\n", include))
	}
	sb.WriteString("\n")

	// Excludes section (commented out)
	sb.WriteString("# excludes: includes にマッチしたファイルから除外するパターンを指定します\n")
	sb.WriteString("# ローカル設定や機密ファイルの除外に便利です\n")
	sb.WriteString("# excludes:\n")
	sb.WriteString("#   - \".github/prompts/local.prompt.md\"\n")
	sb.WriteString("#   - \".github/prompts/secret-*.prompt.md\"\n")

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
