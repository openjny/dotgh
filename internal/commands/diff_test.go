package commands

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openjny/dotgh/internal/config"
)

// executeDiffCmd runs the diff command and returns the output.
func executeDiffCmd(t *testing.T, templatesDir, targetDir, templateName string, reverse, merge bool, excludes []string) (string, error) {
	t.Helper()
	var cfg *config.Config
	if excludes == nil {
		cfg = testConfig()
	} else {
		cfg = testConfigWithExcludes(excludes)
	}
	cmd := NewDiffCmdWithConfig(templatesDir, targetDir, cfg)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	args := []string{templateName}
	if reverse {
		args = append(args, "--reverse")
	}
	if merge {
		args = append(args, "--merge")
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestDiffNoChanges(t *testing.T) {
	// Setup identical template and target
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Same Content",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md": "# Same Content",
	})

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// No changes means no error (exit code 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "(no changes)") {
		t.Errorf("output should indicate no changes, got:\n%s", output)
	}
}

func TestDiffWithAdditions(t *testing.T) {
	// Setup template with files not in target
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/copilot-instructions.md": "# Instructions",
	})
	targetDir := t.TempDir()

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	if !strings.Contains(output, "+ AGENTS.md") {
		t.Errorf("output should show addition of AGENTS.md, got:\n%s", output)
	}
	if !strings.Contains(output, "+ .github/copilot-instructions.md") {
		t.Errorf("output should show addition of copilot-instructions.md, got:\n%s", output)
	}
	if !strings.Contains(output, "2 addition(s)") {
		t.Errorf("output should show 2 additions, got:\n%s", output)
	}
}

func TestDiffWithModifications(t *testing.T) {
	// Setup template and target with different content
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# New Content",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md": "# Old Content",
	})

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification of AGENTS.md, got:\n%s", output)
	}
	if !strings.Contains(output, "1 modification(s)") {
		t.Errorf("output should show 1 modification, got:\n%s", output)
	}
}

func TestDiffWithDeletions(t *testing.T) {
	// Setup template without files that exist in target
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/copilot-instructions.md": "# Will be deleted",
	})

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	if !strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("output should show deletion of copilot-instructions.md, got:\n%s", output)
	}
	if !strings.Contains(output, "1 deletion(s)") {
		t.Errorf("output should show 1 deletion, got:\n%s", output)
	}
}

func TestDiffMergeMode(t *testing.T) {
	// Setup template and target with deletable files
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		".github/copilot-instructions.md": "# Should be kept in merge mode",
	})

	// Without merge mode - should show deletion (and return ErrDiffFound)
	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}
	if !strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("without merge mode, should show deletion, got:\n%s", output)
	}

	// With merge mode - should NOT show deletion (but still has addition)
	output, err = executeDiffCmd(t, templatesDir, targetDir, "my-template", false, true, nil)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}
	if strings.Contains(output, "- .github/copilot-instructions.md") {
		t.Errorf("with merge mode, should NOT show deletion, got:\n%s", output)
	}
	if !strings.Contains(output, "merge mode") {
		t.Errorf("output should indicate merge mode, got:\n%s", output)
	}
}

func TestDiffReverse(t *testing.T) {
	// Setup template with files
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Template Agents",
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md":                       "# Local Agents",
		".github/copilot-instructions.md": "# Local only",
	})

	// Reverse mode: current -> template
	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", true, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	if !strings.Contains(output, "current directory → template") {
		t.Errorf("output should indicate reverse direction, got:\n%s", output)
	}
	// In reverse mode: local-only file should be added to template
	if !strings.Contains(output, "+ .github/copilot-instructions.md") {
		t.Errorf("output should show addition in reverse mode, got:\n%s", output)
	}
	// AGENTS.md has different content, should be modified
	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification in reverse mode, got:\n%s", output)
	}
}

func TestDiffTemplateNotFound(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{})
	targetDir := t.TempDir()

	output, err := executeDiffCmd(t, templatesDir, targetDir, "non-existent", false, false, nil)

	if err == nil {
		t.Error("expected error for non-existent template")
	}

	if !strings.Contains(output, "not found") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention template not found, got:\n%s\nerror: %v", output, err)
	}
}

func TestDiffWithExcludes(t *testing.T) {
	// Setup template with files
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/prompts/test.prompt.md":  "# Test",
		".github/prompts/local.prompt.md": "# Local",
	})
	targetDir := t.TempDir()

	// Exclude local.prompt.md
	excludes := []string{".github/prompts/local.prompt.md"}
	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, excludes)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	// Should show additions for non-excluded files
	if !strings.Contains(output, "+ AGENTS.md") {
		t.Errorf("output should show AGENTS.md addition, got:\n%s", output)
	}
	if !strings.Contains(output, "+ .github/prompts/test.prompt.md") {
		t.Errorf("output should show test.prompt.md addition, got:\n%s", output)
	}
	// Should NOT show excluded file
	if strings.Contains(output, "local.prompt.md") {
		t.Errorf("output should NOT show excluded file, got:\n%s", output)
	}
	// Summary should show 2 additions
	if !strings.Contains(output, "2 addition(s)") {
		t.Errorf("output should show 2 additions, got:\n%s", output)
	}
}

func TestDiffMixedChanges(t *testing.T) {
	// Setup with mixed changes
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md":                       "# New Agents",   // Modified
		".github/copilot-instructions.md": "# Instructions", // Added
		".github/prompts/test.prompt.md":  "# Test",         // Unchanged
	})
	targetDir := t.TempDir()
	createTestFiles(t, targetDir, map[string]string{
		"AGENTS.md":                      "# Old Agents", // Different content
		".github/prompts/test.prompt.md": "# Test",       // Same content
		".vscode/mcp.json":               "{}",           // Will be deleted
	})

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	// Check all types of changes
	if !strings.Contains(output, "+ .github/copilot-instructions.md") {
		t.Errorf("output should show addition, got:\n%s", output)
	}
	if !strings.Contains(output, "M AGENTS.md") {
		t.Errorf("output should show modification, got:\n%s", output)
	}
	if !strings.Contains(output, "- .vscode/mcp.json") {
		t.Errorf("output should show deletion, got:\n%s", output)
	}

	// Check summary
	if !strings.Contains(output, "1 addition(s), 1 modification(s), 1 deletion(s)") {
		t.Errorf("output should show correct summary, got:\n%s", output)
	}
}

func TestDiffRequiresTemplateName(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{"my-template"})
	targetDir := t.TempDir()

	cmd := NewDiffCmd(templatesDir, targetDir)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{}) // No template name

	err := cmd.Execute()

	if err == nil {
		t.Error("expected error when no template name provided")
	}
}

func TestDiffNonExistentTargetDir(t *testing.T) {
	// Template exists but target dir doesn't - should still work
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# Agents",
	})
	targetDir := filepath.Join(t.TempDir(), "non-existent")

	output, err := executeDiffCmd(t, templatesDir, targetDir, "my-template", false, false, nil)

	// Differences found should return ErrDiffFound (exit code 1)
	if !errors.Is(err, ErrDiffFound) {
		t.Fatalf("expected ErrDiffFound, got: %v", err)
	}

	// All template files should be additions
	if !strings.Contains(output, "+ AGENTS.md") {
		t.Errorf("output should show addition, got:\n%s", output)
	}
}
