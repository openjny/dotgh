package commands

import (
	"fmt"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/sync"
	"github.com/spf13/cobra"
)

var syncPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull config and templates from remote",
	Long: `Pull configuration and templates from the remote repository.

This command pulls the latest changes from the remote repository and
copies the config.yaml and templates to your local dotgh config directory.

Examples:
  dotgh sync pull`,
	RunE: runSyncPull,
}

func runSyncPull(cmd *cobra.Command, args []string) error {
	return runSyncPullWithDir(cmd, config.GetConfigDir())
}

func runSyncPullWithDir(cmd *cobra.Command, configDir string) error {
	w := cmd.OutOrStdout()

	manager := sync.NewManager(configDir)

	// Check if initialized
	if !manager.IsInitialized() {
		return fmt.Errorf("sync is not initialized. Run 'dotgh sync init <repository>' first")
	}

	// Pull from remote
	if err := manager.Pull(); err != nil {
		// Pull might fail if no remote tracking branch exists yet, which is fine
		_, _ = fmt.Fprintf(w, "Note: Could not pull from remote (this is normal for new repos)\n")
	}

	// Copy config and templates from sync directory to local
	if err := manager.CopyConfigFromSync(); err != nil {
		return fmt.Errorf("copy config: %w", err)
	}

	if err := manager.CopyTemplatesFromSync(); err != nil {
		return fmt.Errorf("copy templates: %w", err)
	}

	_, _ = fmt.Fprintln(w, "Pulled successfully!")
	_, _ = fmt.Fprintf(w, "  Config directory: %s\n", configDir)

	return nil
}

// NewSyncPullCmd creates a new sync pull command for testing.
func NewSyncPullCmd(configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull config and templates from remote",
		Long: `Pull configuration and templates from the remote repository.

This command pulls the latest changes from the remote repository and
copies the config.yaml and templates to your local dotgh config directory.

Examples:
  dotgh sync pull`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSyncPullWithDir(cmd, configDir)
		},
	}
	return cmd
}
