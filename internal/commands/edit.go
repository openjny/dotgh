package commands

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/editor"
	"github.com/openjny/dotgh/internal/prompt"
	"github.com/spf13/cobra"
)

// editCmdLong is the long description for the edit command.
const editCmdLong = `Open a template directory or the templates directory in the user's preferred editor.

If a template name is provided, opens that specific template directory.
If no argument is provided, opens the templates directory itself.
If the template doesn't exist, you can create it with the --create flag.

The editor is determined in the following order:
1. 'editor' field in config.yaml
2. VISUAL environment variable
3. EDITOR environment variable
4. GIT_EDITOR environment variable
5. Platform default (vi on Linux/macOS, notepad on Windows)

Examples:
  dotgh edit                      # Open templates directory
  dotgh edit my-template          # Open existing template
  dotgh edit new-template --create  # Create and open new template`

var editCmd = &cobra.Command{
	Use:   "edit [template]",
	Short: "Open template in the user's preferred editor",
	Long:  editCmdLong,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runEdit,
}

var editCreateFlag bool

func init() {
	editCmd.Flags().BoolVarP(&editCreateFlag, "create", "c", false, "Create template if it doesn't exist")
}

// EditOptions contains options for the edit command.
type EditOptions struct {
	Create bool
	Stdin  io.Reader
}

func runEdit(cmd *cobra.Command, args []string) error {
	// Load config to get templates directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	opts := EditOptions{
		Create: editCreateFlag,
		Stdin:  cmd.InOrStdin(),
	}
	return runEditWithConfig(cmd, args, cfg.GetTemplatesDir(), config.GetConfigDir(), cfg, opts)
}

func runEditWithConfig(cmd *cobra.Command, args []string, templatesDir, configDir string, cfg *config.Config, opts EditOptions) error {
	w := cmd.OutOrStdout()
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
			// Template doesn't exist - check if we should create it
			if opts.Create && strings.Contains(err.Error(), "not found") {
				templatePath := filepath.Join(templatesDir, templateName)

				// Create the template directory
				if err := os.MkdirAll(templatePath, 0755); err != nil {
					return fmt.Errorf("create template directory: %w", err)
				}

				_, _ = fmt.Fprintf(w, "Created new template: %s\n", templateName)
				targetPath = templatePath
			} else if strings.Contains(err.Error(), "not found") {
				// Offer to create if not using --create flag
				_, _ = fmt.Fprintf(w, "Template %q does not exist.\n", templateName)

				confirmed, promptErr := prompt.Confirm("Create it?", true, w, opts.Stdin)
				if promptErr != nil {
					return fmt.Errorf("confirmation: %w", promptErr)
				}
				if !confirmed {
					return fmt.Errorf("template %q not found", templateName)
				}

				templatePath := filepath.Join(templatesDir, templateName)
				if err := os.MkdirAll(templatePath, 0755); err != nil {
					return fmt.Errorf("create template directory: %w", err)
				}

				_, _ = fmt.Fprintf(w, "Created new template: %s\n", templateName)
				targetPath = templatePath
			} else {
				return err
			}
		} else {
			targetPath = path
		}
	}

	// Load config if not provided
	if cfg == nil {
		var err error
		cfg, err = config.LoadFromDir(configDir)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
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
	return NewEditCmdWithOptions(customTemplatesDir, configDir, nil)
}

// NewEditCmdWithOptions creates a new edit command with custom directories and options.
// This is primarily used for testing.
func NewEditCmdWithOptions(customTemplatesDir, configDir string, defaultOpts *EditOptions) *cobra.Command {
	var create bool
	cmd := &cobra.Command{
		Use:   "edit [template]",
		Short: "Open template in the user's preferred editor",
		Long:  editCmdLong,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := EditOptions{
				Create: create,
				Stdin:  cmd.InOrStdin(),
			}
			if defaultOpts != nil {
				if defaultOpts.Stdin != nil {
					opts.Stdin = defaultOpts.Stdin
				}
			}
			return runEditWithConfig(cmd, args, customTemplatesDir, configDir, nil, opts)
		},
	}
	cmd.Flags().BoolVarP(&create, "create", "c", false, "Create template if it doesn't exist")
	return cmd
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
