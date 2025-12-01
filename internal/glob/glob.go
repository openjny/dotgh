// Package glob provides glob pattern matching utilities for dotgh.
package glob

import (
	"fmt"
	"path/filepath"
)

// ExpandPatterns expands glob patterns and returns matched file paths relative to baseDir.
// Patterns that don't match any files are silently ignored.
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
