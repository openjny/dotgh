package commands

import (
	"fmt"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/sync"
	"github.com/spf13/cobra"
)

var syncStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status",
	Long: `Show the current synchronization status.

Displays information about the sync repository, current branch,
and whether there are any uncommitted local changes.`,
	RunE: runSyncStatus,
}

func runSyncStatus(cmd *cobra.Command, args []string) error {
	return runSyncStatusWithDir(cmd, config.GetConfigDir())
}

func runSyncStatusWithDir(cmd *cobra.Command, configDir string) error {
	w := cmd.OutOrStdout()

	manager := sync.NewManager(configDir)
	status, err := manager.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("get sync status: %w", err)
	}

	if status.State == sync.StatusNotInitialized {
		_, _ = fmt.Fprintln(w, "Sync is not initialized.")
		_, _ = fmt.Fprintln(w, "Run 'dotgh sync init <repository>' to set up synchronization.")
		return nil
	}

	_, _ = fmt.Fprintln(w, "Sync Status:")
	_, _ = fmt.Fprintf(w, "  Repository: %s\n", status.RepoURL)
	_, _ = fmt.Fprintf(w, "  Branch: %s\n", status.Branch)
	_, _ = fmt.Fprintf(w, "  Status: %s\n", status.State)
	_, _ = fmt.Fprintf(w, "  Sync directory: %s\n", manager.SyncDirPath())

	if status.HasChanges {
		_, _ = fmt.Fprintln(w, "\nUncommitted changes:")
		for _, change := range status.Changes {
			_, _ = fmt.Fprintf(w, "  - %s\n", change)
		}
	}

	return nil
}

// NewSyncStatusCmd creates a new sync status command for testing.
func NewSyncStatusCmd(configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show sync status",
		Long: `Show the current synchronization status.

Displays information about the sync repository, current branch,
and whether there are any uncommitted local changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncStatusWithDir(cmd, configDir)
		},
	}
	return cmd
}
