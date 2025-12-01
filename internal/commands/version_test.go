package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/openjny/dotgh/internal/version"
)

func TestRunVersion(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		commit       string
		date         string
		wantContains []string
	}{
		{
			name:         "default values",
			version:      "dev",
			commit:       "none",
			date:         "unknown",
			wantContains: []string{"dotgh version dev", "commit: none", "built: unknown"},
		},
		{
			name:         "release version",
			version:      "1.0.0",
			commit:       "abc1234",
			date:         "2025-12-01",
			wantContains: []string{"dotgh version 1.0.0", "commit: abc1234", "built: 2025-12-01"},
		},
		{
			name:         "prerelease version",
			version:      "0.2.0-beta.1",
			commit:       "def5678",
			date:         "2025-11-15T10:30:00Z",
			wantContains: []string{"dotgh version 0.2.0-beta.1", "commit: def5678", "built: 2025-11-15T10:30:00Z"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values
			origVersion := version.Version
			origCommit := version.Commit
			origDate := version.Date

			// Set test values
			version.Version = tt.version
			version.Commit = tt.commit
			version.Date = tt.date

			// Restore original values after test
			t.Cleanup(func() {
				version.Version = origVersion
				version.Commit = origCommit
				version.Date = origDate
			})

			// Execute command
			cmd := NewVersionCmd()
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			err := cmd.Execute()

			if err != nil {
				t.Errorf("runVersion() error = %v", err)
				return
			}

			output := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("output should contain %q, got:\n%s", want, output)
				}
			}
		})
	}
}

func TestVersionOutputFormat(t *testing.T) {
	// Save original values
	origVersion := version.Version
	origCommit := version.Commit
	origDate := version.Date

	// Set known test values
	version.Version = "1.2.3"
	version.Commit = "abcdef0"
	version.Date = "2025-12-01"

	// Restore original values after test
	t.Cleanup(func() {
		version.Version = origVersion
		version.Commit = origCommit
		version.Date = origDate
	})

	cmd := NewVersionCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	err := cmd.Execute()

	if err != nil {
		t.Fatalf("runVersion() error = %v", err)
	}

	expected := "dotgh version 1.2.3 (commit: abcdef0, built: 2025-12-01)\n"
	if buf.String() != expected {
		t.Errorf("output format mismatch\ngot:  %q\nwant: %q", buf.String(), expected)
	}
}
