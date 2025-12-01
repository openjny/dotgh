package config

import (
	"os"
	"path/filepath"
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

	if len(DefaultTargets) != len(expected) {
		t.Errorf("DefaultTargets length = %d, want %d", len(DefaultTargets), len(expected))
	}

	for i, target := range expected {
		if DefaultTargets[i] != target {
			t.Errorf("DefaultTargets[%d] = %q, want %q", i, DefaultTargets[i], target)
		}
	}
}

func TestLoadConfigFileNotExist(t *testing.T) {
	// Setup: use a temp directory with no config file
	tempDir := t.TempDir()

	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return default targets
	if len(cfg.Targets) != len(DefaultTargets) {
		t.Errorf("Targets length = %d, want %d", len(cfg.Targets), len(DefaultTargets))
	}

	for i, target := range DefaultTargets {
		if cfg.Targets[i] != target {
			t.Errorf("Targets[%d] = %q, want %q", i, cfg.Targets[i], target)
		}
	}
}

func TestLoadConfigFileExists(t *testing.T) {
	// Setup: create a config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `targets:
  - "custom/file.md"
  - "another/*.txt"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return custom targets
	expected := []string{"custom/file.md", "another/*.txt"}
	if len(cfg.Targets) != len(expected) {
		t.Errorf("Targets length = %d, want %d", len(cfg.Targets), len(expected))
	}

	for i, target := range expected {
		if cfg.Targets[i] != target {
			t.Errorf("Targets[%d] = %q, want %q", i, cfg.Targets[i], target)
		}
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Setup: create an invalid config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	invalidContent := `targets: [invalid yaml`
	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := LoadFromDir(tempDir)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfigEmptyTargets(t *testing.T) {
	// Setup: create a config file with empty targets
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `targets: []
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := LoadFromDir(tempDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Empty targets should be respected (user explicitly set empty)
	if len(cfg.Targets) != 0 {
		t.Errorf("Targets length = %d, want 0", len(cfg.Targets))
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
