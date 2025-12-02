package glob

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// setupTestDir creates a temporary directory with test files.
func setupTestDir(t *testing.T, files []string) string {
	t.Helper()
	tempDir := t.TempDir()

	for _, file := range files {
		fullPath := filepath.Join(tempDir, file)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte("test content"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", file, err)
		}
	}

	return tempDir
}

func TestExpandPatterns(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		patterns []string
		want     []string
	}{
		{
			name:     "single file pattern",
			files:    []string{"AGENTS.md", "README.md"},
			patterns: []string{"AGENTS.md"},
			want:     []string{"AGENTS.md"},
		},
		{
			name:     "wildcard pattern in directory",
			files:    []string{".github/prompts/a.prompt.md", ".github/prompts/b.prompt.md", ".github/other.md"},
			patterns: []string{".github/prompts/*.prompt.md"},
			want:     []string{".github/prompts/a.prompt.md", ".github/prompts/b.prompt.md"},
		},
		{
			name:     "multiple patterns",
			files:    []string{"AGENTS.md", ".github/copilot-instructions.md", ".vscode/mcp.json"},
			patterns: []string{"AGENTS.md", ".github/copilot-instructions.md", ".vscode/mcp.json"},
			want:     []string{"AGENTS.md", ".github/copilot-instructions.md", ".vscode/mcp.json"},
		},
		{
			name:     "pattern with no matches",
			files:    []string{"AGENTS.md"},
			patterns: []string{"nonexistent.md"},
			want:     []string{},
		},
		{
			name:     "mixed patterns with and without matches",
			files:    []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			patterns: []string{"AGENTS.md", ".github/instructions/*.instructions.md", ".github/prompts/*.prompt.md"},
			want:     []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
		},
		{
			name:     "instructions pattern",
			files:    []string{".github/instructions/go.instructions.md", ".github/instructions/test.instructions.md"},
			patterns: []string{".github/instructions/*.instructions.md"},
			want:     []string{".github/instructions/go.instructions.md", ".github/instructions/test.instructions.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := setupTestDir(t, tt.files)

			got, err := ExpandPatterns(baseDir, tt.patterns)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Sort both for comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("ExpandPatterns() returned %d files, want %d\ngot: %v\nwant: %v", len(got), len(tt.want), got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ExpandPatterns()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestExpandPatternsInvalidPattern(t *testing.T) {
	tempDir := t.TempDir()

	// Invalid glob pattern (unmatched bracket)
	_, err := ExpandPatterns(tempDir, []string{"[invalid"})
	if err == nil {
		t.Error("expected error for invalid pattern, got nil")
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "AGENTS.md",
			path:    "AGENTS.md",
			want:    true,
		},
		{
			name:    "no match",
			pattern: "AGENTS.md",
			path:    "README.md",
			want:    false,
		},
		{
			name:    "wildcard match",
			pattern: "*.md",
			path:    "AGENTS.md",
			want:    true,
		},
		{
			name:    "directory pattern match",
			pattern: ".github/prompts/*.prompt.md",
			path:    ".github/prompts/test.prompt.md",
			want:    true,
		},
		{
			name:    "directory pattern no match",
			pattern: ".github/prompts/*.prompt.md",
			path:    ".github/prompts/test.md",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchPattern(tt.pattern, tt.path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("MatchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestFilterExcludes(t *testing.T) {
	tests := []struct {
		name            string
		files           []string
		excludePatterns []string
		want            []string
		wantErr         bool
	}{
		{
			name:            "no excludes returns all files",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			excludePatterns: nil,
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "empty excludes returns all files",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			excludePatterns: []string{},
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "exact match exclude",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md", ".github/prompts/local.prompt.md"},
			excludePatterns: []string{".github/prompts/local.prompt.md"},
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "wildcard exclude",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md", ".github/prompts/secret-key.prompt.md", ".github/prompts/secret-token.prompt.md"},
			excludePatterns: []string{".github/prompts/secret-*.prompt.md"},
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "multiple exclude patterns",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md", ".github/prompts/local.prompt.md", ".vscode/mcp.json"},
			excludePatterns: []string{".github/prompts/local.prompt.md", ".vscode/mcp.json"},
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "exclude all files",
			files:           []string{"AGENTS.md"},
			excludePatterns: []string{"AGENTS.md"},
			want:            []string{},
			wantErr:         false,
		},
		{
			name:            "exclude pattern matches nothing",
			files:           []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			excludePatterns: []string{"nonexistent.md"},
			want:            []string{"AGENTS.md", ".github/prompts/test.prompt.md"},
			wantErr:         false,
		},
		{
			name:            "empty files list",
			files:           []string{},
			excludePatterns: []string{"AGENTS.md"},
			want:            []string{},
			wantErr:         false,
		},
		{
			name:            "invalid exclude pattern",
			files:           []string{"AGENTS.md"},
			excludePatterns: []string{"[invalid"},
			want:            nil,
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FilterExcludes(tt.files, tt.excludePatterns)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Sort both for comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("FilterExcludes() returned %d files, want %d\ngot: %v\nwant: %v", len(got), len(tt.want), got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("FilterExcludes()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestFilterExcludesPreservesOrder(t *testing.T) {
	// Test that FilterExcludes preserves the order of non-excluded files
	files := []string{"c.md", "b.md", "a.md"}
	excludePatterns := []string{}

	got, err := FilterExcludes(files, excludePatterns)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Order should be preserved
	want := []string{"c.md", "b.md", "a.md"}
	if len(got) != len(want) {
		t.Fatalf("FilterExcludes() returned %d files, want %d", len(got), len(want))
	}

	for i := range got {
		if got[i] != want[i] {
			t.Errorf("FilterExcludes()[%d] = %q, want %q (order should be preserved)", i, got[i], want[i])
		}
	}
}
