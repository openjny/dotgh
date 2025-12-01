package commands

import (
	"fmt"

	"github.com/openjny/dotgh/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display the version, commit hash, and build date of dotgh.`,
	Run:   runVersion,
}

// NewVersionCmd creates a new version command.
// This is primarily used for testing.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Display the version, commit hash, and build date of dotgh.`,
		Run:   runVersion,
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dotgh version %s (commit: %s, built: %s)\n",
		version.Version, version.Commit, version.Date)
}
