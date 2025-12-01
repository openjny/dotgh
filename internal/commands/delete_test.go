package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// executeDeleteCmd runs the delete command and returns the output.
// stdinInput simulates user input for the confirmation prompt.
func executeDeleteCmd(t *testing.T, templatesDir, templateName, stdinInput string, force bool) (string, error) {
	t.Helper()
	stdin := strings.NewReader(stdinInput)
	cmd := NewDeleteCmd(templatesDir, stdin)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	args := []string{templateName}
	if force {
		args = append(args, "--force")
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestDeleteTemplateWithForce(t *testing.T) {
	// Setup template
	templatesDir := setupTestTemplatesDir(t, []string{"my-template"})
	templateDir := filepath.Join(templatesDir, "my-template")

	// Verify template exists before delete
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Fatal("template should exist before delete")
	}

	output, err := executeDeleteCmd(t, templatesDir, "my-template", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check output
	if !strings.Contains(output, "deleted") {
		t.Errorf("output should indicate template was deleted, got:\n%s", output)
	}

	// Verify template was deleted
	if _, err := os.Stat(templateDir); !os.IsNotExist(err) {
		t.Error("template directory should be deleted")
	}
}

func TestDeleteTemplateWithConfirmYes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase y", "y\n"},
		{"uppercase Y", "Y\n"},
		{"lowercase yes", "yes\n"},
		{"uppercase YES", "YES\n"},
		{"mixed case Yes", "Yes\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templatesDir := setupTestTemplatesDir(t, []string{"confirm-template"})
			templateDir := filepath.Join(templatesDir, "confirm-template")

			output, err := executeDeleteCmd(t, templatesDir, "confirm-template", tt.input, false)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(output, "deleted") {
				t.Errorf("output should indicate template was deleted, got:\n%s", output)
			}

			if _, err := os.Stat(templateDir); !os.IsNotExist(err) {
				t.Error("template directory should be deleted")
			}
		})
	}
}

func TestDeleteTemplateWithConfirmNo(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase n", "n\n"},
		{"uppercase N", "N\n"},
		{"lowercase no", "no\n"},
		{"uppercase NO", "NO\n"},
		{"empty input (default)", "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templatesDir := setupTestTemplatesDir(t, []string{"keep-template"})
			templateDir := filepath.Join(templatesDir, "keep-template")

			output, err := executeDeleteCmd(t, templatesDir, "keep-template", tt.input, false)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(output, "cancelled") {
				t.Errorf("output should indicate deletion was cancelled, got:\n%s", output)
			}

			// Template should still exist
			if _, err := os.Stat(templateDir); os.IsNotExist(err) {
				t.Error("template directory should NOT be deleted")
			}
		})
	}
}

func TestDeleteTemplateNotFound(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{})

	_, err := executeDeleteCmd(t, templatesDir, "non-existent", "", true)

	if err == nil {
		t.Error("expected error for non-existent template")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should indicate template not found, got: %v", err)
	}
}

func TestDeleteRequiresTemplateName(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{})
	stdin := strings.NewReader("")
	cmd := NewDeleteCmd(templatesDir, stdin)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{}) // No template name

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error when no template name provided")
	}
}

func TestDeleteTemplateWithContents(t *testing.T) {
	// Create template with files
	templatesDir := setupTestTemplateWithFiles(t, "full-template", map[string]string{
		".github/copilot-instructions.md": "# Instructions",
		".vscode/settings.json":           "{}",
		"AGENTS.md":                       "# Agents",
	})
	templateDir := filepath.Join(templatesDir, "full-template")

	output, err := executeDeleteCmd(t, templatesDir, "full-template", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "deleted") {
		t.Errorf("output should indicate template was deleted, got:\n%s", output)
	}

	// Verify entire directory tree was deleted
	if _, err := os.Stat(templateDir); !os.IsNotExist(err) {
		t.Error("template directory and all contents should be deleted")
	}
}

func TestDeleteShowsPromptMessage(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{"prompt-test"})

	output, _ := executeDeleteCmd(t, templatesDir, "prompt-test", "n\n", false)

	// Should show confirmation prompt
	if !strings.Contains(output, "Delete template 'prompt-test'?") {
		t.Errorf("output should show confirmation prompt, got:\n%s", output)
	}
}
