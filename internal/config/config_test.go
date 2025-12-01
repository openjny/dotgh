package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDefaultTargets(t *testing.T) {
	expected := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".github/instructions/*.instructions.md",
		".github/prompts/*.prompt.md",
		".vscode/mcp.json",
	}

	if !reflect.DeepEqual(DefaultTargets, expected) {
		t.Errorf("DefaultTargets = %v, want %v", DefaultTargets, expected)
	}
}

func TestLoadFromDir(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string // empty means no config file
		wantTargets []string
		wantErr     bool
	}{
		{
			name:        "no config file returns defaults",
			configYAML:  "",
			wantTargets: DefaultTargets,
			wantErr:     false,
		},
		{
			name: "custom targets",
			configYAML: `targets:
  - "custom/file.md"
  - "another/*.txt"
`,
			wantTargets: []string{"custom/file.md", "another/*.txt"},
			wantErr:     false,
		},
		{
			name:        "empty targets",
			configYAML:  "targets: []\n",
			wantTargets: []string{},
			wantErr:     false,
		},
		{
			name:        "invalid YAML",
			configYAML:  "targets: [invalid yaml",
			wantTargets: nil,
			wantErr:     true,
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

			if !reflect.DeepEqual(cfg.Targets, tt.wantTargets) {
				t.Errorf("Targets = %v, want %v", cfg.Targets, tt.wantTargets)
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
