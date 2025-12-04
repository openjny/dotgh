package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/diff"
	"github.com/openjny/dotgh/internal/prompt"
	"github.com/spf13/cobra"
)

// Command metadata constants
const (
	pullCmdUse   = "pull <template>"
	pullCmdShort = "Pull a template to the current directory"
	pullCmdLong  = `Pull a template to the current directory with Git-style sync behavior.

By default, performs a full sync: adds new files, updates modified files, and
deletes files that exist locally but not in the template.

Use --merge to only add and update files without deleting.
Use --yes to skip the confirmation prompt.

Examples:
  dotgh pull my-template          # Full sync with confirmation
  dotgh pull my-template --yes    # Full sync without confirmation  
  dotgh pull my-template --merge  # Merge only (no deletions)`
)

var pullCmd = &cobra.Command{
	Use:   pullCmdUse,
	Short: pullCmdShort,
	Long:  pullCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runPull,
}

var (
	pullMergeFlag bool
	pullYesFlag   bool
)

func init() {
	pullCmd.Flags().BoolVarP(&pullMergeFlag, "merge", "m", false, "Merge mode: only add/update files, no deletions")
	pullCmd.Flags().BoolVarP(&pullYesFlag, "yes", "y", false, "Skip confirmation prompt")
}

// PullOptions contains options for the pull command.
type PullOptions struct {
	MergeMode bool
	Yes       bool
	Stdin     io.Reader
}

// NewPullCmd creates a new pull command with custom directories.
// This is primarily used for testing.
func NewPullCmd(customTemplatesDir, customTargetDir string) *cobra.Command {
	return NewPullCmdWithConfig(customTemplatesDir, customTargetDir, nil)
}

// NewPullCmdWithConfig creates a new pull command with custom directories and config.
// This is primarily used for testing.
func NewPullCmdWithConfig(customTemplatesDir, customTargetDir string, cfg *config.Config) *cobra.Command {
	return NewPullCmdWithOptions(customTemplatesDir, customTargetDir, cfg, nil)
}

// NewPullCmdWithOptions creates a new pull command with custom directories, config, and options.
// This is primarily used for testing with custom stdin.
func NewPullCmdWithOptions(customTemplatesDir, customTargetDir string, cfg *config.Config, defaultOpts *PullOptions) *cobra.Command {
	var merge, yes bool
	cmd := &cobra.Command{
		Use:   pullCmdUse,
		Short: pullCmdShort,
		Long:  pullCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := PullOptions{
				MergeMode: merge,
				Yes:       yes,
				Stdin:     cmd.InOrStdin(),
			}
			if defaultOpts != nil {
				if defaultOpts.Stdin != nil {
					opts.Stdin = defaultOpts.Stdin
				}
			}
			return pullTemplate(cmd, args[0], customTemplatesDir, customTargetDir, opts, cfg)
		},
	}
	cmd.Flags().BoolVarP(&merge, "merge", "m", false, "Merge mode: only add/update files, no deletions")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func runPull(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Load config to get templates directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	opts := PullOptions{
		MergeMode: pullMergeFlag,
		Yes:       pullYesFlag,
		Stdin:     cmd.InOrStdin(),
	}

	return pullTemplate(cmd, args[0], cfg.GetTemplatesDir(), cwd, opts, cfg)
}

// pullTemplate pulls the specified template to the target directory.
func pullTemplate(cmd *cobra.Command, templateName, templatesDir, targetDir string, opts PullOptions, cfg *config.Config) error {
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

	// Compute diff
	diffResult, err := diff.ComputeDiff(templatePath, targetDir, cfg.Includes, cfg.Excludes, opts.MergeMode)
	if err != nil {
		return fmt.Errorf("compute diff: %w", err)
	}

	// Check if there are any changes
	if !diffResult.HasChanges() {
		_, _ = fmt.Fprintf(w, "Template '%s' is already in sync.\n", templateName)
		return nil
	}

	// Print diff summary
	mode := "full sync"
	if opts.MergeMode {
		mode = "merge"
	}
	_, _ = fmt.Fprintf(w, "Pulling template '%s' (%s):\n", templateName, mode)
	printDiffSummary(w, diffResult)

	// Ask for confirmation unless --yes is specified
	if !opts.Yes {
		confirmed, err := prompt.Confirm("Apply these changes?", true, w, opts.Stdin)
		if err != nil {
			return fmt.Errorf("confirmation: %w", err)
		}
		if !confirmed {
			_, _ = fmt.Fprintln(w, "Aborted.")
			return nil
		}
	}

	// Apply changes
	if err := diff.ApplyChanges(templatePath, targetDir, diffResult); err != nil {
		return fmt.Errorf("apply changes: %w", err)
	}

	// Print result
	_, _ = fmt.Fprintln(w)
	printApplySummary(w, diffResult)

	return nil
}

// printDiffSummary prints the diff summary to the writer.
func printDiffSummary(w io.Writer, d *diff.DiffResult) {
	for _, change := range d.Added {
		_, _ = fmt.Fprintf(w, "  + %s\n", change.Path)
	}
	for _, change := range d.Modified {
		_, _ = fmt.Fprintf(w, "  M %s\n", change.Path)
	}
	for _, change := range d.Deleted {
		_, _ = fmt.Fprintf(w, "  - %s\n", change.Path)
	}
	_, _ = fmt.Fprintln(w)
}

// printApplySummary prints the apply summary to the writer.
func printApplySummary(w io.Writer, d *diff.DiffResult) {
	var parts []string
	if len(d.Added) > 0 {
		parts = append(parts, fmt.Sprintf("%d added", len(d.Added)))
	}
	if len(d.Modified) > 0 {
		parts = append(parts, fmt.Sprintf("%d modified", len(d.Modified)))
	}
	if len(d.Deleted) > 0 {
		parts = append(parts, fmt.Sprintf("%d deleted", len(d.Deleted)))
	}
	_, _ = fmt.Fprintf(w, "Done: %s\n", strings.Join(parts, ", "))
}
