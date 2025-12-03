package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
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
		wantExcludes []string
		wantErr      bool
	}{
		{
			name:         "no config file returns defaults",
			configYAML:   "",
			wantIncludes: DefaultIncludes,
			wantExcludes: nil,
			wantErr:      false,
		},
		{
			name: "custom includes",
			configYAML: `includes:
  - "custom/file.md"
  - "another/*.txt"
`,
			wantIncludes: []string{"custom/file.md", "another/*.txt"},
			wantExcludes: nil,
			wantErr:      false,
		},
		{
			name:         "empty includes",
			configYAML:   "includes: []\n",
			wantIncludes: []string{},
			wantExcludes: nil,
			wantErr:      false,
		},
		{
			name:         "invalid YAML",
			configYAML:   "includes: [invalid yaml",
			wantIncludes: nil,
			wantExcludes: nil,
			wantErr:      true,
		},
		{
			name: "with excludes",
			configYAML: `includes:
  - "AGENTS.md"
excludes:
  - ".github/prompts/local.prompt.md"
  - ".github/prompts/secret-*.prompt.md"
`,
			wantIncludes: []string{"AGENTS.md"},
			wantExcludes: []string{".github/prompts/local.prompt.md", ".github/prompts/secret-*.prompt.md"},
			wantErr:      false,
		},
		{
			name: "empty excludes",
			configYAML: `includes: []
excludes: []
`,
			wantIncludes: []string{},
			wantExcludes: []string{},
			wantErr:      false,
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

			if !reflect.DeepEqual(cfg.Excludes, tt.wantExcludes) {
				t.Errorf("Excludes = %v, want %v", cfg.Excludes, tt.wantExcludes)
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

func TestGenerateDefaultConfigContent(t *testing.T) {
	content := GenerateDefaultConfigContent()

	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "contains editor comment",
			contains: "# editor:",
		},
		{
			name:     "contains editor description",
			contains: "Specify the editor command",
		},
		{
			name:     "contains VISUAL/EDITOR environment variables mention",
			contains: "VISUAL, EDITOR, GIT_EDITOR",
		},
		{
			name:     "contains includes comment",
			contains: "# includes:",
		},
		{
			name:     "contains includes description",
			contains: "Specify file patterns to manage as templates",
		},
		{
			name:     "contains glob description",
			contains: "glob patterns",
		},
		{
			name:     "contains includes field",
			contains: "includes:",
		},
		{
			name:     "contains AGENTS.md",
			contains: `"AGENTS.md"`,
		},
		{
			name:     "contains excludes comment",
			contains: "# excludes:",
		},
		{
			name:     "contains excludes description",
			contains: "Specify patterns to exclude from matched includes",
		},
		{
			name:     "contains commented excludes example",
			contains: "#   - ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(content, tt.contains) {
				t.Errorf("GenerateDefaultConfigContent() should contain %q\nGot:\n%s", tt.contains, content)
			}
		})
	}
}

func TestGenerateDefaultConfigContentContainsAllDefaultIncludes(t *testing.T) {
	content := GenerateDefaultConfigContent()

	for _, include := range DefaultIncludes {
		if !strings.Contains(content, include) {
			t.Errorf("GenerateDefaultConfigContent() should contain default include %q\nGot:\n%s", include, content)
		}
	}
}

func TestGenerateDefaultConfigContentIsParseable(t *testing.T) {
	content := GenerateDefaultConfigContent()

	var cfg Config
	err := yaml.Unmarshal([]byte(content), &cfg)
	if err != nil {
		t.Fatalf("GenerateDefaultConfigContent() should produce valid YAML, got error: %v\nContent:\n%s", err, content)
	}

	// Verify parsed includes match DefaultIncludes
	if !reflect.DeepEqual(cfg.Includes, DefaultIncludes) {
		t.Errorf("Parsed Includes = %v, want %v", cfg.Includes, DefaultIncludes)
	}

	// editor should be empty (commented out)
	if cfg.Editor != "" {
		t.Errorf("Parsed Editor = %q, want empty string", cfg.Editor)
	}

	// excludes should be nil (commented out)
	if cfg.Excludes != nil {
		t.Errorf("Parsed Excludes = %v, want nil", cfg.Excludes)
	}
}

func TestCreateDefaultConfigFileWithComments(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create default config file
	err := CreateDefaultConfigFile(configPath)
	if err != nil {
		t.Fatalf("CreateDefaultConfigFile() error = %v", err)
	}

	// Read file content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}
	content := string(data)

	// Verify comments are present
	tests := []struct {
		name     string
		contains string
	}{
		{
			name:     "contains editor comment",
			contains: "# editor:",
		},
		{
			name:     "contains includes comment",
			contains: "# includes:",
		},
		{
			name:     "contains excludes comment",
			contains: "# excludes:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(content, tt.contains) {
				t.Errorf("Created config file should contain %q\nGot:\n%s", tt.contains, content)
			}
		})
	}

	// Verify file is still parseable and has correct values
	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("LoadFromDir() error = %v", err)
	}

	if !reflect.DeepEqual(cfg.Includes, DefaultIncludes) {
		t.Errorf("Includes = %v, want %v", cfg.Includes, DefaultIncludes)
	}
}
