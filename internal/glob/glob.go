// Package glob provides glob pattern matching utilities for dotgh.
package glob

import (
	"fmt"
	"path/filepath"
)

// ExpandPatterns expands glob patterns and returns matched file paths relative to baseDir.
// Patterns that don't match any files are silently ignored.
// Returned paths always use forward slashes for cross-platform consistency.
func ExpandPatterns(baseDir string, patterns []string) ([]string, error) {
	var result []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		fullPattern := filepath.Join(baseDir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
		}

		for _, match := range matches {
			// Convert to relative path
			relPath, err := filepath.Rel(baseDir, match)
			if err != nil {
				return nil, fmt.Errorf("get relative path: %w", err)
			}

			// Normalize to forward slashes for cross-platform consistency
			relPath = filepath.ToSlash(relPath)

			// Deduplicate
			if !seen[relPath] {
				seen[relPath] = true
				result = append(result, relPath)
			}
		}
	}

	return result, nil
}

// MatchPattern checks if a path matches a glob pattern.
func MatchPattern(pattern, path string) (bool, error) {
	return filepath.Match(pattern, path)
}

// FilterExcludes filters out files that match any of the exclude patterns.
// Files is expected to be a list of relative paths with forward slashes.
// The order of non-excluded files is preserved.
// Returns nil and an error if any exclude pattern is invalid.
func FilterExcludes(files []string, excludePatterns []string) ([]string, error) {
	if len(excludePatterns) == 0 {
		return files, nil
	}

	var result []string
	for _, file := range files {
		excluded := false
		for _, pattern := range excludePatterns {
			matched, err := filepath.Match(pattern, file)
			if err != nil {
				return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
			}
			if matched {
				excluded = true
				break
			}
		}
		if !excluded {
			result = append(result, file)
		}
	}

	return result, nil
}
