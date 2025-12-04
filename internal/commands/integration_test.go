package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Integration tests verify that multiple commands work correctly together.
// These tests use the naming convention TestXxxIntegration and can be skipped
// with `go test -short`.

// TestPushThenListIntegration verifies that a pushed template appears in list output.
func TestPushThenListIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup: create source directory with target files
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# My Agents",
		".github/copilot-instructions.md": "# Instructions",
		".vscode/mcp.json":                `{"servers": {}}`,
	})

	templatesDir := t.TempDir()
	templateName := "integration-test-template"

	// Step 1: Push the template (with --yes to skip confirmation)
	_, err := executePushCmd(t, templatesDir, sourceDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Step 2: List templates and verify the pushed template appears
	output, err := executeListCmd(t, templatesDir)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	if !strings.Contains(output, templateName) {
		t.Errorf("pushed template should appear in list output, got:\n%s", output)
	}
	if !strings.Contains(output, "1 template(s) found") {
		t.Errorf("should show 1 template found, got:\n%s", output)
	}
}

// TestPushThenPullIntegration verifies the push → pull workflow.
func TestPushThenPullIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup source directory with files
	agentsContent := "# My Custom Agents Config"
	githubContent := "# GitHub Instructions"
	vscodeContent := `{"servers": {}}`

	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       agentsContent,
		".github/copilot-instructions.md": githubContent,
		".vscode/mcp.json":                vscodeContent,
	})

	templatesDir := t.TempDir()
	templateName := "full-workflow-template"

	// Step 1: Push template from source directory (with --yes to skip confirmation)
	_, err := executePushCmd(t, templatesDir, sourceDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Step 2: Pull template to a new target directory (with --yes to skip confirmation)
	targetDir := t.TempDir()
	_, err = executePullCmd(t, templatesDir, targetDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("pull failed: %v", err)
	}

	// Step 3: Verify files were copied correctly
	verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), agentsContent)
	verifyFileContent(t, filepath.Join(targetDir, ".github/copilot-instructions.md"), githubContent)
	verifyFileContent(t, filepath.Join(targetDir, ".vscode/mcp.json"), vscodeContent)
}

// TestPullThenDeleteIntegration verifies the pull → delete workflow.
func TestPullThenDeleteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup: create a template
	templatesDir := setupTestTemplateWithFiles(t, "deletable-template", map[string]string{
		"AGENTS.md": "# Content",
	})
	templateName := "deletable-template"

	// Step 1: Pull template to target directory (with --yes to skip confirmation)
	targetDir := t.TempDir()
	_, err := executePullCmd(t, templatesDir, targetDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("pull failed: %v", err)
	}

	// Verify file exists in target
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Fatal("AGENTS.md should exist after pull")
	}

	// Step 2: Delete the template (with force flag to skip confirmation)
	_, err = executeDeleteCmd(t, templatesDir, templateName, "", true)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Step 3: Verify template is deleted
	templatePath := filepath.Join(templatesDir, templateName)
	if _, err := os.Stat(templatePath); !os.IsNotExist(err) {
		t.Errorf("template directory should be deleted")
	}

	// Step 4: Pulled files should still exist (delete only removes template, not pulled files)
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("pulled files should persist after template deletion")
	}
}

// TestMergeModeIntegration verifies the --merge flag behavior across push and pull.
func TestMergeModeIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	templatesDir := t.TempDir()
	templateName := "merge-test"

	// Step 1: Push initial version
	sourceDir1 := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Version 1",
	})
	_, err := executePushCmd(t, templatesDir, sourceDir1, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("initial push failed: %v", err)
	}

	// Step 2: Push updated version (should update)
	sourceDir2 := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Version 2",
	})
	output, err := executePushCmd(t, templatesDir, sourceDir2, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("second push failed: %v", err)
	}
	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("should show modification, got:\n%s", output)
	}

	// Verify now version 2
	verifyFileContent(t, filepath.Join(templatesDir, templateName, "AGENTS.md"), "# Version 2")

	// Step 3: Pull to target directory with existing different file
	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(targetDir, "AGENTS.md"), []byte("# Existing"), 0644); err != nil {
		t.Fatal(err)
	}
	// Add a file that's only in target (will be deleted in full sync)
	if err := os.MkdirAll(filepath.Join(targetDir, ".github"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, ".github/copilot-instructions.md"), []byte("# Local only"), 0644); err != nil {
		t.Fatal(err)
	}

	// Pull with merge mode (should NOT delete local-only file)
	output, err = executePullCmd(t, templatesDir, targetDir, templateName, true, true, nil, "")
	if err != nil {
		t.Fatalf("pull with merge failed: %v", err)
	}
	if strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("merge mode should NOT show deletion, got:\n%s", output)
	}

	// Verify local-only file still exists
	verifyFileContent(t, filepath.Join(targetDir, ".github/copilot-instructions.md"), "# Local only")

	// Verify AGENTS.md was updated
	verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), "# Version 2")

	// Pull without merge mode (full sync should delete local-only file)
	_, err = executePullCmd(t, templatesDir, targetDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("pull without merge failed: %v", err)
	}

	// Verify local-only file was deleted
	if _, err := os.Stat(filepath.Join(targetDir, ".github/copilot-instructions.md")); !os.IsNotExist(err) {
		t.Error("full sync should delete local-only file")
	}
}

// TestMultipleTemplatesIntegration verifies managing multiple templates.
func TestMultipleTemplatesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	templatesDir := t.TempDir()

	// Create multiple templates
	templates := []struct {
		name  string
		files map[string]string
	}{
		{
			name: "golang-template",
			files: map[string]string{
				"AGENTS.md":                       "# Go Agents",
				".github/copilot-instructions.md": "# Go Copilot",
				".vscode/mcp.json":                `{"servers": {"go": {}}}`,
			},
		},
		{
			name: "node-template",
			files: map[string]string{
				"AGENTS.md":                       "# Node Agents",
				".github/copilot-instructions.md": "# Node Copilot",
				".vscode/mcp.json":                `{"servers": {"node": {}}}`,
			},
		},
		{
			name: "python-template",
			files: map[string]string{
				"AGENTS.md":                       "# Python Agents",
				".github/copilot-instructions.md": "# Python Copilot",
				".vscode/mcp.json":                `{"servers": {"python": {}}}`,
			},
		},
	}

	// Push all templates
	for _, tmpl := range templates {
		sourceDir := setupTestSourceDir(t, tmpl.files)
		_, err := executePushCmd(t, templatesDir, sourceDir, tmpl.name, false, true, nil, "")
		if err != nil {
			t.Fatalf("push %s failed: %v", tmpl.name, err)
		}
	}

	// List and verify all appear
	output, err := executeListCmd(t, templatesDir)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	for _, tmpl := range templates {
		if !strings.Contains(output, tmpl.name) {
			t.Errorf("template %s should appear in list, got:\n%s", tmpl.name, output)
		}
	}
	if !strings.Contains(output, "3 template(s) found") {
		t.Errorf("should show 3 templates found, got:\n%s", output)
	}

	// Pull each template to separate directories and verify
	for _, tmpl := range templates {
		targetDir := t.TempDir()
		_, err := executePullCmd(t, templatesDir, targetDir, tmpl.name, false, true, nil, "")
		if err != nil {
			t.Fatalf("pull %s failed: %v", tmpl.name, err)
		}

		// Verify AGENTS.md content is correct
		verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), tmpl.files["AGENTS.md"])
	}

	// Delete one template and verify list updates
	_, err = executeDeleteCmd(t, templatesDir, "node-template", "", true)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	output, err = executeListCmd(t, templatesDir)
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}

	if strings.Contains(output, "node-template") {
		t.Error("deleted template should not appear in list")
	}
	if !strings.Contains(output, "2 template(s) found") {
		t.Errorf("should show 2 templates found after delete, got:\n%s", output)
	}
}

// TestFullWorkflowIntegration tests the complete push → list → apply → delete flow.
func TestFullWorkflowIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	templatesDir := t.TempDir()
	templateName := "complete-workflow"

	// Setup source
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# Complete Workflow Agents",
		".github/copilot-instructions.md": "# Copilot Instructions",
		".vscode/mcp.json":                `{"servers": {}}`,
	})

	// 1. Push
	pushOutput, err := executePushCmd(t, templatesDir, sourceDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if !strings.Contains(pushOutput, "Creating template") && !strings.Contains(pushOutput, "Pushing to template") {
		t.Errorf("push output unexpected: %s", pushOutput)
	}

	// 2. List
	listOutput, err := executeListCmd(t, templatesDir)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(listOutput, templateName) {
		t.Errorf("template not in list: %s", listOutput)
	}

	// 3. Pull
	targetDir := t.TempDir()
	pullOutput, err := executePullCmd(t, templatesDir, targetDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("pull failed: %v", err)
	}
	if !strings.Contains(pullOutput, "Pulling template") {
		t.Errorf("pull output unexpected: %s", pullOutput)
	}

	// Verify pulled files
	expectedFiles := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".vscode/mcp.json",
	}
	verifyFilesExist(t, targetDir, expectedFiles)

	// 4. Delete
	deleteOutput, err := executeDeleteCmd(t, templatesDir, templateName, "", true)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if !strings.Contains(deleteOutput, "deleted") {
		t.Errorf("delete output unexpected: %s", deleteOutput)
	}

	// Verify template is gone
	listOutput, err = executeListCmd(t, templatesDir)
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if strings.Contains(listOutput, templateName) {
		t.Error("deleted template should not appear in list")
	}

	// Applied files should still exist
	verifyFilesExist(t, targetDir, expectedFiles)
}

// verifyFileContent checks that a file exists and has the expected content.
func verifyFileContent(t *testing.T, path, expectedContent string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	if string(content) != expectedContent {
		t.Errorf("content mismatch for %s:\nexpected: %s\ngot: %s", path, expectedContent, string(content))
	}
}

// verifyFilesExist checks that all specified files exist in the given directory.
func verifyFilesExist(t *testing.T, dir string, files []string) {
	t.Helper()
	for _, file := range files {
		path := filepath.Join(dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", file)
		}
	}
}

// TestEditTemplateNotFoundIntegration verifies that edit command returns an error
// for non-existent templates.
func TestEditTemplateNotFoundIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	templatesDir := t.TempDir()
	configDir := t.TempDir()

	// Try to edit a template that doesn't exist
	_, err := executeEditCmd(t, templatesDir, configDir, "non-existent")
	if err == nil {
		t.Error("expected error for non-existent template")
	}

	expectedMsg := `template "non-existent" not found`
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error containing %q, got %q", expectedMsg, err.Error())
	}
}

// TestPushThenEditIntegration verifies that a pushed template can be validated for editing.
// Note: We can't actually launch an editor in tests, so we verify the template path validation.
func TestPushThenEditIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup source directory with files
	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":                       "# Editable Agents",
		".github/copilot-instructions.md": "# Instructions",
	})

	templatesDir := t.TempDir()
	templateName := "editable-template"

	// Step 1: Push the template
	_, err := executePushCmd(t, templatesDir, sourceDir, templateName, false, true, nil, "")
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Step 2: Verify template path is valid for edit command
	templatePath, err := getTemplatePath(templatesDir, templateName)
	if err != nil {
		t.Fatalf("getTemplatePath should succeed after push: %v", err)
	}

	expectedPath := filepath.Join(templatesDir, templateName)
	if templatePath != expectedPath {
		t.Errorf("expected template path %q, got %q", expectedPath, templatePath)
	}

	// Verify template directory contains expected files
	verifyFilesExist(t, templatePath, []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
	})
}

// executeEditCmd executes the edit command and returns output.
// Note: This only validates the template exists; it doesn't actually launch an editor.
func executeEditCmd(t *testing.T, templatesDir, configDir, templateName string) (string, error) {
	t.Helper()

	// We can't execute the full command as it would launch an editor,
	// so we just validate the template path
	_, err := getTemplatePath(templatesDir, templateName)
	if err != nil {
		return "", err
	}

	return "template validated", nil
}
