// Package updater provides self-update functionality for dotgh using GitHub releases.
package updater

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/creativeprojects/go-selfupdate"
)

// ReleaseInfo contains information about a release.
type ReleaseInfo struct {
	Version      string
	URL          string
	ReleaseNotes string
	PublishedAt  time.Time
}

// Updater handles checking for updates and applying them.
type Updater struct {
	Owner string
	Repo  string
}

// New creates a new Updater instance.
func New(owner, repo string) *Updater {
	return &Updater{
		Owner: owner,
		Repo:  repo,
	}
}

// CheckForUpdate checks if a newer version is available.
// Returns the release info, whether an update is available, and any error.
func (u *Updater) CheckForUpdate(ctx context.Context, currentVersion string) (*ReleaseInfo, bool, error) {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return nil, false, fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug(u.Owner, u.Repo))
	if err != nil {
		return nil, false, fmt.Errorf("failed to detect latest version: %w", err)
	}
	if !found {
		return nil, false, nil
	}

	available := isUpdateAvailable(currentVersion, latest.Version())
	if !available {
		return nil, false, nil
	}

	return &ReleaseInfo{
		Version:      latest.Version(),
		URL:          latest.URL,
		ReleaseNotes: latest.ReleaseNotes,
		PublishedAt:  latest.PublishedAt,
	}, true, nil
}

// Update downloads and applies the specified release.
func (u *Updater) Update(ctx context.Context, release *ReleaseInfo) error {
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("failed to create GitHub source: %w", err)
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return fmt.Errorf("failed to create updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(ctx, selfupdate.NewRepositorySlug(u.Owner, u.Repo))
	if err != nil {
		return fmt.Errorf("failed to detect latest version: %w", err)
	}
	if !found {
		return fmt.Errorf("release not found")
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	if err := updater.UpdateTo(ctx, latest, exe); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	return nil
}

// isUpdateAvailable compares the current version with the latest version.
// Returns true if the latest version is newer than the current version.
func isUpdateAvailable(currentVersion, latestVersion string) bool {
	// Skip update if running development version
	if currentVersion == "" || currentVersion == "dev" {
		return false
	}

	// Parse versions using semver
	current, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false
	}

	latest, err := semver.NewVersion(latestVersion)
	if err != nil {
		return false
	}

	return latest.GreaterThan(current)
}
