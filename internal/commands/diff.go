package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/diff"
	"github.com/spf13/cobra"
)

// Command metadata constants for diff
const (
	diffCmdUse   = "diff <template>"
	diffCmdShort = "Show differences between a template and the current directory"
	diffCmdLong  = `Show differences between a template and the current directory.

Displays files that would be added, modified, or deleted when pulling a template.
By default, shows what a full sync (pull) would do. Use --reverse to show what
a push would do.

Exit codes:
  0 - No differences found
  1 - Differences found or error occurred`
)

var diffCmd = &cobra.Command{
	Use:   diffCmdUse,
	Short: diffCmdShort,
	Long:  diffCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runDiff,
}

var (
	diffReverseFlag bool
	diffMergeFlag   bool
)

func init() {
	diffCmd.Flags().BoolVarP(&diffReverseFlag, "reverse", "r", false, "Show differences for push (current → template)")
	diffCmd.Flags().BoolVar(&diffMergeFlag, "merge", false, "Show merge mode differences (no deletions)")
}

// NewDiffCmd creates a new diff command with custom directories.
// This is primarily used for testing.
func NewDiffCmd(customTemplatesDir, customTargetDir string) *cobra.Command {
	return NewDiffCmdWithConfig(customTemplatesDir, customTargetDir, nil)
}

// NewDiffCmdWithConfig creates a new diff command with custom directories and config.
// This is primarily used for testing.
func NewDiffCmdWithConfig(customTemplatesDir, customTargetDir string, cfg *config.Config) *cobra.Command {
	var reverse, merge bool
	cmd := &cobra.Command{
		Use:   diffCmdUse,
		Short: diffCmdShort,
		Long:  diffCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDiffWithOptions(cmd, args[0], customTemplatesDir, customTargetDir, reverse, merge, cfg)
		},
	}
	cmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "Show differences for push (current → template)")
	cmd.Flags().BoolVar(&merge, "merge", false, "Show merge mode differences (no deletions)")
	return cmd
}

func runDiff(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	return runDiffWithOptions(cmd, args[0], cfg.GetTemplatesDir(), cwd, diffReverseFlag, diffMergeFlag, cfg)
}

// runDiffWithOptions runs the diff command with the specified options.
func runDiffWithOptions(cmd *cobra.Command, templateName, templatesDir, targetDir string, reverse, mergeMode bool, cfg *config.Config) error {
	w := cmd.OutOrStdout()
	templatePath := filepath.Join(templatesDir, templateName)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Load config if not provided
	if cfg == nil {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
	}

	var srcDir, dstDir string
	var direction string
	if reverse {
		// Push direction: current -> template
		srcDir = targetDir
		dstDir = templatePath
		direction = fmt.Sprintf("current directory → template '%s'", templateName)
	} else {
		// Pull direction: template -> current
		srcDir = templatePath
		dstDir = targetDir
		direction = fmt.Sprintf("template '%s' → current directory", templateName)
	}

	diffResult, err := diff.ComputeDiff(srcDir, dstDir, cfg.Includes, cfg.Excludes, mergeMode)
	if err != nil {
		return fmt.Errorf("compute diff: %w", err)
	}

	// Print header
	if mergeMode {
		_, _ = fmt.Fprintf(w, "Diff (%s, merge mode):\n", direction)
	} else {
		_, _ = fmt.Fprintf(w, "Diff (%s):\n", direction)
	}

	if !diffResult.HasChanges() {
		_, _ = fmt.Fprintln(w, "  (no changes)")
		return nil
	}

	// Print changes
	for _, change := range diffResult.Added {
		_, _ = fmt.Fprintf(w, "  + %s\n", change.Path)
	}
	for _, change := range diffResult.Modified {
		_, _ = fmt.Fprintf(w, "  M %s\n", change.Path)
	}
	for _, change := range diffResult.Deleted {
		_, _ = fmt.Fprintf(w, "  - %s\n", change.Path)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "Summary: %d addition(s), %d modification(s), %d deletion(s)\n",
		len(diffResult.Added), len(diffResult.Modified), len(diffResult.Deleted))

	// Return error to indicate differences found (exit code 1)
	return ErrDiffFound
}

// ErrDiffFound is returned when differences are found.
// This is used to set exit code 1.
var ErrDiffFound = errors.New("differences found")
