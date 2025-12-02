package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openjny/dotgh/internal/config"
)

// testConfig returns a config with the default test includes.
// This is used across test files to avoid duplication.
func testConfig() *config.Config {
	return &config.Config{
		Includes: config.DefaultIncludes,
	}
}

// testConfigWithExcludes returns a config with the default includes and specified excludes.
func testConfigWithExcludes(excludes []string) *config.Config {
	return &config.Config{
		Includes: config.DefaultIncludes,
		Excludes: excludes,
	}
}

// createTestFile creates a file at the given path with the specified content.
// It creates parent directories as needed.
func createTestFile(t *testing.T, basePath, relativePath, content string) {
	t.Helper()
	fullPath := filepath.Join(basePath, relativePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dir, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file %s: %v", relativePath, err)
	}
}

// createTestFiles creates multiple files in the given directory.
// files is a map of relative path to content.
func createTestFiles(t *testing.T, basePath string, files map[string]string) {
	t.Helper()
	for path, content := range files {
		createTestFile(t, basePath, path, content)
	}
}
