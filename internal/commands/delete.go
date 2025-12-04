package commands

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/openjny/dotgh/internal/config"
	"github.com/spf13/cobra"
)

// Command metadata constants for delete
const (
	deleteCmdUse   = "delete <template>"
	deleteCmdShort = "Delete a template"
	deleteCmdLong  = "Delete a template from the templates directory. Shows a confirmation prompt unless --force is specified."
)

var deleteCmd = &cobra.Command{
	Use:   deleteCmdUse,
	Short: deleteCmdShort,
	Long:  deleteCmdLong,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

var deleteForceFlag bool

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForceFlag, "force", "f", false, "Skip confirmation prompt")
}

// NewDeleteCmd creates a new delete command with custom templates directory and stdin.
// This is primarily used for testing.
func NewDeleteCmd(customTemplatesDir string, stdin io.Reader) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   deleteCmdUse,
		Short: deleteCmdShort,
		Long:  deleteCmdLong,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTemplate(cmd, args[0], customTemplatesDir, stdin, force)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")
	return cmd
}

func runDelete(cmd *cobra.Command, args []string) error {
	// Load config to get templates directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	return deleteTemplate(cmd, args[0], cfg.GetTemplatesDir(), os.Stdin, deleteForceFlag)
}

// deleteTemplate deletes the specified template.
func deleteTemplate(cmd *cobra.Command, templateName, templatesDir string, stdin io.Reader, force bool) error {
	w := cmd.OutOrStdout()
	templatePath := filepath.Join(templatesDir, templateName)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	// Confirm deletion unless force flag is set
	if !force {
		if !confirmDelete(stdin, w, templateName) {
			_, _ = fmt.Fprintln(w, "Deletion cancelled.")
			return nil
		}
	}

	// Delete the template directory
	if err := os.RemoveAll(templatePath); err != nil {
		return fmt.Errorf("delete template: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Template '%s' deleted.\n", templateName)
	return nil
}

// confirmDelete prompts the user for confirmation and returns true if confirmed.
// Default is "no" (returns false on empty input).
func confirmDelete(r io.Reader, w io.Writer, templateName string) bool {
	_, _ = fmt.Fprintf(w, "Delete template '%s'? (y/N): ", templateName)

	reader := bufio.NewReader(r)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
