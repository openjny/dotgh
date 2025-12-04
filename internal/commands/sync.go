package commands

import (
	"github.com/openjny/dotgh/internal/config"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync configuration and templates across machines",
	Long: `Sync configuration and templates across machines using a Git repository.

This command provides subcommands to initialize, push, pull, and check the status
of your dotgh configuration synchronization.

Use 'dotgh sync init <repo>' to set up synchronization with a Git repository.
Use 'dotgh sync push' to push local changes to the remote repository.
Use 'dotgh sync pull' to pull changes from the remote repository.
Use 'dotgh sync status' to check the current sync status.`,
}

func init() {
	syncCmd.AddCommand(syncInitCmd)
	syncCmd.AddCommand(syncStatusCmd)
	syncCmd.AddCommand(syncPushCmd)
	syncCmd.AddCommand(syncPullCmd)
}

// NewSyncCmd creates a new sync command for testing.
func NewSyncCmd(configDir string) *cobra.Command {
	if configDir == "" {
		configDir = config.GetConfigDir()
	}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync configuration and templates across machines",
		Long: `Sync configuration and templates across machines using a Git repository.

This command provides subcommands to initialize, push, pull, and check the status
of your dotgh configuration synchronization.

Use 'dotgh sync init <repo>' to set up synchronization with a Git repository.
Use 'dotgh sync push' to push local changes to the remote repository.
Use 'dotgh sync pull' to pull changes from the remote repository.
Use 'dotgh sync status' to check the current sync status.`,
	}

	cmd.AddCommand(NewSyncInitCmd(configDir))
	cmd.AddCommand(NewSyncStatusCmd(configDir))
	cmd.AddCommand(NewSyncPushCmd(configDir))
	cmd.AddCommand(NewSyncPullCmd(configDir))

	return cmd
}
