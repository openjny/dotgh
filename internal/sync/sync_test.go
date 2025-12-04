package sync

import (
	"os"
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

func TestGetSyncStatus(t *testing.T) {
	t.Run("returns not initialized when sync dir does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		m := NewManager(tmpDir)

		status, err := m.GetSyncStatus()
		require.NoError(t, err)
		assert.Equal(t, StatusNotInitialized, status.State)
	})
}
