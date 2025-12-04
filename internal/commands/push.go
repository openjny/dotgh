package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/diff"
	"github.com/openjny/dotgh/internal/prompt"
	"github.com/spf13/cobra"
)

// Command metadata constants for push
const (
	pushCmdUse   = "push <template>"
	pushCmdShort = "Save the current directory's settings as a template"
	pushCmdLong  = `Save the current directory's settings as a template with Git-style sync behavior.

By default, performs a full sync: adds new files, updates modified files, and
deletes files in the template that don't exist locally.

Use --merge to only add and update files without deleting.
Use --yes to skip the confirmation prompt.

If the template doesn't exist, it will be created.

Examples:
  dotgh push my-template          # Full sync with confirmation
  dotgh push my-template --yes    # Full sync without confirmation
  dotgh push my-template --merge  # Merge only (no deletions)`
)

var pushCmd = &cobra.Command{
	Use:   pushCmdUse,
	Short: pushCmdShort,
	Long:  pushCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runPush,
}

var (
	pushMergeFlag bool
	pushYesFlag   bool
)

func init() {
	pushCmd.Flags().BoolVarP(&pushMergeFlag, "merge", "m", false, "Merge mode: only add/update files, no deletions")
	pushCmd.Flags().BoolVarP(&pushYesFlag, "yes", "y", false, "Skip confirmation prompt")
}

// PushOptions contains options for the push command.
type PushOptions struct {
	MergeMode bool
	Yes       bool
	Stdin     io.Reader
}

// NewPushCmd creates a new push command with custom directories.
// This is primarily used for testing.
func NewPushCmd(customTemplatesDir, customSourceDir string) *cobra.Command {
	return NewPushCmdWithConfig(customTemplatesDir, customSourceDir, nil)
}

// NewPushCmdWithConfig creates a new push command with custom directories and config.
// This is primarily used for testing.
func NewPushCmdWithConfig(customTemplatesDir, customSourceDir string, cfg *config.Config) *cobra.Command {
	return NewPushCmdWithOptions(customTemplatesDir, customSourceDir, cfg, nil)
}

// NewPushCmdWithOptions creates a new push command with custom directories, config, and options.
// This is primarily used for testing with custom stdin.
func NewPushCmdWithOptions(customTemplatesDir, customSourceDir string, cfg *config.Config, defaultOpts *PushOptions) *cobra.Command {
	var merge, yes bool
	cmd := &cobra.Command{
		Use:   pushCmdUse,
		Short: pushCmdShort,
		Long:  pushCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := PushOptions{
				MergeMode: merge,
				Yes:       yes,
				Stdin:     cmd.InOrStdin(),
			}
			if defaultOpts != nil {
				if defaultOpts.Stdin != nil {
					opts.Stdin = defaultOpts.Stdin
				}
			}
			return pushTemplate(cmd, args[0], customTemplatesDir, customSourceDir, opts, cfg)
		},
	}
	cmd.Flags().BoolVarP(&merge, "merge", "m", false, "Merge mode: only add/update files, no deletions")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompt")
	return cmd
}

func runPush(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}

	// Load config to get templates directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	opts := PushOptions{
		MergeMode: pushMergeFlag,
		Yes:       pushYesFlag,
		Stdin:     cmd.InOrStdin(),
	}

	return pushTemplate(cmd, args[0], cfg.GetTemplatesDir(), cwd, opts, cfg)
}

// pushTemplate saves the current directory's target files to a template.
func pushTemplate(cmd *cobra.Command, templateName, templatesDir, sourceDir string, opts PushOptions, cfg *config.Config) error {
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

	// Check if template exists
	templateExists := true
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		templateExists = false
	}

	// Compute diff (source -> template)
	diffResult, err := diff.ComputeDiff(sourceDir, templatePath, cfg.Includes, cfg.Excludes, opts.MergeMode)
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
	if templateExists {
		_, _ = fmt.Fprintf(w, "Pushing to template '%s' (%s):\n", templateName, mode)
	} else {
		_, _ = fmt.Fprintf(w, "Creating template '%s':\n", templateName)
	}
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

	// Create template directory if it doesn't exist
	if !templateExists {
		if err := os.MkdirAll(templatePath, 0755); err != nil {
			return fmt.Errorf("create template directory: %w", err)
		}
	}

	// Apply changes
	if err := diff.ApplyChanges(sourceDir, templatePath, diffResult); err != nil {
		return fmt.Errorf("apply changes: %w", err)
	}

	// Print result
	_, _ = fmt.Fprintln(w)
	printApplySummary(w, diffResult)
	_, _ = fmt.Fprintf(w, "Template saved to: %s\n", templatePath)

	return nil
}
