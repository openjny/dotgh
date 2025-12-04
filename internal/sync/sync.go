// Package sync provides synchronization functionality for dotgh config and templates.
package sync

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/openjny/dotgh/internal/git"
)

const (
	// SyncDirName is the name of the sync directory.
	SyncDirName = ".sync"
)

// SyncState represents the current sync state.
type SyncState string

const (
	// StatusNotInitialized indicates sync has not been set up.
	StatusNotInitialized SyncState = "not_initialized"
	// StatusClean indicates no local changes.
	StatusClean SyncState = "clean"
	// StatusDirty indicates there are uncommitted local changes.
	StatusDirty SyncState = "dirty"
)

// SyncStatus represents the synchronization status.
type SyncStatus struct {
	State      SyncState
	RepoURL    string
	Branch     string
	HasChanges bool
	Changes    []string
}

// Manager handles sync operations.
type Manager struct {
	configDir string
	git       *git.Client
}

// NewManager creates a new sync manager.
func NewManager(configDir string) *Manager {
	syncDir := filepath.Join(configDir, SyncDirName)
	return &Manager{
		configDir: configDir,
		git:       git.New(syncDir),
	}
}

// SyncDirPath returns the path to the sync directory.
func (m *Manager) SyncDirPath() string {
	return filepath.Join(m.configDir, SyncDirName)
}

// IsInitialized returns true if sync has been initialized.
func (m *Manager) IsInitialized() bool {
	return m.git.IsRepo()
}

// Initialize sets up the sync directory with the given repository.
func (m *Manager) Initialize(repoURL, branch string) error {
	syncDir := m.SyncDirPath()

	// Create sync directory
	if err := os.MkdirAll(syncDir, 0755); err != nil {
		return fmt.Errorf("create sync directory: %w", err)
	}

	// Try to clone the repository
	err := m.git.Clone(repoURL, branch)
	if err != nil {
		// Only initialize new repo if the remote is empty
		// For other errors (auth, network, etc.), propagate them
		if !errors.Is(err, git.ErrEmptyRepository) {
			return fmt.Errorf("clone repository: %w", err)
		}

		// If clone fails due to empty repo, initialize a new repo
		if initErr := m.git.Init(); initErr != nil {
			return fmt.Errorf("init git repo: %w", initErr)
		}

		// Add remote
		if remoteErr := m.git.RemoteAdd("origin", repoURL); remoteErr != nil {
			return fmt.Errorf("add remote: %w", remoteErr)
		}

		// Create and checkout the branch
		// First create an initial commit so we can create branch
		readmePath := filepath.Join(syncDir, "README.md")
		if writeErr := os.WriteFile(readmePath, []byte("# dotgh sync\n\nThis repository stores dotgh configuration and templates.\n"), 0644); writeErr != nil {
			return fmt.Errorf("write readme: %w", writeErr)
		}

		if addErr := m.git.Add("."); addErr != nil {
			return fmt.Errorf("git add: %w", addErr)
		}

		if commitErr := m.git.Commit("Initial commit"); commitErr != nil {
			return fmt.Errorf("initial commit: %w", commitErr)
		}

		// Create branch if not main/master
		if branch != "" && branch != "main" && branch != "master" {
			if branchErr := m.git.CheckoutBranch(branch, true); branchErr != nil {
				return fmt.Errorf("create branch: %w", branchErr)
			}
		}
	}

	return nil
}

// CopyConfigToSync copies the config file to the sync directory.
func (m *Manager) CopyConfigToSync() error {
	srcPath := filepath.Join(m.configDir, "config.yaml")
	dstPath := filepath.Join(m.SyncDirPath(), "config.yaml")
	return copyFileIfExists(srcPath, dstPath)
}

// CopyTemplatesToSync copies the templates directory to the sync directory.
func (m *Manager) CopyTemplatesToSync() error {
	srcDir := filepath.Join(m.configDir, "templates")
	dstDir := filepath.Join(m.SyncDirPath(), "templates")
	return copyDirIfExists(srcDir, dstDir)
}

// CopyConfigFromSync copies the config file from the sync directory.
func (m *Manager) CopyConfigFromSync() error {
	srcPath := filepath.Join(m.SyncDirPath(), "config.yaml")
	dstPath := filepath.Join(m.configDir, "config.yaml")
	return copyFileIfExists(srcPath, dstPath)
}

// CopyTemplatesFromSync copies the templates from the sync directory.
func (m *Manager) CopyTemplatesFromSync() error {
	srcDir := filepath.Join(m.SyncDirPath(), "templates")
	dstDir := filepath.Join(m.configDir, "templates")
	return copyDirIfExists(srcDir, dstDir)
}

// GetSyncStatus returns the current sync status.
func (m *Manager) GetSyncStatus() (*SyncStatus, error) {
	status := &SyncStatus{}

	if !m.IsInitialized() {
		status.State = StatusNotInitialized
		return status, nil
	}

	// Get repository info
	if url, err := m.git.RemoteGetURL("origin"); err == nil {
		status.RepoURL = url
	}

	if branch, err := m.git.GetCurrentBranch(); err == nil {
		status.Branch = branch
	}

	// Check for changes
	gitStatus, err := m.git.Status()
	if err != nil {
		return nil, fmt.Errorf("get git status: %w", err)
	}

	if gitStatus.IsClean() {
		status.State = StatusClean
	} else {
		status.State = StatusDirty
		status.HasChanges = true
		status.Changes = append(status.Changes, gitStatus.Added...)
		status.Changes = append(status.Changes, gitStatus.Modified...)
		status.Changes = append(status.Changes, gitStatus.Deleted...)
		status.Changes = append(status.Changes, gitStatus.Untracked...)
	}

	return status, nil
}

// StageAndCommit stages all changes and creates a commit.
func (m *Manager) StageAndCommit(message string) error {
	if err := m.git.Add("."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	if err := m.git.Commit(message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// Push pushes changes to the remote repository.
func (m *Manager) Push() error {
	branch, err := m.git.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("get current branch: %w", err)
	}

	// Try normal push first, fall back to push with upstream
	if err := m.git.Push(); err != nil {
		if upstreamErr := m.git.PushWithUpstream("origin", branch); upstreamErr != nil {
			return fmt.Errorf("git push: %w", upstreamErr)
		}
	}

	return nil
}

// Pull pulls changes from the remote repository.
func (m *Manager) Pull() error {
	return m.git.Pull()
}

// GetGitClient returns the underlying git client.
func (m *Manager) GetGitClient() *git.Client {
	return m.git
}

// copyFileIfExists copies a file if it exists.
func copyFileIfExists(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read source: %w", err)
	}

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Write destination file
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("write destination: %w", err)
	}

	return nil
}

// copyDirIfExists copies a directory recursively if it exists.
func copyDirIfExists(src, dst string) error {
	// Check if source exists
	srcInfo, err := os.Stat(src)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}
	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy content: %w", err)
	}

	return nil
}
