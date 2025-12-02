package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigListWithNoConfigFile(t *testing.T) {
	// Create temp directory without config file
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dotgh")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	cmd := NewConfigListCmd(configDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config list command failed: %v", err)
	}

	output := buf.String()

	// Should show config file path
	if !strings.Contains(output, "# Config file:") {
		t.Errorf("output should contain config file path, got:\n%s", output)
	}

	// Should show default includes
	if !strings.Contains(output, "AGENTS.md") {
		t.Errorf("output should contain default includes, got:\n%s", output)
	}
}

func TestConfigListWithExistingConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dotgh")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	// Create custom config file
	configContent := `editor: "vim"
includes:
  - "custom/*.md"
  - "another/file.txt"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	cmd := NewConfigListCmd(configDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config list command failed: %v", err)
	}

	output := buf.String()

	// Should show editor setting
	if !strings.Contains(output, "editor:") {
		t.Errorf("output should contain editor setting, got:\n%s", output)
	}
	if !strings.Contains(output, "vim") {
		t.Errorf("output should contain 'vim', got:\n%s", output)
	}

	// Should show custom includes
	if !strings.Contains(output, "custom/*.md") {
		t.Errorf("output should contain custom includes, got:\n%s", output)
	}
}

func TestConfigListOutputFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dotgh")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	cmd := NewConfigListCmd(configDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("config list command failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(output, "\n")

	// First line should be a comment with config file path
	if !strings.HasPrefix(lines[0], "# Config file:") {
		t.Errorf("first line should start with '# Config file:', got:\n%s", lines[0])
	}

	// Should have includes section
	if !strings.Contains(output, "includes:") {
		t.Errorf("output should contain 'includes:' section, got:\n%s", output)
	}
}

func TestConfigEditCreatesDefaultConfigIfNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dotgh")
	configPath := filepath.Join(configDir, "config.yaml")

	// Verify config doesn't exist
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Fatal("config file should not exist before test")
	}

	// Run ensureConfigExists
	err := ensureConfigExists(configDir)
	if err != nil {
		t.Fatalf("ensureConfigExists failed: %v", err)
	}

	// Verify config now exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file should exist after ensureConfigExists")
	}

	// Verify content is valid
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if !strings.Contains(string(content), "includes:") {
		t.Errorf("config file should contain includes section, got:\n%s", string(content))
	}
}

func TestConfigEditDoesNotOverwriteExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "dotgh")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	originalContent := "editor: vim\nincludes:\n  - custom.md\n"
	if err := os.WriteFile(configPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Run ensureConfigExists
	err := ensureConfigExists(configDir)
	if err != nil {
		t.Fatalf("ensureConfigExists failed: %v", err)
	}

	// Verify content wasn't changed
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("config file content should not be changed.\nWant:\n%s\nGot:\n%s", originalContent, string(content))
	}
}

func TestConfigParentCommand(t *testing.T) {
	cmd := NewConfigCmd()

	// Check that config command has subcommands
	subCommands := cmd.Commands()
	if len(subCommands) < 2 {
		t.Errorf("config command should have at least 2 subcommands, got %d", len(subCommands))
	}

	// Check for list and edit subcommands
	hasListCmd := false
	hasEditCmd := false
	for _, sub := range subCommands {
		if sub.Name() == "list" {
			hasListCmd = true
		}
		if sub.Name() == "edit" {
			hasEditCmd = true
		}
	}

	if !hasListCmd {
		t.Error("config command should have 'list' subcommand")
	}
	if !hasEditCmd {
		t.Error("config command should have 'edit' subcommand")
	}
}

func TestBuildEditorCommand(t *testing.T) {
	tests := []struct {
		name         string
		configEditor string
		envEditor    string
		target       string
		wantContains string
	}{
		{
			name:         "uses config editor",
			configEditor: "vim",
			envEditor:    "",
			target:       "/path/to/file",
			wantContains: "vim",
		},
		{
			name:         "falls back to env editor",
			configEditor: "",
			envEditor:    "nano",
			target:       "/path/to/file",
			wantContains: "nano",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			oldEditor := os.Getenv("EDITOR")
			defer os.Setenv("EDITOR", oldEditor)

			if tt.envEditor != "" {
				os.Setenv("EDITOR", tt.envEditor)
			} else {
				os.Unsetenv("EDITOR")
			}

			// We can't easily test the full command execution since it opens an editor,
			// but we can test the editor detection logic
			args := buildEditorCommand(tt.configEditor, tt.target)
			if len(args) == 0 {
				t.Fatal("buildEditorCommand returned empty args")
			}
			if !strings.Contains(args[0], tt.wantContains) {
				t.Errorf("first arg should contain %q, got %q", tt.wantContains, args[0])
			}
		})
	}
}
