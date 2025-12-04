package commands

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCommand(t *testing.T) {
	t.Run("has subcommands", func(t *testing.T) {
		cmd := NewSyncCmd("")
		assert.NotNil(t, cmd)
		assert.Equal(t, "sync", cmd.Use)

		// Check subcommands
		subCmds := cmd.Commands()
		var names []string
		for _, c := range subCmds {
			names = append(names, c.Name())
		}
		assert.Contains(t, names, "init")
		assert.Contains(t, names, "push")
		assert.Contains(t, names, "pull")
		assert.Contains(t, names, "status")
	})
}

func TestSyncInitCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("initializes sync with new repo", func(t *testing.T) {
		// Create a bare git repo as remote
		bareDir := t.TempDir()
		bareCmd := exec.Command("git", "init", "--bare")
		bareCmd.Dir = bareDir
		require.NoError(t, bareCmd.Run())

		// Create config dir
		configDir := t.TempDir()

		// Run sync init
		cmd := NewSyncInitCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{bareDir})

		err := cmd.Execute()
		require.NoError(t, err)

		// Verify .sync directory was created
		syncDir := filepath.Join(configDir, ".sync")
		_, err = os.Stat(syncDir)
		require.NoError(t, err)

		// Verify it's a git repo
		gitDir := filepath.Join(syncDir, ".git")
		_, err = os.Stat(gitDir)
		require.NoError(t, err)

		// Verify output
		assert.Contains(t, buf.String(), "Sync initialized")
	})

	t.Run("fails when already initialized", func(t *testing.T) {
		configDir := t.TempDir()

		// Create existing sync dir with git
		syncDir := filepath.Join(configDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))
		initCmd := exec.Command("git", "init")
		initCmd.Dir = syncDir
		require.NoError(t, initCmd.Run())

		cmd := NewSyncInitCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs([]string{"https://github.com/test/repo.git"})

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already initialized")
	})
}

func TestSyncStatusCommand(t *testing.T) {
	t.Run("shows not initialized when sync not set up", func(t *testing.T) {
		configDir := t.TempDir()

		cmd := NewSyncStatusCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)

		err := cmd.Execute()
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "not initialized")
	})

	t.Run("shows clean status when no changes", func(t *testing.T) {
		configDir := t.TempDir()
		syncDir := filepath.Join(configDir, ".sync")
		require.NoError(t, os.MkdirAll(syncDir, 0755))

		// Initialize git repo
		initCmd := exec.Command("git", "init")
		initCmd.Dir = syncDir
		require.NoError(t, initCmd.Run())

		configGitCmd := exec.Command("git", "config", "user.email", "test@test.com")
		configGitCmd.Dir = syncDir
		require.NoError(t, configGitCmd.Run())

		configGitCmd = exec.Command("git", "config", "user.name", "Test")
		configGitCmd.Dir = syncDir
		require.NoError(t, configGitCmd.Run())

		// Create and commit a file
		require.NoError(t, os.WriteFile(filepath.Join(syncDir, "test.txt"), []byte("hello"), 0644))
		addCmd := exec.Command("git", "add", ".")
		addCmd.Dir = syncDir
		require.NoError(t, addCmd.Run())
		commitCmd := exec.Command("git", "commit", "-m", "initial")
		commitCmd.Dir = syncDir
		require.NoError(t, commitCmd.Run())

		cmd := NewSyncStatusCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)

		err := cmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Status:")
		assert.Contains(t, output, "clean")
	})
}

func TestSyncPushCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("pushes config and templates", func(t *testing.T) {
		// Create bare repo as remote
		bareDir := t.TempDir()
		bareCmd := exec.Command("git", "init", "--bare")
		bareCmd.Dir = bareDir
		require.NoError(t, bareCmd.Run())

		// Create config dir with config file and templates
		configDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("includes:\n  - AGENTS.md\n"), 0644))

		templatesDir := filepath.Join(configDir, "templates", "myproject")
		require.NoError(t, os.MkdirAll(templatesDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "AGENTS.md"), []byte("# Agent"), 0644))

		// Initialize sync
		initCmd := NewSyncInitCmd(configDir)
		initCmd.SetArgs([]string{bareDir})
		var initBuf bytes.Buffer
		initCmd.SetOut(&initBuf)
		require.NoError(t, initCmd.Execute())

		// Push
		pushCmd := NewSyncPushCmd(configDir)
		var buf bytes.Buffer
		pushCmd.SetOut(&buf)
		pushCmd.SetArgs([]string{"-m", "sync config"})

		err := pushCmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Pushed")

		// Verify files were copied to sync dir
		syncDir := filepath.Join(configDir, ".sync")
		_, err = os.Stat(filepath.Join(syncDir, "config.yaml"))
		require.NoError(t, err)
		_, err = os.Stat(filepath.Join(syncDir, "templates", "myproject", "AGENTS.md"))
		require.NoError(t, err)
	})

	t.Run("fails when not initialized", func(t *testing.T) {
		configDir := t.TempDir()

		cmd := NewSyncPushCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}

func TestSyncPullCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Run("pulls config and templates", func(t *testing.T) {
		// Create bare repo as remote
		bareDir := t.TempDir()
		bareCmd := exec.Command("git", "init", "--bare")
		bareCmd.Dir = bareDir
		require.NoError(t, bareCmd.Run())

		// Create config dir 1 and push some content
		configDir1 := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(configDir1, "config.yaml"), []byte("includes:\n  - AGENTS.md\n"), 0644))

		templatesDir := filepath.Join(configDir1, "templates", "myproject")
		require.NoError(t, os.MkdirAll(templatesDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "AGENTS.md"), []byte("# Agent"), 0644))

		// Initialize sync on configDir1
		initCmd := NewSyncInitCmd(configDir1)
		initCmd.SetArgs([]string{bareDir})
		var initBuf bytes.Buffer
		initCmd.SetOut(&initBuf)
		require.NoError(t, initCmd.Execute())

		// Push from configDir1
		pushCmd := NewSyncPushCmd(configDir1)
		var pushBuf bytes.Buffer
		pushCmd.SetOut(&pushBuf)
		pushCmd.SetArgs([]string{"-m", "initial sync"})
		require.NoError(t, pushCmd.Execute())

		// Create config dir 2 and initialize sync
		configDir2 := t.TempDir()
		initCmd2 := NewSyncInitCmd(configDir2)
		initCmd2.SetArgs([]string{bareDir})
		var initBuf2 bytes.Buffer
		initCmd2.SetOut(&initBuf2)
		require.NoError(t, initCmd2.Execute())

		// Pull to configDir2
		pullCmd := NewSyncPullCmd(configDir2)
		var buf bytes.Buffer
		pullCmd.SetOut(&buf)

		err := pullCmd.Execute()
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Pulled")

		// Verify files were pulled
		content, err := os.ReadFile(filepath.Join(configDir2, "config.yaml"))
		require.NoError(t, err)
		assert.True(t, strings.Contains(string(content), "AGENTS.md"))

		_, err = os.Stat(filepath.Join(configDir2, "templates", "myproject", "AGENTS.md"))
		require.NoError(t, err)
	})

	t.Run("fails when not initialized", func(t *testing.T) {
		configDir := t.TempDir()

		cmd := NewSyncPullCmd(configDir)
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)

		err := cmd.Execute()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not initialized")
	})
}
