package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/openjny/dotgh/internal/updater"
	"github.com/openjny/dotgh/internal/version"
	"github.com/spf13/cobra"
)

const (
	// Repository owner and name for self-update
	repoOwner = "openjny"
	repoName  = "dotgh"
)

var (
	checkOnly bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update dotgh to the latest version",
	Long: `Update dotgh to the latest version from GitHub releases.

Use --check to only check if an update is available without installing it.`,
	RunE: runUpdate,
}

// NewUpdateCmd creates a new update command for testing.
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update dotgh to the latest version",
		Long: `Update dotgh to the latest version from GitHub releases.

Use --check to only check if an update is available without installing it.`,
		RunE: runUpdate,
	}
	cmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
	return cmd
}

func init() {
	updateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	w := cmd.OutOrStdout()
	currentVersion := version.Version

	// Check if running development version
	if currentVersion == "dev" {
		_, _ = fmt.Fprintln(w, "Running development version, skipping update check")
		return nil
	}

	_, _ = fmt.Fprintf(w, "Current version: %s\n", currentVersion)
	_, _ = fmt.Fprintln(w, "Checking for updates...")

	u := updater.New(repoOwner, repoName)
	release, available, err := u.CheckForUpdate(ctx, currentVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !available {
		_, _ = fmt.Fprintln(w, "Already up to date!")
		return nil
	}

	_, _ = fmt.Fprintf(w, "New version available: %s\n", release.Version)

	if checkOnly {
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "Run 'dotgh update' to install the update")
		return nil
	}

	_, _ = fmt.Fprintln(w, "Downloading and installing update...")

	if err := u.Update(ctx, release); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	_, _ = fmt.Fprintf(w, "Successfully updated to version %s\n", release.Version)
	return nil
}
