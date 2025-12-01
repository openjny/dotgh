package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display a list of available templates",
	Long:  `Display a list of available templates stored in the configuration directory.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	// TODO: Implement template listing
	fmt.Fprintln(cmd.OutOrStdout(), "Available templates:")
	fmt.Fprintln(cmd.OutOrStdout(), "  (no templates found)")
	return nil
}
