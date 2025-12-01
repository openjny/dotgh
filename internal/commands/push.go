package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Command metadata constants for push
const (
	pushCmdUse   = "push <template>"
	pushCmdShort = "Save the current directory's settings as a template"
	pushCmdLong  = "Save the current directory's settings as a template. Copies .github/, .vscode/, and AGENTS.md to the template directory."
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
	var force bool
	cmd := &cobra.Command{
		Use:   pushCmdUse,
		Short: pushCmdShort,
		Long:  pushCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pushTemplate(cmd, args[0], customTemplatesDir, customSourceDir, force)
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
	return pushTemplate(cmd, args[0], templatesDir, cwd, pushForceFlag)
}

// pushTemplate saves the current directory's target files to a template.
func pushTemplate(cmd *cobra.Command, templateName, templatesDir, sourceDir string, force bool) error {
	w := cmd.OutOrStdout()
	templatePath := filepath.Join(templatesDir, templateName)

	// Scan for target files in source directory
	var targetsFound []string
	for _, target := range defaultTargets {
		srcPath := filepath.Join(sourceDir, target)
		if _, err := os.Stat(srcPath); err == nil {
			targetsFound = append(targetsFound, target)
		}
	}

	// Check if any targets exist
	if len(targetsFound) == 0 {
		_, _ = fmt.Fprintf(w, "No target files found in current directory.\n")
		_, _ = fmt.Fprintf(w, "Targets: %v\n", defaultTargets)
		return nil
	}

	_, _ = fmt.Fprintf(w, "Pushing to template '%s'...\n", templateName)

	// Create template directory if it doesn't exist
	if err := os.MkdirAll(templatePath, 0755); err != nil {
		return fmt.Errorf("create template directory: %w", err)
	}

	totalCopied := 0
	totalSkipped := 0

	for _, target := range targetsFound {
		srcPath := filepath.Join(sourceDir, target)
		dstPath := filepath.Join(templatePath, target)

		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", target, err)
		}

		if srcInfo.IsDir() {
			copied, skipped, err := copyDir(srcPath, dstPath, force)
			if err != nil {
				return fmt.Errorf("copy %s: %w", target, err)
			}
			totalCopied += copied
			totalSkipped += skipped
			_, _ = fmt.Fprintf(w, "  %s/ (%s)\n", target, formatCopyResult(copied, skipped))
		} else {
			copied, err := copyFile(srcPath, dstPath, force)
			if err != nil {
				return fmt.Errorf("copy %s: %w", target, err)
			}
			if copied {
				totalCopied++
				_, _ = fmt.Fprintf(w, "  %s (copied)\n", target)
			} else {
				totalSkipped++
				_, _ = fmt.Fprintf(w, "  %s (skipped, already exists)\n", target)
			}
		}
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Done: %d file(s) copied, %d skipped\n", totalCopied, totalSkipped)
	_, _ = fmt.Fprintf(w, "Template saved to: %s\n", templatePath)

	return nil
}
