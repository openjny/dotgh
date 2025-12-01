package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// defaultTargets defines the files/directories to copy from templates.
var defaultTargets = []string{
	".github",
	".vscode",
	"AGENTS.md",
}

// Command metadata constants
const (
	applyCmdUse   = "apply <template>"
	applyCmdShort = "Apply a template to the current directory"
	applyCmdLong  = "Apply a template to the current directory. Copies .github/, .vscode/, and AGENTS.md from the template."
)

var applyCmd = &cobra.Command{
	Use:   applyCmdUse,
	Short: applyCmdShort,
	Long:  applyCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runApply,
}

var forceFlag bool

func init() {
	applyCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Overwrite existing files")
}

// NewApplyCmd creates a new apply command with custom directories.
// This is primarily used for testing.
func NewApplyCmd(customTemplatesDir, customTargetDir string) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   applyCmdUse,
		Short: applyCmdShort,
		Long:  applyCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return applyTemplate(cmd, args[0], customTemplatesDir, customTargetDir, force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	return cmd
}

func runApply(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory: %w", err)
	}
	return applyTemplate(cmd, args[0], templatesDir, cwd, forceFlag)
}

// applyTemplate applies the specified template to the target directory.
func applyTemplate(cmd *cobra.Command, templateName, templatesDir, targetDir string, force bool) error {
	w := cmd.OutOrStdout()
	templatePath := filepath.Join(templatesDir, templateName)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	_, _ = fmt.Fprintf(w, "Applying template '%s'...\n", templateName)

	totalCopied := 0
	totalSkipped := 0

	for _, target := range defaultTargets {
		srcPath := filepath.Join(templatePath, target)
		dstPath := filepath.Join(targetDir, target)

		// Check if source exists in template
		srcInfo, err := os.Stat(srcPath)
		if os.IsNotExist(err) {
			continue // Target doesn't exist in template, skip
		}
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

	return nil
}

// copyDir recursively copies a directory from src to dst.
// Returns the number of files copied and skipped.
func copyDir(src, dst string, force bool) (copied, skipped int, err error) {
	err = filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			// Create directory if it doesn't exist
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("create directory %s: %w", dstPath, err)
			}
			return nil
		}

		// Copy file
		fileCopied, err := copyFile(path, dstPath, force)
		if err != nil {
			return err
		}
		if fileCopied {
			copied++
		} else {
			skipped++
		}
		return nil
	})

	return copied, skipped, err
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

// formatCopyResult formats the copy result for display.
func formatCopyResult(copied, skipped int) string {
	switch {
	case copied == 0 && skipped > 0:
		return "skipped, already exists"
	case skipped > 0:
		return fmt.Sprintf("%d files copied, %d skipped", copied, skipped)
	default:
		return fmt.Sprintf("%d files copied", copied)
	}
}
