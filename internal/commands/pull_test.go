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
// If excludes is nil, the default config is used.
func executePullCmd(t *testing.T, templatesDir, targetDir, templateName string, force bool, excludes []string) (string, error) {
	t.Helper()
	var cfg *config.Config
	if excludes == nil {
		cfg = testConfig()
	} else {
		cfg = testConfigWithExcludes(excludes)
	}
	cmd := NewPullCmdWithConfig(templatesDir, targetDir, cfg)
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

func TestPullTemplate(t *testing.T) {
	tests := []struct {
		name            string
		templateName    string
		templateFiles   map[string]string
		existingFiles   map[string]string // files already in target dir
		force           bool
		wantContains    []string
		wantNotContains []string
		wantFiles       []string // files that should exist after pull
		wantErr         bool
	}{
		{
			name:         "pull template with AGENTS.md",
			templateName: "my-template",
			templateFiles: map[string]string{
				"AGENTS.md": "# My Agents",
			},
			existingFiles: nil,
			force:         false,
			wantContains:  []string{"Pulling template", "my-template", "AGENTS.md", "copied"},
			wantFiles:     []string{"AGENTS.md"},
			wantErr:       false,
		},
		{
			name:         "pull template with .github copilot-instructions",
			templateName: "github-template",
			templateFiles: map[string]string{
				".github/copilot-instructions.md": "# Instructions",
			},
			existingFiles: nil,
			force:         false,
			wantContains:  []string{"Pulling template", "copilot-instructions.md", "copied"},
			wantFiles:     []string{".github/copilot-instructions.md"},
			wantErr:       false,
		},
		{
			name:         "pull template with glob pattern files",
			templateName: "glob-template",
			templateFiles: map[string]string{
				".github/prompts/test.prompt.md":          "# Test prompt",
				".github/instructions/go.instructions.md": "# Go instructions",
			},
			existingFiles: nil,
			force:         false,
			wantContains:  []string{"Pulling template", "prompt.md", "instructions.md", "copied"},
			wantFiles:     []string{".github/prompts/test.prompt.md", ".github/instructions/go.instructions.md"},
			wantErr:       false,
		},
		{
			name:         "skip existing file without force",
			templateName: "skip-template",
			templateFiles: map[string]string{
				"AGENTS.md": "# New Content",
			},
			existingFiles: map[string]string{
				"AGENTS.md": "# Existing Content",
			},
			force:        false,
			wantContains: []string{"skipped"},
			wantFiles:    []string{"AGENTS.md"},
			wantErr:      false,
		},
		{
			name:         "overwrite existing file with force",
			templateName: "force-template",
			templateFiles: map[string]string{
				"AGENTS.md": "# New Content",
			},
			existingFiles: map[string]string{
				"AGENTS.md": "# Existing Content",
			},
			force:        true,
			wantContains: []string{"copied"},
			wantFiles:    []string{"AGENTS.md"},
			wantErr:      false,
		},
		{
			name:         "template with multiple targets",
			templateName: "full-template",
			templateFiles: map[string]string{
				".github/copilot-instructions.md": "# Copilot",
				".vscode/mcp.json":                "{}",
				"AGENTS.md":                       "# Agents",
			},
			existingFiles: nil,
			force:         false,
			wantContains:  []string{"copilot-instructions.md", "mcp.json", "AGENTS.md"},
			wantFiles:     []string{".github/copilot-instructions.md", ".vscode/mcp.json", "AGENTS.md"},
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup template
			templatesDir := setupTestTemplateWithFiles(t, tt.templateName, tt.templateFiles)

			// Setup target directory
			targetDir := t.TempDir()
			for path, content := range tt.existingFiles {
				fullPath := filepath.Join(targetDir, path)
				dir := filepath.Dir(fullPath)
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatalf("failed to create directory: %v", err)
				}
				if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
					t.Fatalf("failed to create existing file: %v", err)
				}
			}

			// Execute
			output, err := executePullCmd(t, templatesDir, targetDir, tt.templateName, tt.force, nil)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("pull error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check output contains expected strings
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}

			// Check output does not contain unexpected strings
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("output should NOT contain %q, got:\n%s", notWant, output)
				}
			}

			// Check files exist
			for _, file := range tt.wantFiles {
				fullPath := filepath.Join(targetDir, file)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("expected file %s to exist", file)
				}
			}
		})
	}
}

func TestPullTemplateNotFound(t *testing.T) {
	templatesDir := setupTestTemplatesDir(t, []string{})
	targetDir := t.TempDir()

	output, err := executePullCmd(t, templatesDir, targetDir, "non-existent", false, nil)

	if err == nil {
		t.Error("expected error for non-existent template")
	}

	if !strings.Contains(output, "not found") {
		t.Errorf("output should indicate template not found, got:\n%s", output)
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

func TestPullPreservesExistingContent(t *testing.T) {
	// When skip happens, existing file content should be preserved
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": "# New Content",
	})

	targetDir := t.TempDir()
	existingContent := "# Existing Content - Should Stay"
	if err := os.WriteFile(filepath.Join(targetDir, "AGENTS.md"), []byte(existingContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := executePullCmd(t, templatesDir, targetDir, "my-template", false, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check content is preserved
	content, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != existingContent {
		t.Errorf("existing content should be preserved, got: %s", string(content))
	}
}

func TestPullOverwritesWithForce(t *testing.T) {
	newContent := "# New Content"
	templatesDir := setupTestTemplateWithFiles(t, "my-template", map[string]string{
		"AGENTS.md": newContent,
	})

	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(targetDir, "AGENTS.md"), []byte("# Old"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := executePullCmd(t, templatesDir, targetDir, "my-template", true, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check content is overwritten
	content, err := os.ReadFile(filepath.Join(targetDir, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}

	if string(content) != newContent {
		t.Errorf("content should be overwritten, got: %s", string(content))
	}
}

func TestPullWithExcludes(t *testing.T) {
	tests := []struct {
		name          string
		templateFiles map[string]string
		excludes      []string
		wantFiles     []string // files that should exist after pull
		wantNotFiles  []string // files that should NOT exist after pull
		wantContains  []string
	}{
		{
			name: "exclude specific file",
			templateFiles: map[string]string{
				"AGENTS.md":                       "# Agents",
				".github/prompts/test.prompt.md":  "# Test",
				".github/prompts/local.prompt.md": "# Local",
			},
			excludes:     []string{".github/prompts/local.prompt.md"},
			wantFiles:    []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantNotFiles: []string{".github/prompts/local.prompt.md"},
			wantContains: []string{"Pulling template"},
		},
		{
			name: "exclude with wildcard pattern",
			templateFiles: map[string]string{
				"AGENTS.md":                              "# Agents",
				".github/prompts/test.prompt.md":         "# Test",
				".github/prompts/secret-key.prompt.md":   "# Secret Key",
				".github/prompts/secret-token.prompt.md": "# Secret Token",
			},
			excludes:     []string{".github/prompts/secret-*.prompt.md"},
			wantFiles:    []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantNotFiles: []string{".github/prompts/secret-key.prompt.md", ".github/prompts/secret-token.prompt.md"},
			wantContains: []string{"Pulling template"},
		},
		{
			name: "exclude multiple patterns",
			templateFiles: map[string]string{
				"AGENTS.md":                       "# Agents",
				".github/prompts/test.prompt.md":  "# Test",
				".github/prompts/local.prompt.md": "# Local",
				".vscode/mcp.json":                "{}",
			},
			excludes:     []string{".github/prompts/local.prompt.md", ".vscode/mcp.json"},
			wantFiles:    []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantNotFiles: []string{".github/prompts/local.prompt.md", ".vscode/mcp.json"},
			wantContains: []string{"Pulling template"},
		},
		{
			name: "empty excludes copies all files",
			templateFiles: map[string]string{
				"AGENTS.md":                      "# Agents",
				".github/prompts/test.prompt.md": "# Test",
			},
			excludes:     []string{},
			wantFiles:    []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantNotFiles: []string{},
			wantContains: []string{"Pulling template"},
		},
		{
			name: "exclude all matching files",
			templateFiles: map[string]string{
				"AGENTS.md": "# Agents",
			},
			excludes:     []string{"AGENTS.md"},
			wantFiles:    []string{},
			wantNotFiles: []string{"AGENTS.md"},
			wantContains: []string{"no matching files found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup template
			templatesDir := setupTestTemplateWithFiles(t, "exclude-test", tt.templateFiles)
			targetDir := t.TempDir()

			// Execute
			output, err := executePullCmd(t, templatesDir, targetDir, "exclude-test", false, tt.excludes)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check output contains expected strings
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}

			// Check expected files exist
			for _, file := range tt.wantFiles {
				fullPath := filepath.Join(targetDir, file)
				if _, err := os.Stat(fullPath); os.IsNotExist(err) {
					t.Errorf("expected file %s to exist", file)
				}
			}

			// Check excluded files do NOT exist
			for _, file := range tt.wantNotFiles {
				fullPath := filepath.Join(targetDir, file)
				if _, err := os.Stat(fullPath); !os.IsNotExist(err) {
					t.Errorf("file %s should NOT exist (should be excluded)", file)
				}
			}
		})
	}
}
