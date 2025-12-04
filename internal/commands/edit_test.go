package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEditCmdValidation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "too many arguments",
			args:      []string{"template1", "template2"},
			wantError: true,
			errorMsg:  "accepts at most 1 arg(s), received 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cmd := NewEditCmd(tmpDir, tmpDir)
			cmd.SetArgs(tt.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestEditCmdTemplateNotFound(t *testing.T) {
	templatesDir := t.TempDir()
	configDir := t.TempDir()

	// Simulate user typing "n" to decline creation
	opts := &EditOptions{
		Stdin: strings.NewReader("n\n"),
	}
	cmd := NewEditCmdWithOptions(templatesDir, configDir, opts)
	cmd.SetArgs([]string{"non-existent-template"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for non-existent template, got nil")
	}

	expectedErrMsg := `template "non-existent-template" not found`
	if !strings.Contains(err.Error(), expectedErrMsg) {
		t.Errorf("expected error containing %q, got %q", expectedErrMsg, err.Error())
	}
}

func TestEditCmdWithExistingTemplateValidatesPath(t *testing.T) {
	// Create a template directory
	templatesDir := t.TempDir()
	templateName := "my-template"
	templateDir := filepath.Join(templatesDir, templateName)
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template directory: %v", err)
	}

	// Create a config directory
	configDir := t.TempDir()

	// Test that getTemplatePath correctly validates an existing template
	path, err := getTemplatePath(templatesDir, templateName)
	if err != nil {
		t.Fatalf("getTemplatePath should succeed for existing template: %v", err)
	}
	if path != templateDir {
		t.Errorf("expected path %q, got %q", templateDir, path)
	}

	// Test that the command is properly constructed
	cmd := NewEditCmd(templatesDir, configDir)
	if cmd.Use != "edit [template]" {
		t.Errorf("expected Use to be 'edit [template]', got %q", cmd.Use)
	}
	if cmd.Args == nil {
		t.Error("command should have Args validation")
	}
}

func TestGetTemplatePath(t *testing.T) {
	templatesDir := t.TempDir()

	// Create a template directory
	templateName := "test-template"
	templateDir := filepath.Join(templatesDir, templateName)
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template directory: %v", err)
	}

	// Create a file (not a directory) with a template-like name
	filePath := filepath.Join(templatesDir, "not-a-template")
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	tests := []struct {
		name          string
		templateName  string
		wantPath      string
		wantError     bool
		errorContains string
	}{
		{
			name:         "existing template",
			templateName: templateName,
			wantPath:     templateDir,
			wantError:    false,
		},
		{
			name:          "non-existing template",
			templateName:  "non-existent",
			wantPath:      "",
			wantError:     true,
			errorContains: "not found",
		},
		{
			name:          "file instead of directory",
			templateName:  "not-a-template",
			wantPath:      "",
			wantError:     true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := getTemplatePath(templatesDir, tt.templateName)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if path != tt.wantPath {
					t.Errorf("expected path %q, got %q", tt.wantPath, path)
				}
			}
		})
	}
}

func TestNewEditCmdWithConfig(t *testing.T) {
	templatesDir := t.TempDir()
	configDir := t.TempDir()

	// Create config with custom editor
	configContent := `editor: "vim"
includes:
  - "*.md"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	// Create a template
	templateDir := filepath.Join(templatesDir, "my-template")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatalf("failed to create template directory: %v", err)
	}

	cmd := NewEditCmdWithConfig(templatesDir, configDir)
	if cmd == nil {
		t.Fatal("NewEditCmdWithConfig returned nil")
	}

	// Verify command is properly configured
	if cmd.Use != "edit [template]" {
		t.Errorf("expected Use to be 'edit [template]', got %q", cmd.Use)
	}
}

func TestEditCmdNoArgs(t *testing.T) {
	tests := []struct {
		name              string
		templatesDirExist bool
		wantError         bool
		errorContains     string
	}{
		{
			name:              "templates directory exists",
			templatesDirExist: true,
			wantError:         false,
		},
		{
			name:              "templates directory does not exist",
			templatesDirExist: false,
			wantError:         true,
			errorContains:     "templates directory not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var templatesDir string
			if tt.templatesDirExist {
				templatesDir = t.TempDir()
			} else {
				templatesDir = filepath.Join(t.TempDir(), "non-existent")
			}
			configDir := t.TempDir()

			// Create config with echo as editor (safe command that exits immediately)
			configContent := `editor: "echo"
includes:
  - "*.md"
`
			if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
				t.Fatalf("failed to create config file: %v", err)
			}

			cmd := NewEditCmd(templatesDir, configDir)
			cmd.SetArgs([]string{})

			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			err := cmd.Execute()
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestEditCmdCreateWithFlag(t *testing.T) {
	templatesDir := t.TempDir()
	configDir := t.TempDir()

	// Create config with echo as editor
	configContent := `editor: "echo"
includes:
  - "*.md"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	templateName := "new-template"
	templatePath := filepath.Join(templatesDir, templateName)

	// Verify template doesn't exist
	if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
		t.Fatal("template should not exist before test")
	}

	opts := &EditOptions{
		Stdin: strings.NewReader(""), // No input needed with --create flag
	}
	cmd := NewEditCmdWithOptions(templatesDir, configDir, opts)
	cmd.SetArgs([]string{templateName, "--create"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify template was created
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Error("template directory should be created with --create flag")
	}

	if !strings.Contains(buf.String(), "Created new template") {
		t.Errorf("output should indicate template was created, got:\n%s", buf.String())
	}
}

func TestEditCmdCreateWithPrompt(t *testing.T) {
	templatesDir := t.TempDir()
	configDir := t.TempDir()

	// Create config with echo as editor
	configContent := `editor: "echo"
includes:
  - "*.md"
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}

	templateName := "prompt-template"
	templatePath := filepath.Join(templatesDir, templateName)

	// Simulate user typing "y" to confirm creation
	opts := &EditOptions{
		Stdin: strings.NewReader("y\n"),
	}
	cmd := NewEditCmdWithOptions(templatesDir, configDir, opts)
	cmd.SetArgs([]string{templateName})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify template was created
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Error("template directory should be created when user confirms")
	}

	output := buf.String()
	if !strings.Contains(output, "does not exist") {
		t.Errorf("output should ask about non-existent template, got:\n%s", output)
	}
	if !strings.Contains(output, "Create it?") {
		t.Errorf("output should ask to create, got:\n%s", output)
	}
	if !strings.Contains(output, "Created new template") {
		t.Errorf("output should indicate template was created, got:\n%s", output)
	}
}

func TestEditCmdCreateDeclined(t *testing.T) {
	templatesDir := t.TempDir()
	configDir := t.TempDir()

	templateName := "declined-template"
	templatePath := filepath.Join(templatesDir, templateName)

	// Simulate user typing "n" to decline creation
	opts := &EditOptions{
		Stdin: strings.NewReader("n\n"),
	}
	cmd := NewEditCmdWithOptions(templatesDir, configDir, opts)
	cmd.SetArgs([]string{templateName})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when user declines creation")
	}

	// Verify template was NOT created
	if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
		t.Error("template directory should NOT be created when user declines")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should indicate not found, got: %v", err)
	}
}
