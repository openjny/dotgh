package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/editor"
	"github.com/spf13/cobra"
)

// editCmdLong is the long description for the edit command.
const editCmdLong = `Open a template directory or the templates directory in the user's preferred editor.

If a template name is provided, opens that specific template directory.
If no argument is provided, opens the templates directory itself.

The editor is determined in the following order:
1. 'editor' field in config.yaml
2. VISUAL environment variable
3. EDITOR environment variable
4. GIT_EDITOR environment variable
5. Platform default (vi on Linux/macOS, notepad on Windows)`

var editCmd = &cobra.Command{
	Use:   "edit [template]",
	Short: "Open template in the user's preferred editor",
	Long:  editCmdLong,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	return runEditWithDirs(cmd, args, templatesDir, config.GetConfigDir())
}

func runEditWithDirs(cmd *cobra.Command, args []string, templatesDir, configDir string) error {
	var targetPath string

	if len(args) == 0 {
		// No argument: open templates directory itself
		info, err := os.Stat(templatesDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("templates directory not found: %s", templatesDir)
			}
			return fmt.Errorf("check templates directory: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("templates directory not found: %s", templatesDir)
		}
		targetPath = templatesDir
	} else {
		// Argument provided: open specific template
		templateName := args[0]
		path, err := getTemplatePath(templatesDir, templateName)
		if err != nil {
			return err
		}
		targetPath = path
	}

	// Load config to get editor setting
	cfg, err := config.LoadFromDir(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build and execute editor command (use ForDir since we're opening a directory)
	editorArgs := buildEditorCommandForDir(cfg.Editor, targetPath)
	execCmd := exec.Command(editorArgs[0], editorArgs[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// getTemplatePath returns the path to the template directory.
// It returns an error if the template doesn't exist or is not a directory.
func getTemplatePath(templatesDir, templateName string) (string, error) {
	templatePath := filepath.Join(templatesDir, templateName)

	info, err := os.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("template %q not found", templateName)
		}
		return "", fmt.Errorf("check template: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("template %q not found", templateName)
	}

	return templatePath, nil
}

// NewEditCmd creates a new edit command with custom directories.
// This is primarily used for testing.
func NewEditCmd(customTemplatesDir, configDir string) *cobra.Command {
	return &cobra.Command{
		Use:   "edit [template]",
		Short: "Open template in the user's preferred editor",
		Long:  editCmdLong,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEditWithDirs(cmd, args, customTemplatesDir, configDir)
		},
	}
}

// NewEditCmdWithConfig is an alias for NewEditCmd for consistency with other commands.
func NewEditCmdWithConfig(customTemplatesDir, configDir string) *cobra.Command {
	return NewEditCmd(customTemplatesDir, configDir)
}

// buildEditorCommandForDir returns the command arguments to launch the editor for a directory.
// Unlike buildEditorCommand, it does not add --wait flag since GUI editors don't support
// waiting for directories to be closed.
func buildEditorCommandForDir(configEditor, target string) []string {
	editorStr := editor.Detect(configEditor)
	return editor.PrepareCommandForDir(editorStr, target)
}
