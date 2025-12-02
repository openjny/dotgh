package commands

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dotgh",
	Short: "A CLI tool to manage AI coding guidelines and templates",
	Long: `dotgh is a cross-platform CLI tool that allows you to easily apply,
update, and manage AI coding guidelines and configuration templates
across multiple projects.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(configCmd)
}
