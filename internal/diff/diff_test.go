package diff

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openjny/dotgh/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestFile creates a file with the given content in the base directory.
func createTestFile(t *testing.T, baseDir, relativePath, content string) {
	t.Helper()
	fullPath := filepath.Join(baseDir, relativePath)
	dir := filepath.Dir(fullPath)
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
}

// createTestFiles creates multiple files in the given directory.
func createTestFiles(t *testing.T, baseDir string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		createTestFile(t, baseDir, path, content)
	}
}

func TestComputeDiff_AddedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create files in source only
	createTestFiles(t, srcDir, map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/copilot-instructions.md": "# Instructions",
	})

	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Len(t, diff.Added, 2)
	assert.Empty(t, diff.Modified)
	assert.Empty(t, diff.Deleted)
	assert.Empty(t, diff.Unchanged)
	assert.True(t, diff.HasChanges())
	assert.Equal(t, 2, diff.TotalChanges())
}

func TestComputeDiff_ModifiedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create same file with different content
	createTestFile(t, srcDir, "AGENTS.md", "# New Content")
	createTestFile(t, dstDir, "AGENTS.md", "# Old Content")

	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Empty(t, diff.Added)
	assert.Len(t, diff.Modified, 1)
	assert.Equal(t, "AGENTS.md", diff.Modified[0].Path)
	assert.Empty(t, diff.Deleted)
	assert.Empty(t, diff.Unchanged)
}

func TestComputeDiff_DeletedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create file only in destination
	createTestFile(t, dstDir, "AGENTS.md", "# Old Content")

	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Empty(t, diff.Added)
	assert.Empty(t, diff.Modified)
	assert.Len(t, diff.Deleted, 1)
	assert.Equal(t, "AGENTS.md", diff.Deleted[0].Path)
	assert.Empty(t, diff.Unchanged)
}

func TestComputeDiff_UnchangedFiles(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Create same file with same content
	createTestFile(t, srcDir, "AGENTS.md", "# Same Content")
	createTestFile(t, dstDir, "AGENTS.md", "# Same Content")

	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Empty(t, diff.Added)
	assert.Empty(t, diff.Modified)
	assert.Empty(t, diff.Deleted)
	assert.Len(t, diff.Unchanged, 1)
	assert.False(t, diff.HasChanges())
}

func TestComputeDiff_MergeMode(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Source has one file, destination has another
	createTestFile(t, srcDir, "AGENTS.md", "# Source")
	createTestFile(t, dstDir, ".github/copilot-instructions.md", "# Dest only")

	// Full sync mode - should include deletion
	diffFullSync, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)
	assert.Len(t, diffFullSync.Added, 1)
	assert.Len(t, diffFullSync.Deleted, 1)

	// Merge mode - should NOT include deletion
	diffMerge, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, true)
	require.NoError(t, err)
	assert.Len(t, diffMerge.Added, 1)
	assert.Empty(t, diffMerge.Deleted)
}

func TestComputeDiff_WithExcludes(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	createTestFiles(t, srcDir, map[string]string{
		"AGENTS.md":                       "# Agents",
		".github/prompts/test.prompt.md":  "# Test",
		".github/prompts/local.prompt.md": "# Local",
	})

	excludes := []string{".github/prompts/local.prompt.md"}
	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, excludes, false)
	require.NoError(t, err)

	// local.prompt.md should be excluded
	assert.Len(t, diff.Added, 2)
	paths := []string{diff.Added[0].Path, diff.Added[1].Path}
	assert.Contains(t, paths, "AGENTS.md")
	assert.Contains(t, paths, ".github/prompts/test.prompt.md")
}

func TestComputeDiff_MixedChanges(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Source files
	createTestFiles(t, srcDir, map[string]string{
		"AGENTS.md":                       "# New Agents",   // Will be modified
		".github/copilot-instructions.md": "# Instructions", // Will be added
		".github/prompts/test.prompt.md":  "# Test",         // Unchanged
	})

	// Destination files
	createTestFiles(t, dstDir, map[string]string{
		"AGENTS.md":                      "# Old Agents", // Different content
		".github/prompts/test.prompt.md": "# Test",       // Same content
		".vscode/mcp.json":               "{}",           // Will be deleted
	})

	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Len(t, diff.Added, 1)
	assert.Equal(t, ".github/copilot-instructions.md", diff.Added[0].Path)

	assert.Len(t, diff.Modified, 1)
	assert.Equal(t, "AGENTS.md", diff.Modified[0].Path)

	assert.Len(t, diff.Deleted, 1)
	assert.Equal(t, ".vscode/mcp.json", diff.Deleted[0].Path)

	assert.Len(t, diff.Unchanged, 1)
	assert.Equal(t, ".github/prompts/test.prompt.md", diff.Unchanged[0].Path)

	assert.Equal(t, 3, diff.TotalChanges())
}

func TestComputeDiff_NonExistentSourceDir(t *testing.T) {
	srcDir := filepath.Join(t.TempDir(), "non-existent")
	dstDir := t.TempDir()

	createTestFile(t, dstDir, "AGENTS.md", "# Content")

	// Should not error, just return deletions for files in dest
	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Empty(t, diff.Added)
	assert.Empty(t, diff.Modified)
	assert.Len(t, diff.Deleted, 1)
}

func TestComputeDiff_NonExistentDestDir(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := filepath.Join(t.TempDir(), "non-existent")

	createTestFile(t, srcDir, "AGENTS.md", "# Content")

	// Should not error, just return additions for files in src
	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	assert.Len(t, diff.Added, 1)
	assert.Empty(t, diff.Modified)
	assert.Empty(t, diff.Deleted)
}

func TestApplyChanges(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Setup source files
	createTestFiles(t, srcDir, map[string]string{
		"AGENTS.md":                       "# New Agents",
		".github/copilot-instructions.md": "# Instructions",
	})

	// Setup destination files (one to be modified, one to be deleted)
	createTestFiles(t, dstDir, map[string]string{
		"AGENTS.md":        "# Old Agents",
		".vscode/mcp.json": "{}",
	})

	// Compute diff
	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, false)
	require.NoError(t, err)

	// Apply changes
	err = ApplyChanges(srcDir, dstDir, diff)
	require.NoError(t, err)

	// Verify added file
	content, err := os.ReadFile(filepath.Join(dstDir, ".github/copilot-instructions.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Instructions", string(content))

	// Verify modified file
	content, err = os.ReadFile(filepath.Join(dstDir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "# New Agents", string(content))

	// Verify deleted file
	_, err = os.Stat(filepath.Join(dstDir, ".vscode/mcp.json"))
	assert.True(t, os.IsNotExist(err))
}

func TestApplyChanges_MergeMode(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	// Setup source files
	createTestFile(t, srcDir, "AGENTS.md", "# New Content")

	// Setup destination files
	createTestFiles(t, dstDir, map[string]string{
		".github/copilot-instructions.md": "# Keep this",
	})

	// Compute diff in merge mode (no deletions)
	diff, err := ComputeDiff(srcDir, dstDir, config.DefaultIncludes, nil, true)
	require.NoError(t, err)

	// Apply changes
	err = ApplyChanges(srcDir, dstDir, diff)
	require.NoError(t, err)

	// Verify added file
	content, err := os.ReadFile(filepath.Join(dstDir, "AGENTS.md"))
	require.NoError(t, err)
	assert.Equal(t, "# New Content", string(content))

	// Verify file that would be deleted in full sync is preserved
	content, err = os.ReadFile(filepath.Join(dstDir, ".github/copilot-instructions.md"))
	require.NoError(t, err)
	assert.Equal(t, "# Keep this", string(content))
}

func TestDiffResult_AllChanges(t *testing.T) {
	diff := &DiffResult{
		Added:    []FileChange{{Path: "a.md", ChangeType: ChangeAdd}},
		Modified: []FileChange{{Path: "b.md", ChangeType: ChangeModify}},
		Deleted:  []FileChange{{Path: "c.md", ChangeType: ChangeDelete}},
	}

	changes := diff.AllChanges()
	assert.Len(t, changes, 3)
}

func TestFilesAreEqual(t *testing.T) {
	dir := t.TempDir()

	// Same content
	createTestFile(t, dir, "file1.txt", "content")
	createTestFile(t, dir, "file2.txt", "content")

	equal, err := filesAreEqual(filepath.Join(dir, "file1.txt"), filepath.Join(dir, "file2.txt"))
	require.NoError(t, err)
	assert.True(t, equal)

	// Different content
	createTestFile(t, dir, "file3.txt", "different")

	equal, err = filesAreEqual(filepath.Join(dir, "file1.txt"), filepath.Join(dir, "file3.txt"))
	require.NoError(t, err)
	assert.False(t, equal)

	// Different size
	createTestFile(t, dir, "file4.txt", "longer content here")

	equal, err = filesAreEqual(filepath.Join(dir, "file1.txt"), filepath.Join(dir, "file4.txt"))
	require.NoError(t, err)
	assert.False(t, equal)
}
