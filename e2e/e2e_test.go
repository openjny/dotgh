//go:build e2e

// Package e2e contains end-to-end tests that run the compiled dotgh binary.
// These tests verify the CLI works correctly from a user's perspective.
//
// Run with: go test -v -tags=e2e ./e2e/
// Note: The binary must be built first with: go build -o dotgh ./cmd/dotgh
package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// binaryName returns the name of the dotgh binary for the current OS.
func binaryName() string {
	if runtime.GOOS == "windows" {
		return "dotgh.exe"
	}
	return "dotgh"
}

// findBinary locates the dotgh binary relative to the test file or in common locations.
func findBinary(t *testing.T) string {
	t.Helper()

	// Try relative path from project root
	candidates := []string{
		binaryName(),                            // current directory
		filepath.Join("..", binaryName()),       // parent directory
		filepath.Join("..", "..", binaryName()), // grandparent
	}

	for _, path := range candidates {
		absPath, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath
		}
	}

	t.Fatalf("dotgh binary not found. Build it first with: go build -o %s ./cmd/dotgh", binaryName())
	return ""
}

// runDotgh executes the dotgh binary with the given arguments and environment.
func runDotgh(t *testing.T, binary string, args []string, workDir string, env map[string]string) (string, string, error) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Dir = workDir

	// Set up environment
	cmd.Env = os.Environ()
	for key, val := range env {
		cmd.Env = append(cmd.Env, key+"="+val)
	}

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// setupE2EEnvironment creates isolated directories for E2E testing.
// Returns templatesDir, workDir, and the environment variables map for XDG_CONFIG_HOME.
func setupE2EEnvironment(t *testing.T) (templatesDir, workDir string, env map[string]string) {
	t.Helper()

	baseDir := t.TempDir()
	templatesDir = filepath.Join(baseDir, "config", "dotgh", "templates")
	workDir = filepath.Join(baseDir, "project")

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}

	// Calculate config dir for XDG_CONFIG_HOME override
	configDir := filepath.Dir(filepath.Dir(templatesDir))
	env = map[string]string{
		"XDG_CONFIG_HOME": configDir,
	}

	return templatesDir, workDir, env
}

// createTestFiles creates files in the given directory.
func createTestFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", path, err)
		}
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

// TestE2E_VersionCommand verifies the version command works.
func TestE2E_VersionCommand(t *testing.T) {
	binary := findBinary(t)

	stdout, _, err := runDotgh(t, binary, []string{"version"}, "", nil)
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	// Version output should contain "dotgh" and version info
	if !strings.Contains(stdout, "dotgh") {
		t.Errorf("version output should contain 'dotgh', got: %s", stdout)
	}
}

// TestE2E_HelpCommand verifies the help command works.
func TestE2E_HelpCommand(t *testing.T) {
	binary := findBinary(t)

	stdout, _, err := runDotgh(t, binary, []string{"--help"}, "", nil)
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	// Help should list available commands
	expectedCommands := []string{"list", "apply", "push", "delete", "update", "version"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("help output should contain '%s' command, got: %s", cmd, stdout)
		}
	}
}

// TestE2E_ListEmptyTemplates verifies list command with no templates.
func TestE2E_ListEmptyTemplates(t *testing.T) {
	binary := findBinary(t)
	_, _, env := setupE2EEnvironment(t)

	stdout, _, err := runDotgh(t, binary, []string{"list"}, "", env)
	if err != nil {
		t.Fatalf("list command failed: %v", err)
	}

	if !strings.Contains(stdout, "no templates found") {
		t.Errorf("list should show no templates, got: %s", stdout)
	}
}

// TestE2E_FullWorkflow tests the complete push → list → apply → delete workflow.
func TestE2E_FullWorkflow(t *testing.T) {
	binary := findBinary(t)
	_, workDir, env := setupE2EEnvironment(t)

	// Create source files in work directory (using new default patterns)
	createTestFiles(t, workDir, map[string]string{
		"AGENTS.md":                       "# E2E Test Agents",
		".github/copilot-instructions.md": "# Copilot Instructions",
		".vscode/mcp.json":                `{"servers": {}}`,
	})

	templateName := "e2e-test-template"

	// Step 1: Push template
	stdout, stderr, err := runDotgh(t, binary, []string{"push", templateName}, workDir, env)
	if err != nil {
		t.Fatalf("push failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "Pushing to template") {
		t.Errorf("push output unexpected: %s", stdout)
	}

	// Step 2: List templates
	stdout, _, err = runDotgh(t, binary, []string{"list"}, workDir, env)
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if !strings.Contains(stdout, templateName) {
		t.Errorf("pushed template should appear in list, got: %s", stdout)
	}

	// Step 3: Apply to new directory
	applyDir := filepath.Join(t.TempDir(), "apply-target")
	if err := os.MkdirAll(applyDir, 0755); err != nil {
		t.Fatalf("failed to create apply dir: %v", err)
	}

	stdout, stderr, err = runDotgh(t, binary, []string{"apply", templateName}, applyDir, env)
	if err != nil {
		t.Fatalf("apply failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "Applying template") {
		t.Errorf("apply output unexpected: %s", stdout)
	}

	// Verify files were copied
	expectedFiles := []string{"AGENTS.md", ".github/copilot-instructions.md", ".vscode/mcp.json"}
	verifyFilesExist(t, applyDir, expectedFiles)
	verifyFileContent(t, filepath.Join(applyDir, "AGENTS.md"), "# E2E Test Agents")

	// Step 4: Delete template (with force flag)
	stdout, stderr, err = runDotgh(t, binary, []string{"delete", templateName, "-f"}, workDir, env)
	if err != nil {
		t.Fatalf("delete failed: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	if !strings.Contains(stdout, "deleted") {
		t.Errorf("delete output unexpected: %s", stdout)
	}

	// Step 5: Verify template is gone
	stdout, _, err = runDotgh(t, binary, []string{"list"}, workDir, env)
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if strings.Contains(stdout, templateName) {
		t.Error("deleted template should not appear in list")
	}

	// Applied files should still exist
	verifyFilesExist(t, applyDir, expectedFiles)
}

// TestE2E_ForceOverwrite tests the -f flag behavior.
func TestE2E_ForceOverwrite(t *testing.T) {
	binary := findBinary(t)
	templatesDir, workDir, env := setupE2EEnvironment(t)

	templateName := "overwrite-test"

	// Create initial content and push
	createTestFiles(t, workDir, map[string]string{
		"AGENTS.md": "# Version 1",
	})
	_, _, err := runDotgh(t, binary, []string{"push", templateName}, workDir, env)
	if err != nil {
		t.Fatalf("initial push failed: %v", err)
	}

	// Update content and push without force (should skip)
	if err := os.WriteFile(filepath.Join(workDir, "AGENTS.md"), []byte("# Version 2"), 0644); err != nil {
		t.Fatal(err)
	}
	stdout, _, err := runDotgh(t, binary, []string{"push", templateName}, workDir, env)
	if err != nil {
		t.Fatalf("second push failed: %v", err)
	}
	if !strings.Contains(stdout, "skipped") {
		t.Errorf("should skip existing without force: %s", stdout)
	}

	// Verify still version 1 in template
	verifyFileContent(t, filepath.Join(templatesDir, templateName, "AGENTS.md"), "# Version 1")

	// Push with force (should overwrite)
	_, _, err = runDotgh(t, binary, []string{"push", templateName, "-f"}, workDir, env)
	if err != nil {
		t.Fatalf("force push failed: %v", err)
	}

	// Verify now version 2
	verifyFileContent(t, filepath.Join(templatesDir, templateName, "AGENTS.md"), "# Version 2")

	// Test apply with force
	applyDir := t.TempDir()
	createTestFiles(t, applyDir, map[string]string{
		"AGENTS.md": "# Existing Content",
	})

	// Apply without force (should skip)
	stdout, _, err = runDotgh(t, binary, []string{"apply", templateName}, applyDir, env)
	if err != nil {
		t.Fatalf("apply without force failed: %v", err)
	}
	if !strings.Contains(stdout, "skipped") {
		t.Errorf("should skip existing without force: %s", stdout)
	}

	// Apply with force (should overwrite)
	_, _, err = runDotgh(t, binary, []string{"apply", templateName, "-f"}, applyDir, env)
	if err != nil {
		t.Fatalf("apply with force failed: %v", err)
	}

	verifyFileContent(t, filepath.Join(applyDir, "AGENTS.md"), "# Version 2")
}

// TestE2E_ErrorHandling tests error scenarios.
func TestE2E_ErrorHandling(t *testing.T) {
	binary := findBinary(t)
	_, workDir, env := setupE2EEnvironment(t)

	// Apply non-existent template
	stdout, stderr, err := runDotgh(t, binary, []string{"apply", "non-existent"}, workDir, env)
	if err == nil {
		t.Error("apply non-existent template should fail")
	}
	if !strings.Contains(stdout+stderr, "not found") {
		t.Errorf("error should indicate template not found: stdout=%s stderr=%s", stdout, stderr)
	}

	// Delete non-existent template
	_, _, err = runDotgh(t, binary, []string{"delete", "non-existent", "-f"}, workDir, env)
	if err == nil {
		t.Error("delete non-existent template should fail")
	}

	// Push with no target files
	emptyDir := t.TempDir()
	stdout, _, err = runDotgh(t, binary, []string{"push", "empty-template"}, emptyDir, env)
	if err != nil {
		t.Logf("push empty dir returned error (may be expected): %v", err)
	}
	// Should indicate no target files or succeed with 0 files
	if !strings.Contains(stdout, "No target files found") && !strings.Contains(stdout, "0 file") {
		t.Logf("empty push output: %s", stdout)
	}
}

// TestE2E_CrossPlatformPaths tests path handling across platforms.
func TestE2E_CrossPlatformPaths(t *testing.T) {
	binary := findBinary(t)
	_, workDir, env := setupE2EEnvironment(t)

	// Create nested directory structure (using new default patterns)
	createTestFiles(t, workDir, map[string]string{
		".github/copilot-instructions.md":           "# Copilot",
		".github/prompts/test.prompt.md":            "# Prompt",
		".github/instructions/dev.instructions.md":  "# Instructions",
		".vscode/mcp.json":                          `{"servers": {}}`,
		"AGENTS.md":                                 "# Agents",
	})

	templateName := "nested-paths"

	// Push
	_, _, err := runDotgh(t, binary, []string{"push", templateName}, workDir, env)
	if err != nil {
		t.Fatalf("push failed: %v", err)
	}

	// Apply to new directory
	applyDir := t.TempDir()
	_, _, err = runDotgh(t, binary, []string{"apply", templateName}, applyDir, env)
	if err != nil {
		t.Fatalf("apply failed: %v", err)
	}

	// Verify all nested paths work correctly
	expectedFiles := []string{
		".github/copilot-instructions.md",
		".github/prompts/test.prompt.md",
		".github/instructions/dev.instructions.md",
		".vscode/mcp.json",
		"AGENTS.md",
	}
	verifyFilesExist(t, applyDir, expectedFiles)
	verifyFileContent(t, filepath.Join(applyDir, ".github", "copilot-instructions.md"), "# Copilot")
}
