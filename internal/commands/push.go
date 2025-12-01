package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/glob"
	"github.com/spf13/cobra"
)

// Command metadata constants for push
const (
	pushCmdUse   = "push <template>"
	pushCmdShort = "Save the current directory's settings as a template"
	pushCmdLong  = "Save the current directory's settings as a template. Copies files matching configured patterns to the template directory."
)

var pushCmd = &cobra.Command{
	Use:   pushCmdUse,
	Short: pushCmdShort,
	Long:  pushCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runPush,
}

var pushForceFlag bool

func init() {
	pushCmd.Flags().BoolVarP(&pushForceFlag, "force", "f", false, "Overwrite existing files in the template")
}

// NewPushCmd creates a new push command with custom directories.
// This is primarily used for testing.
func NewPushCmd(customTemplatesDir, customSourceDir string) *cobra.Command {
	return NewPushCmdWithConfig(customTemplatesDir, customSourceDir, nil)
}

// NewPushCmdWithConfig creates a new push command with custom directories and config.
// This is primarily used for testing.
func NewPushCmdWithConfig(customTemplatesDir, customSourceDir string, cfg *config.Config) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   pushCmdUse,
		Short: pushCmdShort,
		Long:  pushCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pushTemplate(cmd, args[0], customTemplatesDir, customSourceDir, force, cfg)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files in the template")
	return cmd
}

func runPush(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	return pushTemplate(cmd, args[0], templatesDir, cwd, pushForceFlag, nil)
}

// pushTemplate saves the current directory's target files to a template.
func pushTemplate(cmd *cobra.Command, templateName, templatesDir, sourceDir string, force bool, cfg *config.Config) error {
	w := cmd.OutOrStdout()
	templatePath := filepath.Join(templatesDir, templateName)

	// Load config if not provided
	if cfg == nil {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	// Expand glob patterns to get actual files in source directory
	files, err := glob.ExpandPatterns(sourceDir, cfg.Targets)
	if err != nil {
		return fmt.Errorf("expand patterns: %w", err)
	}

	// Check if any files exist
	if len(files) == 0 {
		_, _ = fmt.Fprintf(w, "No target files found in current directory.\n")
		_, _ = fmt.Fprintf(w, "Configured patterns: %v\n", cfg.Targets)
		return nil
	}

	_, _ = fmt.Fprintf(w, "Pushing to template '%s'...\n", templateName)

	// Create template directory if it doesn't exist
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		return fmt.Errorf("create template directory: %w", err)
	}

	totalCopied := 0
	totalSkipped := 0

	for _, file := range files {
		srcPath := filepath.Join(sourceDir, file)
		dstPath := filepath.Join(templatePath, file)

		copied, err := copyFile(srcPath, dstPath, force)
		if err != nil {
			return fmt.Errorf("copy %s: %w", file, err)
		}
		if copied {
			totalCopied++
			_, _ = fmt.Fprintf(w, "  %s (copied)\n", file)
		} else {
			totalSkipped++
			_, _ = fmt.Fprintf(w, "  %s (skipped, already exists)\n", file)
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Done: %d file(s) copied, %d skipped\n", totalCopied, totalSkipped)
	_, _ = fmt.Fprintf(w, "Template saved to: %s\n", templatePath)

	return nil
}
