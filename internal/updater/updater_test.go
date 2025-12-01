package updater

import (
	"context"
	"testing"
)

func TestNewUpdater(t *testing.T) {
	u := New("openjny", "dotgh")

	if u.Owner != "openjny" {
		t.Errorf("Owner = %q, want %q", u.Owner, "openjny")
	}
	if u.Repo != "dotgh" {
		t.Errorf("Repo = %q, want %q", u.Repo, "dotgh")
	}
}

func TestUpdater_IsUpdateAvailable(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		want           bool
	}{
		{
			name:           "newer version available",
			currentVersion: "1.0.0",
			latestVersion:  "1.1.0",
			want:           true,
		},
		{
			name:           "same version",
			currentVersion: "1.0.0",
			latestVersion:  "1.0.0",
			want:           false,
		},
		{
			name:           "older version (current is newer)",
			currentVersion: "2.0.0",
			latestVersion:  "1.0.0",
			want:           false,
		},
		{
			name:           "dev version skips update",
			currentVersion: "dev",
			latestVersion:  "1.0.0",
			want:           false,
		},
		{
			name:           "with v prefix in current",
			currentVersion: "v1.0.0",
			latestVersion:  "1.1.0",
			want:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUpdateAvailable(tt.currentVersion, tt.latestVersion)
			if got != tt.want {
				t.Errorf("isUpdateAvailable(%q, %q) = %v, want %v",
					tt.currentVersion, tt.latestVersion, got, tt.want)
			}
		})
	}
}

func TestUpdater_CheckForUpdate_Cancelled(t *testing.T) {
	u := New("openjny", "dotgh")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := u.CheckForUpdate(ctx, "1.0.0")
	if err == nil {
		t.Error("expected error for cancelled context, got nil")
	}
}

func TestReleaseInfo(t *testing.T) {
	info := &ReleaseInfo{
		Version:      "1.2.3",
		ReleaseNotes: "Test release notes",
		URL:          "https://github.com/openjny/dotgh/releases/tag/v1.2.3",
	}

	if info.Version != "1.2.3" {
		t.Errorf("Version = %q, want %q", info.Version, "1.2.3")
	}
	if info.ReleaseNotes != "Test release notes" {
		t.Errorf("ReleaseNotes = %q, want %q", info.ReleaseNotes, "Test release notes")
	}
	if info.URL != "https://github.com/openjny/dotgh/releases/tag/v1.2.3" {
		t.Errorf("URL = %q, want %q", info.URL, "https://github.com/openjny/dotgh/releases/tag/v1.2.3")
	}
}
