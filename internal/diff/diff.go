// Package diff provides file difference calculation utilities for dotgh.
package diff

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/openjny/dotgh/internal/glob"
)

// ChangeType represents the type of file change.
type ChangeType string

const (
	// ChangeAdd indicates a file that exists in source but not in target.
	ChangeAdd ChangeType = "add"
	// ChangeModify indicates a file that exists in both and has different content.
	ChangeModify ChangeType = "modify"
	// ChangeDelete indicates a file that exists in target but not in source.
	ChangeDelete ChangeType = "delete"
	// ChangeUnchanged indicates a file that exists in both with same content.
	ChangeUnchanged ChangeType = "unchanged"
)

// FileChange represents a single file change.
type FileChange struct {
	Path       string     // Relative path of the file
	ChangeType ChangeType // Type of change
}

// DiffResult contains the result of a diff operation.
type DiffResult struct {
	Added     []FileChange // Files to add
	Modified  []FileChange // Files to modify
	Deleted   []FileChange // Files to delete
	Unchanged []FileChange // Files that are unchanged
}

// HasChanges returns true if there are any changes (add, modify, or delete).
func (r *DiffResult) HasChanges() bool {
	return len(r.Added) > 0 || len(r.Modified) > 0 || len(r.Deleted) > 0
}

// TotalChanges returns the total number of changes (add + modify + delete).
func (r *DiffResult) TotalChanges() int {
	return len(r.Added) + len(r.Modified) + len(r.Deleted)
}

// AllChanges returns all changes that will be applied (add + modify + delete).
func (r *DiffResult) AllChanges() []FileChange {
	result := make([]FileChange, 0, r.TotalChanges())
	result = append(result, r.Added...)
	result = append(result, r.Modified...)
	result = append(result, r.Deleted...)
	return result
}

// ComputeDiff calculates the difference between source and target directories.
// If mergeMode is true, deletions are not computed (files only in target are ignored).
// If mergeMode is false, it computes full sync (including deletions).
func ComputeDiff(srcDir, dstDir string, includes, excludes []string, mergeMode bool) (*DiffResult, error) {
	result := &DiffResult{
		Added:     []FileChange{},
		Modified:  []FileChange{},
		Deleted:   []FileChange{},
		Unchanged: []FileChange{},
	}

	// Get files from source directory
	srcFiles, err := getFilteredFiles(srcDir, includes, excludes)
	if err != nil {
		return nil, fmt.Errorf("get source files: %w", err)
	}

	// Get files from destination directory
	dstFiles, err := getFilteredFiles(dstDir, includes, excludes)
	if err != nil {
		return nil, fmt.Errorf("get destination files: %w", err)
	}

	// Create maps for quick lookup
	srcFileSet := make(map[string]bool)
	for _, f := range srcFiles {
		srcFileSet[f] = true
	}

	dstFileSet := make(map[string]bool)
	for _, f := range dstFiles {
		dstFileSet[f] = true
	}

	// Process source files
	for _, file := range srcFiles {
		if !dstFileSet[file] {
			// File exists only in source -> add
			result.Added = append(result.Added, FileChange{Path: file, ChangeType: ChangeAdd})
		} else {
			// File exists in both -> check if modified
			srcPath := filepath.Join(srcDir, file)
			dstPath := filepath.Join(dstDir, file)

			same, err := filesAreEqual(srcPath, dstPath)
			if err != nil {
				return nil, fmt.Errorf("compare files %s: %w", file, err)
			}

			if same {
				result.Unchanged = append(result.Unchanged, FileChange{Path: file, ChangeType: ChangeUnchanged})
			} else {
				result.Modified = append(result.Modified, FileChange{Path: file, ChangeType: ChangeModify})
			}
		}
	}

	// Process destination files (for deletions) - only in full sync mode
	if !mergeMode {
		for _, file := range dstFiles {
			if !srcFileSet[file] {
				// File exists only in destination -> delete
				result.Deleted = append(result.Deleted, FileChange{Path: file, ChangeType: ChangeDelete})
			}
		}
	}

	// Sort all slices for consistent output
	sortChanges(result.Added)
	sortChanges(result.Modified)
	sortChanges(result.Deleted)
	sortChanges(result.Unchanged)

	return result, nil
}

// getFilteredFiles returns files in the directory matching includes and not matching excludes.
func getFilteredFiles(dir string, includes, excludes []string) ([]string, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []string{}, nil
	}

	// Expand glob patterns
	files, err := glob.ExpandPatterns(dir, includes)
	if err != nil {
		return nil, fmt.Errorf("expand patterns: %w", err)
	}

	// Filter excludes
	files, err = glob.FilterExcludes(files, excludes)
	if err != nil {
		return nil, fmt.Errorf("filter excludes: %w", err)
	}

	return files, nil
}

// filesAreEqual compares two files and returns true if they have the same content.
func filesAreEqual(path1, path2 string) (bool, error) {
	// Compare file sizes first (quick check)
	info1, err := os.Stat(path1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(path2)
	if err != nil {
		return false, err
	}

	if info1.Size() != info2.Size() {
		return false, nil
	}

	// Read and compare content
	content1, err := os.ReadFile(path1)
	if err != nil {
		return false, err
	}
	content2, err := os.ReadFile(path2)
	if err != nil {
		return false, err
	}

	return bytes.Equal(content1, content2), nil
}

// sortChanges sorts a slice of FileChange by path.
func sortChanges(changes []FileChange) {
	sort.Slice(changes, func(i, j int) bool {
		return changes[i].Path < changes[j].Path
	})
}

// ApplyChanges applies the diff changes from source to destination directory.
// It copies added and modified files, and deletes files marked for deletion.
func ApplyChanges(srcDir, dstDir string, diff *DiffResult) error {
	// Apply additions and modifications
	for _, change := range diff.Added {
		if err := copyFileSync(filepath.Join(srcDir, change.Path), filepath.Join(dstDir, change.Path)); err != nil {
			return fmt.Errorf("add %s: %w", change.Path, err)
		}
	}

	for _, change := range diff.Modified {
		if err := copyFileSync(filepath.Join(srcDir, change.Path), filepath.Join(dstDir, change.Path)); err != nil {
			return fmt.Errorf("modify %s: %w", change.Path, err)
		}
	}

	// Apply deletions
	for _, change := range diff.Deleted {
		dstPath := filepath.Join(dstDir, change.Path)
		if err := os.Remove(dstPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("delete %s: %w", change.Path, err)
		}
	}

	return nil
}

// copyFileSync copies a file from src to dst, preserving permissions.
func copyFileSync(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy content: %w", err)
	}

	return nil
}
