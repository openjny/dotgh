package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/openjny/dotgh/internal/config"
)

// testConfig returns a config with the default test targets.
// This is used across test files to avoid duplication.
func testConfig() *config.Config {
	return &config.Config{
		Targets: config.DefaultTargets,
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

// executeCommand runs a cobra command and captures its output.
func executeCommand(t *testing.T, cmd interface {
	SetOut(*bytes.Buffer)
	SetErr(*bytes.Buffer)
	SetArgs([]string)
	Execute() error
}, args ...string) (string, error) {
	t.Helper()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}
