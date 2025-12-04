package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openjny/dotgh/internal/config"
)

// setupTestTemplateWithFiles creates a template with the specified files/directories.
// files is a map of relative path to content (empty string for directories).
// This builds on top of setupTestTemplatesDir from list_test.go.
func setupTestTemplateWithFiles(t *testing.T, templateName string, files map[string]string) string {
	t.Helper()
	// Use the shared helper to create the base templates directory
	templatesDir := setupTestTemplatesDir(t, []string{templateName})
	templateDir := filepath.Join(templatesDir, templateName)
	createTestFiles(t, templateDir, files)
	return templatesDir
}

// executePullCmd runs the pull command and returns the output.
func executePullCmd(t *testing.T, templatesDir, targetDir, templateName string, merge, yes bool, excludes []string, stdin string) (string, error) {
	t.Helper()
	var cfg *config.Config
	if excludes == nil {
		cfg = testConfig()
	} else {
		cfg = testConfigWithExcludes(excludes)
	}

	opts := &PullOptions{
		Stdin: strings.NewReader(stdin),
	}

	cmd := NewPullCmdWithOptions(templatesDir, targetDir, cfg, opts)
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

func TestPullWithYesFlag(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := t.TempDir()

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "+ AGENTS.md") {
		t.Errorf("output should show addition, got:\n%s", output)
	}
	if !strings.Contains(output, "Done:") {
		t.Errorf("output should show done message, got:\n%s", output)
	}

	// Verify file was created
	content, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("file should exist: %v", err)
	}
	if string(content) != "# Agents" {
		t.Errorf("content mismatch: got %s", string(content))
	}
}

func TestPullWithConfirmationYes(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := t.TempDir()

	// Simulate user typing "y"
	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, false, nil, "y\n")

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
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("file should exist after confirmation")
	}
}

func TestPullWithConfirmationNo(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := t.TempDir()

	// Simulate user typing "n"
	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, false, nil, "n\n")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Aborted") {
		t.Errorf("output should show aborted message, got:\n%s", output)
	}

	// Verify file was NOT created
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); !os.IsNotExist(err) {
		t.Error("file should NOT exist after abortion")
	}
}

func TestPullFullSync(t *testing.T) {
	// Template has one file, target has different file
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Template Agents",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		".github/copilot-instructions.md": "# Will be deleted",
	})

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should show deletion
	if !strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("output should show deletion, got:\n%s", output)
	}

	// Verify file was deleted
	if _, err := os.Stat(filepath.Join(targetDir, ".github/copilot-instructions.md")); !os.IsNotExist(err) {
		t.Error("file should be deleted in full sync mode")
	}

	// Verify template file was added
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("template file should be added")
	}
}

func TestPullMergeMode(t *testing.T) {
	// Template has one file, target has different file
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Template Agents",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		".github/copilot-instructions.md": "# Should be kept",
	})

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", true, true, nil, "")

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

	// Verify file was NOT deleted
	content, err := os.ReadFile(filepath.Join(targetDir, ".github/copilot-instructions.md"))
	if err != nil {
		t.Fatalf("file should still exist: %v", err)
	}
	if string(content) != "# Should be kept" {
		t.Errorf("content should be preserved: got %s", string(content))
	}

	// Verify template file was added
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("template file should be added")
	}
}

func TestPullAlreadyInSync(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Same Content",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md": "# Same Content",
	})

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "already in sync") {
		t.Errorf("output should indicate already in sync, got:\n%s", output)
	}
}

func TestPullModifiesExistingFiles(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# New Content",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md": "# Old Content",
	})

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification, got:\n%s", output)
	}

	// Verify content was updated
	content, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "# New Content" {
		t.Errorf("content should be updated: got %s", string(content))
	}
}

func TestPullTemplateNotFound(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{})
	targetDir := t.TempDir()

	_, err := executePullCmd(t, templatesDir, targetDir, "non-existent", false, true, nil, "")

	if err == nil {
		t.Error("expected error for non-existent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention not found: %v", err)
	}
}

func TestPullRequiresTemplateName(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{"my-template"})
	targetDir := t.TempDir()

	cmd := NewPullCmd(templatesDir, targetDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{}) // No template name

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error when no template name provided")
	}
}

func TestPullWithExcludes(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/prompts/test.prompt.md":  "# Test",
		".github/prompts/local.prompt.md": "# Local - should be excluded",
	})
	targetDir := t.TempDir()

	excludes := []string{".github/prompts/local.prompt.md"}
	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, excludes, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should NOT show excluded file
	if strings.Contains(output, "local.prompt.md") {
		t.Errorf("output should NOT show excluded file, got:\n%s", output)
	}

	// Verify excluded file was NOT created
	if _, err := os.Stat(filepath.Join(targetDir, ".github/prompts/local.prompt.md")); !os.IsNotExist(err) {
		t.Error("excluded file should NOT be created")
	}

	// Verify non-excluded files were created
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("AGENTS.md should be created")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".github/prompts/test.prompt.md")); os.IsNotExist(err) {
		t.Error("test.prompt.md should be created")
	}
}

func TestPullMultipleFiles(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/copilot-instructions.md": "# Instructions",
		".vscode/mcp.json":                `{"servers": {}}`,
	})
	targetDir := t.TempDir()

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "3 added") {
		t.Errorf("output should show 3 additions, got:\n%s", output)
	}

	// Verify all files were created
	expectedFiles := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".vscode/mcp.json",
	}
	for _, file := range expectedFiles {
		if _, err := os.Stat(filepath.Join(targetDir, file)); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestPullMixedChanges(t *testing.T) {
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# New Agents",   // Modified
		".github/copilot-instructions.md": "# Instructions", // Added
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md":        "# Old Agents",
		".vscode/mcp.json": "{}", // Will be deleted
	})

	output, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, true, nil, "")

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
