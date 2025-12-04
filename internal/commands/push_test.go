package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openjny/dotgh/internal/config"
)

// setupTestSourceDir creates a source directory with the specified files.
// files is a map of relative path to content.
func setupTestSourceDir(t *testing.T, files map[string]string) string {
	t.Helper()
	sourceDir := t.TempDir()
	createTestFiles(t, sourceDir, files)
	return sourceDir
}

// executePushCmd runs the push command and returns the output.
func executePushCmd(t *testing.T, templatesDir, sourceDir, templateName string, merge, yes bool, excludes []string, stdin string) (string, error) {
	t.Helper()
	var cfg *config.Config
	if excludes == nil {
		cfg = testConfig()
	} else {
		cfg = testConfigWithExcludes(excludes)
	}

	opts := &PushOptions{
		Stdin: strings.NewReader(stdin),
	}

	cmd := NewPushCmdWithOptions(templatesDir, sourceDir, cfg, opts)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	args := []string{templateName}
	if merge {
		args = append(args, "--merge")
	}
	if yes {
		args = append(args, "--yes")
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestPushNewTemplate(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# My Agents",
		".github/copilot-instructions.md": "# Instructions",
		".vscode/mcp.json":                `{"servers": {}}`,
	})

	templatesDir := t.TempDir()

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check output
	if !strings.Contains(output, "Creating template 'my-template'") {
		t.Errorf("output should indicate creating new template, got:\n%s", output)
	}
	if !strings.Contains(output, "Done:") {
		t.Errorf("output should show done message, got:\n%s", output)
	}

	// Check template was created
	templateDir := filepath.Join(templatesDir, "my-template")
	if _, err := os.Stat(templateDir); os.IsNotExist(err) {
		t.Error("template directory should be created")
	}

	// Check files were copied
	expectedFiles := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".vscode/mcp.json",
	}
	for _, file := range expectedFiles {
		fullPath := filepath.Join(templateDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist in template", file)
		}
	}
}

func TestPushWithConfirmationYes(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# My Agents",
	})

	templatesDir := t.TempDir()

	// Simulate user typing "y"
	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, false, nil, "y\n")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Apply these changes?") {
		t.Errorf("output should ask for confirmation, got:\n%s", output)
	}
	if !strings.Contains(output, "Done:") {
		t.Errorf("output should show done message, got:\n%s", output)
	}

	// Verify file was created
	templateDir := filepath.Join(templatesDir, "my-template")
	if _, err := os.Stat(filepath.Join(templateDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("file should exist after confirmation")
	}
}

func TestPushWithConfirmationNo(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# My Agents",
	})

	templatesDir := t.TempDir()

	// Simulate user typing "n"
	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, false, nil, "n\n")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Aborted") {
		t.Errorf("output should show aborted message, got:\n%s", output)
	}

	// Verify template was NOT created
	templateDir := filepath.Join(templatesDir, "my-template")
	if _, err := os.Stat(templateDir); !os.IsNotExist(err) {
		t.Error("template directory should NOT exist after abortion")
	}
}

func TestPushFullSync(t *testing.T) {
	// Source has one file, template has different file
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Source Agents",
	})

	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		".github/copilot-instructions.md": "# Will be deleted from template",
	})

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show deletion
	if !strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("output should show deletion, got:\n%s", output)
	}

	templateDir := filepath.Join(templatesDir, "my-template")

	// Verify file was deleted from template
	if _, err := os.Stat(filepath.Join(templateDir, ".github/copilot-instructions.md")); !os.IsNotExist(err) {
		t.Error("file should be deleted in full sync mode")
	}

	// Verify source file was added to template
	if _, err := os.Stat(filepath.Join(templateDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("source file should be added to template")
	}
}

func TestPushMergeMode(t *testing.T) {
	// Source has one file, template has different file
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Source Agents",
	})

	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		".github/copilot-instructions.md": "# Should be kept in merge mode",
	})

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", true, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT show deletion in merge mode
	if strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("merge mode should NOT show deletion, got:\n%s", output)
	}
	if !strings.Contains(output, "merge") {
		t.Errorf("output should indicate merge mode, got:\n%s", output)
	}

	templateDir := filepath.Join(templatesDir, "my-template")

	// Verify file was NOT deleted
	content, err := os.ReadFile(filepath.Join(templateDir, ".github/copilot-instructions.md"))
	if err != nil {
		t.Fatalf("file should still exist: %v", err)
	}
	if string(content) != "# Should be kept in merge mode" {
		t.Errorf("content should be preserved: got %s", string(content))
	}

	// Verify source file was added
	if _, err := os.Stat(filepath.Join(templateDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("source file should be added to template")
	}
}

func TestPushAlreadyInSync(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Same Content",
	})

	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Same Content",
	})

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "already in sync") {
		t.Errorf("output should indicate already in sync, got:\n%s", output)
	}
}

func TestPushModifiesExistingFiles(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# New Content",
	})

	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Old Content",
	})

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification, got:\n%s", output)
	}

	// Verify content was updated
	templateDir := filepath.Join(templatesDir, "my-template")
	content, err := os.ReadFile(filepath.Join(templateDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# New Content" {
		t.Errorf("content should be updated: got %s", string(content))
	}
}

func TestPushNoTargetsFound(t *testing.T) {
	// Empty source directory
	sourceDir := t.TempDir()
	templatesDir := t.TempDir()

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should indicate already in sync (no files to sync)
	if !strings.Contains(output, "already in sync") {
		t.Errorf("output should indicate already in sync when no files, got:\n%s", output)
	}
}

func TestPushRequiresTemplateName(t *testing.T) {
	sourceDir := t.TempDir()
	templatesDir := t.TempDir()

	cmd := NewPushCmd(templatesDir, sourceDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{}) // No template name

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error when no template name provided")
	}
}

func TestPushWithExcludes(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/prompts/test.prompt.md":  "# Test",
		".github/prompts/local.prompt.md": "# Local - should be excluded",
	})

	templatesDir := t.TempDir()

	excludes := []string{".github/prompts/local.prompt.md"}
	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, excludes, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT show excluded file
	if strings.Contains(output, "local.prompt.md") {
		t.Errorf("output should NOT show excluded file, got:\n%s", output)
	}

	templateDir := filepath.Join(templatesDir, "my-template")

	// Verify excluded file was NOT created
	if _, err := os.Stat(filepath.Join(templateDir, ".github/prompts/local.prompt.md")); !os.IsNotExist(err) {
		t.Error("excluded file should NOT be created in template")
	}

	// Verify non-excluded files were created
	if _, err := os.Stat(filepath.Join(templateDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("AGENTS.md should be created in template")
	}
	if _, err := os.Stat(filepath.Join(templateDir, ".github/prompts/test.prompt.md")); os.IsNotExist(err) {
		t.Error("test.prompt.md should be created in template")
	}
}

func TestPushPreservesFileContent(t *testing.T) {
	expectedContent := map[string]string{
		"AGENTS.md":                       "# Agents\n\nSome content here",
		".github/copilot-instructions.md": "Line 1\nLine 2\nLine 3",
		".vscode/mcp.json":                `{"key": "value", "nested": {"a": 1}}`,
	}

	sourceDir := setupTestSourceDir(t, expectedContent)
	templatesDir := t.TempDir()

	_, err := executePushCmd(t, templatesDir, sourceDir, "content-test", false, true, nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify each file's content
	templateDir := filepath.Join(templatesDir, "content-test")
	for file, expected := range expectedContent {
		content, err := os.ReadFile(filepath.Join(templateDir, file))
		if err != nil {
			t.Errorf("failed to read %s: %v", file, err)
			continue
		}
		if string(content) != expected {
			t.Errorf("content mismatch for %s:\nexpected: %s\ngot: %s", file, expected, string(content))
		}
	}
}

func TestPushMixedChanges(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# New Agents",   // Modified
		".github/copilot-instructions.md": "# Instructions", // Added
	})

	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":        "# Old Agents",
		".vscode/mcp.json": "{}", // Will be deleted
	})

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all types of changes in output
	if !strings.Contains(output, "+ .github/copilot-instructions.md") {
		t.Errorf("output should show addition, got:\n%s", output)
	}
	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification, got:\n%s", output)
	}
	if !strings.Contains(output, "- .vscode/mcp.json") {
		t.Errorf("output should show deletion, got:\n%s", output)
	}
}
