package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsGitInstalled(t *testing.T) {
	t.Run("git is available", func(t *testing.T) {
		// This test assumes git is installed in the test environment
		assert.True(t, IsGitInstalled())
	})
}

func TestInit(t *testing.T) {
	t.Run("initializes a new git repository", func(t *testing.T) {
		tmpDir := t.TempDir()

		client := New(tmpDir)
		err := client.Init()
		require.NoError(t, err)

		// Verify .git directory exists
		gitDir := filepath.Join(tmpDir, ".git")
		_, err = os.Stat(gitDir)
		assert.NoError(t, err)
	})
}

func TestClone(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping clone test in short mode")
	}

	t.Run("clones a local repository", func(t *testing.T) {
		// Create source repo
		srcDir := t.TempDir()
		cmd := exec.Command("git", "init")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		// Create a file and commit
		testFile := filepath.Join(srcDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = srcDir
		require.NoError(t, cmd.Run())

		// Clone to destination (use empty string for branch to use default)
		dstDir := t.TempDir()
		client := New(dstDir)
		err := client.Clone(srcDir, "")
		require.NoError(t, err)

		// Verify cloned file exists
		clonedFile := filepath.Join(dstDir, "test.txt")
		content, err := os.ReadFile(clonedFile)
		require.NoError(t, err)
		assert.Equal(t, "hello", string(content))
	})
}

func TestAddAndCommit(t *testing.T) {
	t.Run("adds and commits files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Create a file
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		client := New(tmpDir)

		// Add and commit
		err := client.Add(".")
		require.NoError(t, err)

		err = client.Commit("test commit")
		require.NoError(t, err)

		// Verify commit exists
		cmd = exec.Command("git", "log", "--oneline")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "test commit")
	})
}

func TestPushAndPull(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping push/pull test in short mode")
	}

	t.Run("pushes and pulls changes", func(t *testing.T) {
		// Create bare repo as remote
		bareDir := t.TempDir()
		cmd := exec.Command("git", "init", "--bare")
		cmd.Dir = bareDir
		require.NoError(t, cmd.Run())

		// Create local repo 1
		local1Dir := t.TempDir()
		cmd = exec.Command("git", "clone", bareDir, ".")
		cmd.Dir = local1Dir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = local1Dir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = local1Dir
		require.NoError(t, cmd.Run())

		// Create and push a file from local1
		testFile := filepath.Join(local1Dir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		client1 := New(local1Dir)
		require.NoError(t, client1.Add("."))
		require.NoError(t, client1.Commit("initial commit"))
		require.NoError(t, client1.Push())

		// Create local repo 2 and pull
		local2Dir := t.TempDir()
		cmd = exec.Command("git", "clone", bareDir, ".")
		cmd.Dir = local2Dir
		require.NoError(t, cmd.Run())

		client2 := New(local2Dir)

		// Update file in local1 and push
		require.NoError(t, os.WriteFile(testFile, []byte("updated"), 0644))
		require.NoError(t, client1.Add("."))
		require.NoError(t, client1.Commit("update commit"))
		require.NoError(t, client1.Push())

		// Pull in local2
		require.NoError(t, client2.Pull())

		// Verify updated file in local2
		content, err := os.ReadFile(filepath.Join(local2Dir, "test.txt"))
		require.NoError(t, err)
		assert.Equal(t, "updated", string(content))
	})
}

func TestRemote(t *testing.T) {
	t.Run("adds remote", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		err := client.RemoteAdd("origin", "https://github.com/test/repo.git")
		require.NoError(t, err)

		// Verify remote was added
		cmd = exec.Command("git", "remote", "-v")
		cmd.Dir = tmpDir
		output, err := cmd.Output()
		require.NoError(t, err)
		assert.Contains(t, string(output), "origin")
		assert.Contains(t, string(output), "https://github.com/test/repo.git")
	})

	t.Run("gets remote URL", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		url, err := client.RemoteGetURL("origin")
		require.NoError(t, err)
		assert.Equal(t, "https://github.com/test/repo.git", url)
	})
}

func TestStatus(t *testing.T) {
	t.Run("returns clean status for clean repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		status, err := client.Status()
		require.NoError(t, err)
		assert.True(t, status.IsClean())
	})

	t.Run("returns modified files", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Modify file
		require.NoError(t, os.WriteFile(testFile, []byte("modified"), 0644))

		client := New(tmpDir)
		status, err := client.Status()
		require.NoError(t, err)
		assert.False(t, status.IsClean())
		assert.Contains(t, status.Modified, "test.txt")
	})
}

func TestIsRepo(t *testing.T) {
	t.Run("returns true for git repo", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		assert.True(t, client.IsRepo())
	})

	t.Run("returns false for non-git directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		client := New(tmpDir)
		assert.False(t, client.IsRepo())
	})
}

func TestGetCurrentBranch(t *testing.T) {
	t.Run("returns current branch name", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Create initial commit to establish branch
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		branch, err := client.GetCurrentBranch()
		require.NoError(t, err)
		// Could be "main" or "master" depending on git config
		assert.NotEmpty(t, branch)
	})
}

func TestCheckout(t *testing.T) {
	t.Run("creates and switches to new branch", func(t *testing.T) {
		tmpDir := t.TempDir()

		cmd := exec.Command("git", "init")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.email", "test@test.com")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "config", "user.name", "Test")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		// Create initial commit
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("hello"), 0644))

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = tmpDir
		require.NoError(t, cmd.Run())

		client := New(tmpDir)
		err := client.CheckoutBranch("test-branch", true)
		require.NoError(t, err)

		branch, err := client.GetCurrentBranch()
		require.NoError(t, err)
		assert.Equal(t, "test-branch", branch)
	})
}
