package commands

import (
	"fmt"
	"time"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/sync"
	"github.com/spf13/cobra"
)

var syncPushMessage string

var syncPushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local config and templates to remote",
	Long: `Push local configuration and templates to the remote repository.

This command copies your local config.yaml and templates directory to the
sync repository, commits the changes, and pushes to the remote.

Examples:
  dotgh sync push
  dotgh sync push -m "Update templates"`,
	RunE: runSyncPush,
}

func init() {
	syncPushCmd.Flags().StringVarP(&syncPushMessage, "message", "m", "", "Commit message (default: auto-generated)")
}

func runSyncPush(cmd *cobra.Command, args []string) error {
	return runSyncPushWithDir(cmd, config.GetConfigDir())
}

func runSyncPushWithDir(cmd *cobra.Command, configDir string) error {
	w := cmd.OutOrStdout()

	manager := sync.NewManager(configDir)

	// Check if initialized
	if !manager.IsInitialized() {
		return fmt.Errorf("sync is not initialized. Run 'dotgh sync init <repository>' first")
	}

	// Copy config and templates to sync directory
	if err := manager.CopyConfigToSync(); err != nil {
		return fmt.Errorf("copy config: %w", err)
	}

	if err := manager.CopyTemplatesToSync(); err != nil {
		return fmt.Errorf("copy templates: %w", err)
	}

	// Check if there are changes to commit
	status, err := manager.GetSyncStatus()
	if err != nil {
		return fmt.Errorf("get status: %w", err)
	}

	if !status.HasChanges {
		_, _ = fmt.Fprintln(w, "Nothing to push. Local config and templates are in sync.")
		return nil
	}

	// Generate commit message if not provided
	message := syncPushMessage
	if message == "" {
		message = fmt.Sprintf("Sync update: %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	// Commit and push
	if err := manager.StageAndCommit(message); err != nil {
		return fmt.Errorf("commit changes: %w", err)
	}

	if err := manager.Push(); err != nil {
		return fmt.Errorf("push to remote: %w", err)
	}

	_, _ = fmt.Fprintln(w, "Pushed successfully!")
	_, _ = fmt.Fprintf(w, "  Commit message: %s\n", message)
	if len(status.Changes) > 0 {
		_, _ = fmt.Fprintln(w, "  Changes:")
		for _, change := range status.Changes {
			_, _ = fmt.Fprintf(w, "    - %s\n", change)
		}
	}

	return nil
}

// NewSyncPushCmd creates a new sync push command for testing.
func NewSyncPushCmd(configDir string) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push local config and templates to remote",
		Long: `Push local configuration and templates to the remote repository.

This command copies your local config.yaml and templates directory to the
sync repository, commits the changes, and pushes to the remote.

Examples:
  dotgh sync push
  dotgh sync push -m "Update templates"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Temporarily set the global variable
			oldMessage := syncPushMessage
			syncPushMessage = message
			defer func() { syncPushMessage = oldMessage }()

			return runSyncPushWithDir(cmd, configDir)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Commit message (default: auto-generated)")
	return cmd
}
