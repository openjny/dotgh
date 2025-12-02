// Package editor provides editor detection and launching functionality.
package editor

import (
	"os"
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	tests := []struct {
		name           string
		configEditor   string
		envVars        map[string]string
		expectedEditor string
	}{
		{
			name:           "config editor takes priority",
			configEditor:   "vim",
			envVars:        map[string]string{"VISUAL": "nano", "EDITOR": "emacs"},
			expectedEditor: "vim",
		},
		{
			name:           "VISUAL takes priority over EDITOR",
			configEditor:   "",
			envVars:        map[string]string{"VISUAL": "nano", "EDITOR": "emacs"},
			expectedEditor: "nano",
		},
		{
			name:           "EDITOR takes priority over GIT_EDITOR",
			configEditor:   "",
			envVars:        map[string]string{"EDITOR": "emacs", "GIT_EDITOR": "vim"},
			expectedEditor: "emacs",
		},
		{
			name:           "GIT_EDITOR is used as fallback",
			configEditor:   "",
			envVars:        map[string]string{"GIT_EDITOR": "vim"},
			expectedEditor: "vim",
		},
		{
			name:           "platform default when no editor set",
			configEditor:   "",
			envVars:        map[string]string{},
			expectedEditor: platformDefault(),
		},
		{
			name:           "config editor with arguments",
			configEditor:   "code --wait",
			envVars:        map[string]string{},
			expectedEditor: "code --wait",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and clear environment variables
			savedEnvVars := saveAndClearEnvVars(t, "VISUAL", "EDITOR", "GIT_EDITOR")
			defer restoreEnvVars(savedEnvVars)

			// Set test environment variables
			for k, v := range tt.envVars {
				_ = os.Setenv(k, v)
			}

			editor := Detect(tt.configEditor)
			if editor != tt.expectedEditor {
				t.Errorf("Detect() = %q, want %q", editor, tt.expectedEditor)
			}
		})
	}
}

func TestPrepareCommand(t *testing.T) {
	tests := []struct {
		name         string
		editor       string
		target       string
		expectedArgs []string
	}{
		{
			name:         "simple editor",
			editor:       "vim",
			target:       "/path/to/file",
			expectedArgs: []string{"vim", "/path/to/file"},
		},
		{
			name:         "editor with arguments",
			editor:       "code --wait",
			target:       "/path/to/file",
			expectedArgs: []string{"code", "--wait", "/path/to/file"},
		},
		{
			name:         "code without wait flag gets it added",
			editor:       "code",
			target:       "/path/to/file",
			expectedArgs: []string{"code", "--wait", "/path/to/file"},
		},
		{
			name:         "code-insiders without wait flag gets it added",
			editor:       "code-insiders",
			target:       "/path/to/file",
			expectedArgs: []string{"code-insiders", "--wait", "/path/to/file"},
		},
		{
			name:         "subl without wait flag gets it added",
			editor:       "subl",
			target:       "/path/to/file",
			expectedArgs: []string{"subl", "--wait", "/path/to/file"},
		},
		{
			name:         "sublime_text without wait flag gets it added",
			editor:       "sublime_text",
			target:       "/path/to/file",
			expectedArgs: []string{"sublime_text", "--wait", "/path/to/file"},
		},
		{
			name:         "code with wait flag already present",
			editor:       "code --wait",
			target:       "/path/to/file",
			expectedArgs: []string{"code", "--wait", "/path/to/file"},
		},
		{
			name:         "editor with multiple arguments",
			editor:       "vim -u NONE",
			target:       "/path/to/file",
			expectedArgs: []string{"vim", "-u", "NONE", "/path/to/file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := PrepareCommand(tt.editor, tt.target)
			if len(args) != len(tt.expectedArgs) {
				t.Errorf("PrepareCommand() returned %d args, want %d: got %v, want %v",
					len(args), len(tt.expectedArgs), args, tt.expectedArgs)
				return
			}
			for i, arg := range args {
				if arg != tt.expectedArgs[i] {
					t.Errorf("PrepareCommand()[%d] = %q, want %q", i, arg, tt.expectedArgs[i])
				}
			}
		})
	}
}

func TestNeedsWaitFlag(t *testing.T) {
	tests := []struct {
		name     string
		editor   string
		expected bool
	}{
		{"code needs wait", "code", true},
		{"code-insiders needs wait", "code-insiders", true},
		{"subl needs wait", "subl", true},
		{"sublime_text needs wait", "sublime_text", true},
		{"atom needs wait", "atom", true},
		{"vim does not need wait", "vim", false},
		{"nano does not need wait", "nano", false},
		{"emacs does not need wait", "emacs", false},
		{"notepad does not need wait", "notepad", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := needsWaitFlag(tt.editor); got != tt.expected {
				t.Errorf("needsWaitFlag(%q) = %v, want %v", tt.editor, got, tt.expected)
			}
		})
	}
}

func TestPlatformDefault(t *testing.T) {
	def := platformDefault()
	switch runtime.GOOS {
	case "windows":
		if def != "notepad" {
			t.Errorf("platformDefault() on Windows = %q, want %q", def, "notepad")
		}
	default:
		if def != "vi" {
			t.Errorf("platformDefault() on %s = %q, want %q", runtime.GOOS, def, "vi")
		}
	}
}

// saveAndClearEnvVars saves the current values and clears the specified environment variables
func saveAndClearEnvVars(t *testing.T, keys ...string) map[string]string {
	t.Helper()
	saved := make(map[string]string)
	for _, key := range keys {
		saved[key] = os.Getenv(key)
		_ = os.Unsetenv(key)
	}
	return saved
}

// restoreEnvVars restores the environment variables to their saved values
func restoreEnvVars(saved map[string]string) {
	for key, value := range saved {
		if value == "" {
			_ = os.Unsetenv(key)
		} else {
			_ = os.Setenv(key, value)
		}
	}
}
