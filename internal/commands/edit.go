package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <template>",
	Short: "Open template in the user's preferred editor",
	Long: `Open the specified template directory in the user's preferred editor.

The editor is determined in the following order:
1. 'editor' field in config.yaml
2. VISUAL environment variable
3. EDITOR environment variable
4. GIT_EDITOR environment variable
5. Platform default (vi on Linux/macOS, notepad on Windows)`,
	Args: cobra.ExactArgs(1),
	RunE: runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
	return runEditWithDirs(cmd, args, templatesDir, config.GetConfigDir())
}

func runEditWithDirs(cmd *cobra.Command, args []string, templatesDir, configDir string) error {
	templateName := args[0]

	// Get template path and validate it exists
	templatePath, err := getTemplatePath(templatesDir, templateName)
	if err != nil {
		return err
	}

	// Load config to get editor setting
	cfg, err := config.LoadFromDir(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build and execute editor command
	editorArgs := buildEditorCommand(cfg.Editor, templatePath)
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
	cmd := &cobra.Command{
		Use:   "edit <template>",
		Short: "Open template in the user's preferred editor",
		Long: `Open the specified template directory in the user's preferred editor.

The editor is determined in the following order:
1. 'editor' field in config.yaml
2. VISUAL environment variable
3. EDITOR environment variable
4. GIT_EDITOR environment variable
5. Platform default (vi on Linux/macOS, notepad on Windows)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEditWithDirs(cmd, args, customTemplatesDir, configDir)
		},
	}
	return cmd
}

// NewEditCmdWithConfig creates a new edit command with custom directories.
// Alias for NewEditCmd for consistency with other commands.
func NewEditCmdWithConfig(customTemplatesDir, configDir string) *cobra.Command {
	return NewEditCmd(customTemplatesDir, configDir)
}
