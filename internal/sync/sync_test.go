package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("creates manager with config dir", func(t *testing.T) {
		configDir := "/tmp/test-config"
		m := NewManager(configDir)
		assert.Equal(t, configDir, m.configDir)
	})
}

func TestSyncDirPath(t *testing.T) {
	t.Run("returns sync directory path", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)
		expected := filepath.Join(tmpDir, ".sync")
		assert.Equal(t, expected, m.SyncDirPath())
	})
}

func TestIsInitialized(t *testing.T) {
	t.Run("returns false when sync dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)
		assert.False(t, m.IsInitialized())
	})

	t.Run("returns false when sync dir exists but is not git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		m := NewManager(tmpDir)
		assert.False(t, m.IsInitialized())
	})

	t.Run("returns true when sync dir is a git repo", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(filepath.Join(syncDir, ".git"), 0755))

		m := NewManager(tmpDir)
		assert.True(t, m.IsInitialized())
	})
}

func TestInitialize(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("initializes with empty repository", func(t *testing.T) {
		// Create bare repo as remote
		bareDir := t.TempDir()
		bareCmd := exec.Command("git", "init", "--bare", "--initial-branch=main")
		bareCmd.Dir = bareDir
		require.NoError(t, bareCmd.Run())

		// Create config dir and initialize
		configDir := t.TempDir()
		m := NewManager(configDir)

		err := m.Initialize(bareDir, "main")
		require.NoError(t, err)

		// Verify sync directory was created
		syncDir := m.SyncDirPath()
		_, err = os.Stat(syncDir)
		require.NoError(t, err)

		// Verify it's a git repo
		assert.True(t, m.IsInitialized())

		// Verify README was created
		_, err = os.Stat(filepath.Join(syncDir, "README.md"))
		require.NoError(t, err)

		// Verify remote was added
		client := m.GetGitClient()
		url, err := client.RemoteGetURL("origin")
		require.NoError(t, err)
		assert.Equal(t, bareDir, url)

		// Verify correct branch
		branch, err := client.GetCurrentBranch()
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("clones existing repository with content", func(t *testing.T) {
		// Create source repo with content
		srcDir := t.TempDir()
		setupGitRepo(t, srcDir)

		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "config.yaml"), []byte("test: value\n"), 0644))
		require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "templates", "test"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "templates", "test", "AGENTS.md"), []byte("# Test"), 0644))

		// Commit files
		cmd := exec.Command("git", "add", ".")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "add files")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		// Create config dir and initialize from source
		configDir := t.TempDir()
		m := NewManager(configDir)

		err := m.Initialize(srcDir, "")
		require.NoError(t, err)

		// Verify files were cloned
		syncDir := m.SyncDirPath()
		content, err := os.ReadFile(filepath.Join(syncDir, "config.yaml"))
		require.NoError(t, err)
		assert.Equal(t, "test: value\n", string(content))

		_, err = os.Stat(filepath.Join(syncDir, "templates", "test", "AGENTS.md"))
		require.NoError(t, err)
	})

	t.Run("uses default branch main when not specified", func(t *testing.T) {
		// Create a source repo with content so we can test branch
		srcDir := t.TempDir()
		srcCmd := exec.Command("git", "init", "--initial-branch=main")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		srcCmd = exec.Command("git", "config", "user.email", "test@test.com")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		srcCmd = exec.Command("git", "config", "user.name", "Test")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		// Create a file and commit
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("test"), 0644))
		srcCmd = exec.Command("git", "add", ".")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())
		srcCmd = exec.Command("git", "commit", "-m", "initial")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		configDir := t.TempDir()
		m := NewManager(configDir)

		err := m.Initialize(srcDir, "")
		require.NoError(t, err)

		branch, err := m.GetGitClient().GetCurrentBranch()
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("uses specified branch", func(t *testing.T) {
		// Create a source repo with content on develop branch
		srcDir := t.TempDir()
		srcCmd := exec.Command("git", "init", "--initial-branch=develop")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		srcCmd = exec.Command("git", "config", "user.email", "test@test.com")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		srcCmd = exec.Command("git", "config", "user.name", "Test")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		// Create a file and commit
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "test.txt"), []byte("test"), 0644))
		srcCmd = exec.Command("git", "add", ".")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())
		srcCmd = exec.Command("git", "commit", "-m", "initial")
		srcCmd.Dir = srcDir
		require.NoError(t, srcCmd.Run())

		configDir := t.TempDir()
		m := NewManager(configDir)

		err := m.Initialize(srcDir, "develop")
		require.NoError(t, err)

		branch, err := m.GetGitClient().GetCurrentBranch()
		require.NoError(t, err)
		assert.Equal(t, "develop", branch)
	})
}

func TestStageAndCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("stages and commits changes", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Initialize git repo
		setupGitRepo(t, syncDir)

		// Create initial commit
		readmePath := filepath.Join(syncDir, "README.md")
		require.NoError(t, os.WriteFile(readmePath, []byte("# Test"), 0644))

		cmd := exec.Command("git", "add", ".")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		// Add a new file
		require.NoError(t, os.WriteFile(filepath.Join(syncDir, "config.yaml"), []byte("test"), 0644))

		m := NewManager(tmpDir)
		err := m.StageAndCommit("test commit")
		require.NoError(t, err)

		// Verify commit was created
		cmd = exec.Command("git", "log", "--oneline")
		cmd.Dir = syncDir
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "test commit")
	})
}

func TestGetSyncStatus(t *testing.T) {
	t.Run("returns not initialized when sync dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		status, err := m.GetSyncStatus()
		require.NoError(t, err)
		assert.Equal(t, StatusNotInitialized, status.State)
	})

	t.Run("returns clean status for initialized repo without changes", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Initialize git repo
		setupGitRepo(t, syncDir)

		// Create initial commit
		readmePath := filepath.Join(syncDir, "README.md")
		require.NoError(t, os.WriteFile(readmePath, []byte("# Test"), 0644))

		cmd := exec.Command("git", "add", ".")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		// Add remote
		cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		m := NewManager(tmpDir)
		status, err := m.GetSyncStatus()
		require.NoError(t, err)

		assert.Equal(t, StatusClean, status.State)
		assert.Equal(t, "https://github.com/test/repo.git", status.RepoURL)
		assert.False(t, status.HasChanges)
	})

	t.Run("returns dirty status when there are uncommitted changes", func(t *testing.T) {
		if testing.Short() {
			t.Skip("skipping integration test in short mode")
		}

		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Initialize git repo
		setupGitRepo(t, syncDir)

		// Create initial commit
		readmePath := filepath.Join(syncDir, "README.md")
		require.NoError(t, os.WriteFile(readmePath, []byte("# Test"), 0644))

		cmd := exec.Command("git", "add", ".")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = syncDir
		require.NoError(t, cmd.Run())

		// Add uncommitted file
		require.NoError(t, os.WriteFile(filepath.Join(syncDir, "new-file.txt"), []byte("new"), 0644))

		m := NewManager(tmpDir)
		status, err := m.GetSyncStatus()
		require.NoError(t, err)

		assert.Equal(t, StatusDirty, status.State)
		assert.True(t, status.HasChanges)
		assert.Contains(t, status.Changes, "new-file.txt")
	})
}

func TestCopyConfigToSync(t *testing.T) {
	t.Run("copies config file to sync directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Create config file
		configContent := "includes:\n  - AGENTS.md\n"
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(configContent), 0644))

		m := NewManager(tmpDir)
		err := m.CopyConfigToSync()
		require.NoError(t, err)

		// Verify config was copied
		syncedConfig := filepath.Join(syncDir, "config.yaml")
		content, err := os.ReadFile(syncedConfig)
		require.NoError(t, err)
		assert.Equal(t, configContent, string(content))
	})

	t.Run("skips when config file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		m := NewManager(tmpDir)
		err := m.CopyConfigToSync()
		require.NoError(t, err)
	})
}

func TestCopyTemplatesToSync(t *testing.T) {
	t.Run("copies templates directory to sync", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Create templates directory with files
		templatesDir := filepath.Join(tmpDir, "templates")
		require.NoError(t, os.MkdirAll(filepath.Join(templatesDir, "project1"), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "project1", "AGENTS.md"), []byte("# Agent"), 0644))

		m := NewManager(tmpDir)
		err := m.CopyTemplatesToSync()
		require.NoError(t, err)

		// Verify templates were copied
		syncedFile := filepath.Join(syncDir, "templates", "project1", "AGENTS.md")
		content, err := os.ReadFile(syncedFile)
		require.NoError(t, err)
		assert.Equal(t, "# Agent", string(content))
	})

	t.Run("skips when templates directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		m := NewManager(tmpDir)
		err := m.CopyTemplatesToSync()
		require.NoError(t, err)
	})
}

func TestCopyConfigFromSync(t *testing.T) {
	t.Run("copies config from sync to config dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Create synced config file
		configContent := "includes:\n  - AGENTS.md\n"
		require.NoError(t, os.WriteFile(filepath.Join(syncDir, "config.yaml"), []byte(configContent), 0644))

		m := NewManager(tmpDir)
		err := m.CopyConfigFromSync()
		require.NoError(t, err)

		// Verify config was copied
		localConfig := filepath.Join(tmpDir, "config.yaml")
		content, err := os.ReadFile(localConfig)
		require.NoError(t, err)
		assert.Equal(t, configContent, string(content))
	})

	t.Run("skips when synced config does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		m := NewManager(tmpDir)
		err := m.CopyConfigFromSync()
		require.NoError(t, err)
	})
}

func TestCopyTemplatesFromSync(t *testing.T) {
	t.Run("copies templates from sync to config dir", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		syncTemplates := filepath.Join(syncDir, "templates", "project1")
		require.NoError(t, os.MkdirAll(syncTemplates, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(syncTemplates, "AGENTS.md"), []byte("# Agent"), 0644))

		m := NewManager(tmpDir)
		err := m.CopyTemplatesFromSync()
		require.NoError(t, err)

		// Verify templates were copied
		localFile := filepath.Join(tmpDir, "templates", "project1", "AGENTS.md")
		content, err := os.ReadFile(localFile)
		require.NoError(t, err)
		assert.Equal(t, "# Agent", string(content))
	})

	t.Run("skips when synced templates do not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		syncDir := filepath.Join(tmpDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		m := NewManager(tmpDir)
		err := m.CopyTemplatesFromSync()
		require.NoError(t, err)
	})
}

// setupGitRepo initializes a git repository with user config
func setupGitRepo(t *testing.T, dir string) {
	t.Helper()

	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}
