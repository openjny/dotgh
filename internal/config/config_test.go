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
