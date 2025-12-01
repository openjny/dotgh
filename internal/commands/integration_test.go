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
		"AGENTS.md":             "# My Agents",
		".github/instructions":  "# Instructions",
		".vscode/settings.json": `{"editor.formatOnSave": true}`,
	})

	templatesDir := t.TempDir()
	templateName := "integration-test-template"

	// Step 1: Push the template
	_, err := executePushCmd(t, templatesDir, sourceDir, templateName, false)
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

// TestPushThenApplyIntegration verifies the push → apply workflow.
func TestPushThenApplyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup source directory with files
	agentsContent := "# My Custom Agents Config"
	githubContent := "# GitHub Instructions"
	vscodeContent := `{"editor.tabSize": 4}`

	sourceDir := setupTestSourceDir(t, map[string]string{
		"AGENTS.md":             agentsContent,
		".github/instructions":  githubContent,
		".vscode/settings.json": vscodeContent,
	})

	templatesDir := t.TempDir()
	templateName := "full-workflow-template"

	// Step 1: Push template from source directory
	_, err := executePushCmd(t, templatesDir, sourceDir, templateName, false)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Step 2: Apply template to a new target directory
	targetDir := t.TempDir()
	_, err = executeApplyCmd(t, templatesDir, targetDir, templateName, false)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Step 3: Verify files were copied correctly
	verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), agentsContent)
	verifyFileContent(t, filepath.Join(targetDir, ".github/instructions"), githubContent)
	verifyFileContent(t, filepath.Join(targetDir, ".vscode/settings.json"), vscodeContent)
}

// TestApplyThenDeleteIntegration verifies the apply → delete workflow.
func TestApplyThenDeleteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup: create a template
	templatesDir := setupTestTemplateWithFiles(t, "deletable-template", map[string]string{
		"AGENTS.md": "# Content",
	})
	templateName := "deletable-template"

	// Step 1: Apply template to target directory
	targetDir := t.TempDir()
	_, err := executeApplyCmd(t, templatesDir, targetDir, templateName, false)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify file exists in target
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Fatal("AGENTS.md should exist after apply")
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

	// Step 4: Applied files should still exist (delete only removes template, not applied files)
	if _, err := os.Stat(filepath.Join(targetDir, "AGENTS.md")); os.IsNotExist(err) {
		t.Error("applied files should persist after template deletion")
	}
}

// TestForceOverwriteIntegration verifies the -f flag behavior across push and apply.
func TestForceOverwriteIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	templatesDir := t.TempDir()
	templateName := "overwrite-test"

	// Step 1: Push initial version
	sourceDir1 := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Version 1",
	})
	_, err := executePushCmd(t, templatesDir, sourceDir1, templateName, false)
	if err != nil {
		t.Fatalf("initial push failed: %v", err)
	}

	// Step 2: Push updated version without force (should skip)
	sourceDir2 := setupTestSourceDir(t, map[string]string{
		"AGENTS.md": "# Version 2",
	})
	output, err := executePushCmd(t, templatesDir, sourceDir2, templateName, false)
	if err != nil {
		t.Fatalf("second push failed: %v", err)
	}
	if !strings.Contains(output, "skipped") {
		t.Errorf("should skip existing file without force, got:\n%s", output)
	}

	// Verify still version 1
	verifyFileContent(t, filepath.Join(templatesDir, templateName, "AGENTS.md"), "# Version 1")

	// Step 3: Push with force (should overwrite)
	_, err = executePushCmd(t, templatesDir, sourceDir2, templateName, true)
	if err != nil {
		t.Fatalf("force push failed: %v", err)
	}

	// Verify now version 2
	verifyFileContent(t, filepath.Join(templatesDir, templateName, "AGENTS.md"), "# Version 2")

	// Step 4: Apply to target directory
	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(targetDir, "AGENTS.md"), []byte("# Existing"), 0644); err != nil {
		t.Fatal(err)
	}

	// Apply without force (should skip)
	output, err = executeApplyCmd(t, templatesDir, targetDir, templateName, false)
	if err != nil {
		t.Fatalf("apply without force failed: %v", err)
	}
	if !strings.Contains(output, "skipped") {
		t.Errorf("should skip existing file without force, got:\n%s", output)
	}

	// Verify still existing content
	verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), "# Existing")

	// Apply with force (should overwrite)
	_, err = executeApplyCmd(t, templatesDir, targetDir, templateName, true)
	if err != nil {
		t.Fatalf("apply with force failed: %v", err)
	}

	// Verify now version 2
	verifyFileContent(t, filepath.Join(targetDir, "AGENTS.md"), "# Version 2")
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
				".github/workflows/go.yml":        "name: Go CI",
				".vscode/settings.json":           `{"go.lintTool": "golangci-lint"}`,
			},
		},
		{
			name: "node-template",
			files: map[string]string{
				"AGENTS.md":                        "# Node Agents",
				".github/workflows/node.yml":       "name: Node CI",
				".vscode/settings.json":            `{"editor.defaultFormatter": "esbenp.prettier-vscode"}`,
			},
		},
		{
			name: "python-template",
			files: map[string]string{
				"AGENTS.md":                          "# Python Agents",
				".github/workflows/python.yml":       "name: Python CI",
				".vscode/settings.json":              `{"python.linting.enabled": true}`,
			},
		},
	}

	// Push all templates
	for _, tmpl := range templates {
		sourceDir := setupTestSourceDir(t, tmpl.files)
		_, err := executePushCmd(t, templatesDir, sourceDir, tmpl.name, false)
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

	// Apply each template to separate directories and verify
	for _, tmpl := range templates {
		targetDir := t.TempDir()
		_, err := executeApplyCmd(t, templatesDir, targetDir, tmpl.name, false)
		if err != nil {
			t.Fatalf("apply %s failed: %v", tmpl.name, err)
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
		".vscode/extensions.json":         `{"recommendations": ["golang.go"]}`,
	})

	// 1. Push
	pushOutput, err := executePushCmd(t, templatesDir, sourceDir, templateName, false)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}
	if !strings.Contains(pushOutput, "Pushing to template") {
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

	// 3. Apply
	targetDir := t.TempDir()
	applyOutput, err := executeApplyCmd(t, templatesDir, targetDir, templateName, false)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}
	if !strings.Contains(applyOutput, "Applying template") {
		t.Errorf("apply output unexpected: %s", applyOutput)
	}

	// Verify applied files
	expectedFiles := []string{
		"AGENTS.md",
		".github/copilot-instructions.md",
		".vscode/extensions.json",
	}
	for _, file := range expectedFiles {
		path := filepath.Join(targetDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist after apply", file)
		}
	}

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
	for _, file := range expectedFiles {
		path := filepath.Join(targetDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("applied file %s should persist after template deletion", file)
		}
	}
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
