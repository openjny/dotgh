package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultIncludes(t *testing.T) {
	expected := []string{
		"AGENTS.md",
		".github/agents/*.agent.md",
		".github/copilot-chat-modes/*.chatmode.md",
		".github/copilot-instructions.md",
		".github/instructions/*.instructions.md",
		".github/prompts/*.prompt.md",
		".vscode/mcp.json",
	}

	if !reflect.DeepEqual(DefaultIncludes, expected) {
		t.Errorf("DefaultIncludes = %v, want %v", DefaultIncludes, expected)
	}
}

func TestLoadFromDir(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string // empty means no config file
		wantIncludes []string
		wantErr      bool
	}{
		{
			name:         "no config file returns defaults",
			configYAML:   "",
			wantIncludes: DefaultIncludes,
			wantErr:      false,
		},
		{
			name: "custom includes",
			configYAML: `includes:
  - "custom/file.md"
  - "another/*.txt"
`,
			wantIncludes: []string{"custom/file.md", "another/*.txt"},
			wantErr:      false,
		},
		{
			name:         "empty includes",
			configYAML:   "includes: []\n",
			wantIncludes: []string{},
			wantErr:      false,
		},
		{
			name:         "invalid YAML",
			configYAML:   "includes: [invalid yaml",
			wantIncludes: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			// Create config file if content provided
			if tt.configYAML != "" {
				configPath := filepath.Join(tempDir, "config.yaml")
				if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
			}

			cfg, err := LoadFromDir(tempDir)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(cfg.Includes, tt.wantIncludes) {
				t.Errorf("Includes = %v, want %v", cfg.Includes, tt.wantIncludes)
			}
		})
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir()
	if dir == "" {
		t.Error("GetConfigDir() returned empty string")
	}

	// Should contain "dotgh" in the path
	if !filepath.IsAbs(dir) {
		t.Errorf("GetConfigDir() should return absolute path, got %q", dir)
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}

	// Should be an absolute path ending with config.yaml
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() should return absolute path, got %q", path)
	}

	if filepath.Base(path) != "config.yaml" {
		t.Errorf("GetConfigPath() should end with config.yaml, got %q", path)
	}
}

func TestLoadFromDirWithEditor(t *testing.T) {
	tests := []struct {
		name       string
		configYAML string
		wantEditor string
	}{
		{
			name:       "no editor field",
			configYAML: "includes: []\n",
			wantEditor: "",
		},
		{
			name: "with editor field",
			configYAML: `editor: "vim"
includes: []
`,
			wantEditor: "vim",
		},
		{
			name: "editor with arguments",
			configYAML: `editor: "code --wait"
includes: []
`,
			wantEditor: "code --wait",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			configPath := filepath.Join(tempDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			cfg, err := LoadFromDir(tempDir)
			if err != nil {
				t.Fatalf("LoadFromDir() error = %v", err)
			}

			if cfg.Editor != tt.wantEditor {
				t.Errorf("Editor = %q, want %q", cfg.Editor, tt.wantEditor)
			}
		})
	}
}

func TestCreateDefaultConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create default config file
	err := CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("CreateDefaultConfigFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load and verify contents
	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if !reflect.DeepEqual(cfg.Includes, DefaultIncludes) {
		t.Errorf("Includes = %v, want %v", cfg.Includes, DefaultIncludes)
	}
}

func TestCreateDefaultConfigFileInNestedDir(t *testing.T) {
	tempDir := t.TempDir()
	nestedDir := filepath.Join(tempDir, "nested", "dir")
	configPath := filepath.Join(nestedDir, "config.yaml")

	// Create default config file in nested directory (should create parent dirs)
	err := CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("CreateDefaultConfigFile() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestLoadFromDirWithExcludes(t *testing.T) {
	tests := []struct {
		name         string
		configYAML   string
		wantExcludes []string
	}{
		{
			name:         "no excludes field returns nil",
			configYAML:   "includes: []\n",
			wantExcludes: nil,
		},
		{
			name: "empty excludes",
			configYAML: `includes: []
excludes: []
`,
			wantExcludes: []string{},
		},
		{
			name: "single exclude pattern",
			configYAML: `includes: []
excludes:
  - ".github/prompts/local.prompt.md"
`,
			wantExcludes: []string{".github/prompts/local.prompt.md"},
		},
		{
			name: "multiple exclude patterns",
			configYAML: `includes: []
excludes:
  - ".github/prompts/local.prompt.md"
  - ".github/prompts/secret-*.prompt.md"
  - "AGENTS.md"
`,
			wantExcludes: []string{
				".github/prompts/local.prompt.md",
				".github/prompts/secret-*.prompt.md",
				"AGENTS.md",
			},
		},
		{
			name: "excludes with includes",
			configYAML: `includes:
  - "AGENTS.md"
  - ".github/prompts/*.prompt.md"
excludes:
  - ".github/prompts/local.prompt.md"
`,
			wantExcludes: []string{".github/prompts/local.prompt.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			configPath := filepath.Join(tempDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write config file: %v", err)
			}

			cfg, err := LoadFromDir(tempDir)
			if err != nil {
				t.Fatalf("LoadFromDir() error = %v", err)
			}

			if !reflect.DeepEqual(cfg.Excludes, tt.wantExcludes) {
				t.Errorf("Excludes = %v, want %v", cfg.Excludes, tt.wantExcludes)
			}
		})
	}
}

func TestDefaultExcludes(t *testing.T) {
	// Test that default config has no excludes (nil or empty)
	tempDir := t.TempDir()
	// No config file - should return defaults

	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	// Default excludes should be nil or empty
	if len(cfg.Excludes) != 0 {
		t.Errorf("Default Excludes should be empty, got %v", cfg.Excludes)
	}
}
