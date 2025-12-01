package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestSourceDir creates a source directory with the specified files.
// files is a map of relative path to content.
func setupTestSourceDir(t *testing.T, files map[string]string) string {
	t.Helper()
	sourceDir := t.TempDir()

	for path, content := range files {
		fullPath := filepath.Join(sourceDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
	}

	return sourceDir
}

// executePushCmd runs the push command and returns the output.
func executePushCmd(t *testing.T, templatesDir, sourceDir, templateName string, force bool) (string, error) {
	t.Helper()
	cmd := NewPushCmd(templatesDir, sourceDir)
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

func TestPushNewTemplate(t *testing.T) {
	// Setup source directory with target files
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# My Agents",
		".github/copilot-instructions.md": "# Instructions",
		".vscode/settings.json":           `{"editor.formatOnSave": true}`,
	})

	templatesDir := t.TempDir()

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check output
	if !strings.Contains(output, "Pushing to template 'my-template'") {
		t.Errorf("output should contain push message, got:\n%s", output)
	}
	if !strings.Contains(output, "copied") {
		t.Errorf("output should indicate files were copied, got:\n%s", output)
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
		".vscode/settings.json",
	}
	for _, file := range expectedFiles {
		fullPath := filepath.Join(templateDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist in template", file)
		}
	}
}

func TestPushExistingTemplateWithoutForce(t *testing.T) {
	// Setup source directory
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# New Content",
	})

	// Setup existing template
	templatesDir := setupTestTemplatesDir(t, []string{"existing-template"})
	templateDir := filepath.Join(templatesDir, "existing-template")
	existingContent := "# Existing Content"
	if err := os.WriteFile(filepath.Join(templateDir, "AGENTS.md"), []byte(existingContent), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := executePushCmd(t, templatesDir, sourceDir, "existing-template", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should skip existing files
	if !strings.Contains(output, "skipped") {
		t.Errorf("output should indicate files were skipped, got:\n%s", output)
	}

	// Existing content should be preserved
	content, err := os.ReadFile(filepath.Join(templateDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != existingContent {
		t.Errorf("existing content should be preserved, got: %s", string(content))
	}
}

func TestPushExistingTemplateWithForce(t *testing.T) {
	// Setup source directory
	newContent := "# New Content"
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": newContent,
	})

	// Setup existing template
	templatesDir := setupTestTemplatesDir(t, []string{"existing-template"})
	templateDir := filepath.Join(templatesDir, "existing-template")
	if err := os.WriteFile(filepath.Join(templateDir, "AGENTS.md"), []byte("# Old"), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := executePushCmd(t, templatesDir, sourceDir, "existing-template", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should overwrite
	if !strings.Contains(output, "copied") {
		t.Errorf("output should indicate files were copied, got:\n%s", output)
	}

	// Content should be overwritten
	content, err := os.ReadFile(filepath.Join(templateDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != newContent {
		t.Errorf("content should be overwritten, got: %s", string(content))
	}
}

func TestPushNoTargetsFound(t *testing.T) {
	// Empty source directory
	sourceDir := t.TempDir()
	templatesDir := t.TempDir()

	output, err := executePushCmd(t, templatesDir, sourceDir, "my-template", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should indicate no targets found
	if !strings.Contains(output, "No target files found") {
		t.Errorf("output should indicate no targets found, got:\n%s", output)
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

func TestPushWithGitHubDir(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		".github/copilot-instructions.md": "# Instructions",
		".github/workflows/ci.yml":        "name: CI",
		".github/prompts/test.prompt.md":  "# Test Prompt",
	})

	templatesDir := t.TempDir()

	_, err := executePushCmd(t, templatesDir, sourceDir, "github-only", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all .github files were copied
	expectedFiles := []string{
		".github/copilot-instructions.md",
		".github/workflows/ci.yml",
		".github/prompts/test.prompt.md",
	}
	templateDir := filepath.Join(templatesDir, "github-only")
	for _, file := range expectedFiles {
		fullPath := filepath.Join(templateDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestPushWithVSCodeDir(t *testing.T) {
	sourceDir := setupTestSourceDir(t, map[string]string{
		".vscode/settings.json":   `{"editor.formatOnSave": true}`,
		".vscode/extensions.json": `{"recommendations": []}`,
	})

	templatesDir := t.TempDir()

	_, err := executePushCmd(t, templatesDir, sourceDir, "vscode-only", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all .vscode files were copied
	expectedFiles := []string{
		".vscode/settings.json",
		".vscode/extensions.json",
	}
	templateDir := filepath.Join(templatesDir, "vscode-only")
	for _, file := range expectedFiles {
		fullPath := filepath.Join(templateDir, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

func TestPushWithAgentsMdOnly(t *testing.T) {
	agentsContent := "# My Custom Agents"
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": agentsContent,
	})

	templatesDir := t.TempDir()

	output, err := executePushCmd(t, templatesDir, sourceDir, "agents-only", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "AGENTS.md") {
		t.Errorf("output should mention AGENTS.md, got:\n%s", output)
	}

	// Check content
	content, err := os.ReadFile(filepath.Join(templatesDir, "agents-only", "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != agentsContent {
		t.Errorf("content mismatch, got: %s", string(content))
	}
}

func TestPushPreservesFileContent(t *testing.T) {
	// Test that file content is correctly preserved during push
	expectedContent := map[string]string{
		"AGENTS.md":             "# Agents\n\nSome content here",
		".github/instructions":  "Line 1\nLine 2\nLine 3",
		".vscode/settings.json": `{"key": "value", "nested": {"a": 1}}`,
	}

	sourceDir := setupTestSourceDir(t, expectedContent)
	templatesDir := t.TempDir()

	_, err := executePushCmd(t, templatesDir, sourceDir, "content-test", false)
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
