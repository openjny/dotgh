// Package editor provides editor detection and launching functionality.
package editor

import (
	"os"
	"runtime"
	"strings"
)

// guiEditors is a list of editors that need the --wait flag
var guiEditors = []string{"code", "code-insiders", "subl", "sublime_text", "atom"}

// Detect returns the editor to use based on configuration and environment.
// Priority order:
// 1. configEditor (from config.yaml)
// 2. VISUAL environment variable
// 3. EDITOR environment variable
// 4. GIT_EDITOR environment variable
// 5. Platform-specific fallback (vi for Unix, notepad for Windows)
func Detect(configEditor string) string {
	if configEditor != "" {
		return configEditor
	}

	if visual := os.Getenv("VISUAL"); visual != "" {
		return visual
	}

	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	if gitEditor := os.Getenv("GIT_EDITOR"); gitEditor != "" {
		return gitEditor
	}

	return platformDefault()
}

// PrepareCommand returns the command arguments to launch the editor with the target.
// It automatically adds --wait flag for GUI editors if not already present.
func PrepareCommand(editor, target string) []string {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return []string{platformDefault(), target}
	}

	editorName := parts[0]
	args := parts[1:]

	// Add --wait flag for GUI editors if not already present
	if needsWaitFlag(editorName) && !hasWaitFlag(args) {
		args = append(args, "--wait")
	}

	result := make([]string, 0, len(args)+2)
	result = append(result, editorName)
	result = append(result, args...)
	result = append(result, target)

	return result
}

// platformDefault returns the default editor for the current platform.
func platformDefault() string {
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

// needsWaitFlag returns true if the editor needs the --wait flag.
func needsWaitFlag(editor string) bool {
	for _, guiEditor := range guiEditors {
		if editor == guiEditor {
			return true
		}
	}
	return false
}

// hasWaitFlag returns true if the arguments already contain --wait or -w.
func hasWaitFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--wait" || arg == "-w" {
			return true
		}
	}
	return false
}
