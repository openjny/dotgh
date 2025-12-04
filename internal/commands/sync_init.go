package commands

import (
	"fmt"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/git"
	"github.com/openjny/dotgh/internal/sync"
	"github.com/spf13/cobra"
)

var syncInitBranch string

var syncInitCmd = &cobra.Command{
	Use:   "init <repository>",
	Short: "Initialize sync with a Git repository",
	Long: `Initialize synchronization with a Git repository.

The repository will be cloned to store your dotgh configuration and templates.
If the repository is empty, it will be initialized with a README file.

Examples:
  dotgh sync init git@github.com:user/dotgh-sync.git
  dotgh sync init https://github.com/user/dotgh-sync.git
  dotgh sync init git@github.com:user/dotgh-sync.git --branch main`,
	Args: cobra.ExactArgs(1),
	RunE: runSyncInit,
}

func init() {
	syncInitCmd.Flags().StringVarP(&syncInitBranch, "branch", "b", "main", "Branch to use for sync")
}

func runSyncInit(cmd *cobra.Command, args []string) error {
	return runSyncInitWithDir(cmd, args, config.GetConfigDir())
}

func runSyncInitWithDir(cmd *cobra.Command, args []string, configDir string) error {
	w := cmd.OutOrStdout()
	repoURL := args[0]
	branch := syncInitBranch

	// Check if git is installed
	if !git.IsGitInstalled() {
		return fmt.Errorf("git is not installed or not in PATH")
	}

	// Create sync manager
	manager := sync.NewManager(configDir)

	// Check if already initialized
	if manager.IsInitialized() {
		return fmt.Errorf("sync is already initialized at %s", manager.SyncDirPath())
	}

	// Initialize sync
	if err := manager.Initialize(repoURL, branch); err != nil {
		return fmt.Errorf("initialize sync: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Sync initialized successfully!\n")
	_, _ = fmt.Fprintf(w, "  Repository: %s\n", repoURL)
	_, _ = fmt.Fprintf(w, "  Branch: %s\n", branch)
	_, _ = fmt.Fprintf(w, "  Sync directory: %s\n", manager.SyncDirPath())
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Next steps:")
	_, _ = fmt.Fprintln(w, "  dotgh sync push    # Push your local config and templates")
	_, _ = fmt.Fprintln(w, "  dotgh sync pull    # Pull config and templates from remote")
	_, _ = fmt.Fprintln(w, "  dotgh sync status  # Check sync status")

	return nil
}

// NewSyncInitCmd creates a new sync init command for testing.
func NewSyncInitCmd(configDir string) *cobra.Command {
	var branch string

	cmd := &cobra.Command{
		Use:   "init <repository>",
		Short: "Initialize sync with a Git repository",
		Long: `Initialize synchronization with a Git repository.

The repository will be cloned to store your dotgh configuration and templates.
If the repository is empty, it will be initialized with a README file.

Examples:
  dotgh sync init git@github.com:user/dotgh-sync.git
  dotgh sync init https://github.com/user/dotgh-sync.git
  dotgh sync init git@github.com:user/dotgh-sync.git --branch main`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Temporarily set the global variable for the function
			oldBranch := syncInitBranch
			syncInitBranch = branch
			defer func() { syncInitBranch = oldBranch }()

			return runSyncInitWithDir(cmd, args, configDir)
		},
	}

	cmd.Flags().StringVarP(&branch, "branch", "b", "main", "Branch to use for sync")
	return cmd
}
