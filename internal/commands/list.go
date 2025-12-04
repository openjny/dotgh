package commands

import (
	"fmt"
	"os"

	"github.com/openjny/dotgh/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display a list of available templates",
	Long:  `Display a list of available templates stored in the configuration directory.`,
	RunE:  runList,
}

// NewListCmd creates a new list command with a custom templates directory.
// This is primarily used for testing.
func NewListCmd(customTemplatesDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display a list of available templates",
		Long:  `Display a list of available templates stored in the configuration directory.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return listTemplates(cmd, customTemplatesDir)
		},
	}
	return cmd
}

func runList(cmd *cobra.Command, args []string) error {
	// Load config to get templates directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	return listTemplates(cmd, cfg.GetTemplatesDir())
}

// listTemplates scans the templates directory and displays available templates.
func listTemplates(cmd *cobra.Command, dir string) error {
	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintln(w, "Available templates:")

	templates, err := scanTemplates(dir)
	if err != nil {
		// Directory doesn't exist or can't be read - show no templates
		_, _ = fmt.Fprintln(w, "  (no templates found)")
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "Template directory: %s\n", dir)
		return nil
	}

	if len(templates) == 0 {
		_, _ = fmt.Fprintln(w, "  (no templates found)")
		_, _ = fmt.Fprintln(w)
		_, _ = fmt.Fprintf(w, "Template directory: %s\n", dir)
		return nil
	}

	for _, tmpl := range templates {
		_, _ = fmt.Fprintf(w, "  %s\n", tmpl)
	}
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintf(w, "%d template(s) found\n", len(templates))

	return nil
}

// scanTemplates reads the templates directory and returns a list of template names.
// Only directories are considered as templates (files are ignored).
func scanTemplates(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var templates []string
	for _, entry := range entries {
		if entry.IsDir() {
			templates = append(templates, entry.Name())
		}
	}

	return templates, nil
}
