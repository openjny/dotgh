package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/glob"
	"github.com/spf13/cobra"
)

// Command metadata constants
const (
	pullCmdUse   = "pull <template>"
	pullCmdShort = "Pull a template to the current directory"
	pullCmdLong  = "Pull a template to the current directory. Copies files matching configured patterns from the template."
)

var pullCmd = &cobra.Command{
	Use:   pullCmdUse,
	Short: pullCmdShort,
	Long:  pullCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runPull,
}

var pullForceFlag bool

func init() {
	pullCmd.Flags().BoolVarP(&pullForceFlag, "force", "f", false, "Overwrite existing files")
}

// NewPullCmd creates a new pull command with custom directories.
// This is primarily used for testing.
func NewPullCmd(customTemplatesDir, customTargetDir string) *cobra.Command {
	return NewPullCmdWithConfig(customTemplatesDir, customTargetDir, nil)
}

// NewPullCmdWithConfig creates a new pull command with custom directories and config.
// This is primarily used for testing.
func NewPullCmdWithConfig(customTemplatesDir, customTargetDir string, cfg *config.Config) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   pullCmdUse,
		Short: pullCmdShort,
		Long:  pullCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return pullTemplate(cmd, args[0], customTemplatesDir, customTargetDir, force, cfg)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
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

	return pullTemplate(cmd, args[0], cfg.GetTemplatesDir(), cwd, pullForceFlag, cfg)
}

// pullTemplate pulls the specified template to the target directory.
func pullTemplate(cmd *cobra.Command, templateName, templatesDir, targetDir string, force bool, cfg *config.Config) error {
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

	_, _ = fmt.Fprintf(w, "Pulling template '%s'...\n", templateName)

	// Expand glob patterns to get actual files in template
	files, err := glob.ExpandPatterns(templatePath, cfg.Includes)
	if err != nil {
		return fmt.Errorf("expand patterns: %w", err)
	}

	// Filter out excluded files
	files, err = glob.FilterExcludes(files, cfg.Excludes)
	if err != nil {
		return fmt.Errorf("filter excludes: %w", err)
	}

	if len(files) == 0 {
		_, _ = fmt.Fprintln(w, "  (no matching files found in template)")
		return nil
	}

	totalCopied := 0
	totalSkipped := 0

	for _, file := range files {
		srcPath := filepath.Join(templatePath, file)
		dstPath := filepath.Join(targetDir, file)

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

	return nil
}

// copyFile copies a single file from src to dst.
// Returns true if the file was copied, false if it was skipped.
func copyFile(src, dst string, force bool) (bool, error) {
	// Check if destination exists
	if _, err := os.Stat(dst); err == nil {
		if !force {
			return false, nil // Skip existing file
		}
	}

	// Ensure destination directory exists
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return false, fmt.Errorf("create directory %s: %w", dstDir, err)
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return false, fmt.Errorf("open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return false, fmt.Errorf("stat source file: %w", err)
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return false, fmt.Errorf("create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return false, fmt.Errorf("copy content: %w", err)
	}

	return true, nil
}
