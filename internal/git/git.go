// Package git provides Git operations wrapper for dotgh sync functionality.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client represents a Git client for a specific directory.
type Client struct {
	dir string
}

// Status represents the status of a Git repository.
type Status struct {
	Added     []string
	Modified  []string
	Deleted   []string
	Untracked []string
}

// IsClean returns true if there are no uncommitted changes.
func (s *Status) IsClean() bool {
	return len(s.Added) == 0 && len(s.Modified) == 0 && len(s.Deleted) == 0 && len(s.Untracked) == 0
}

// New creates a new Git client for the specified directory.
func New(dir string) *Client {
	return &Client{dir: dir}
}

// IsGitInstalled checks if git is available in the PATH.
func IsGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// IsRepo returns true if the directory is a Git repository.
func (c *Client) IsRepo() bool {
	gitDir := filepath.Join(c.dir, ".git")
	_, err := os.Stat(gitDir)
	return err == nil
}

// Init initializes a new Git repository.
func (c *Client) Init() error {
	return c.run("init")
}

// Clone clones a repository to the client's directory.
func (c *Client) Clone(repo, branch string) error {
	// Clone into current directory
	args := []string{"clone"}
	if branch != "" {
		args = append(args, "-b", branch)
	}
	args = append(args, repo, ".")
	return c.run(args...)
}

// Add stages files for commit.
func (c *Client) Add(paths ...string) error {
	args := append([]string{"add"}, paths...)
	return c.run(args...)
}

// Commit creates a commit with the given message.
func (c *Client) Commit(message string) error {
	return c.run("commit", "-m", message)
}

// Push pushes commits to the remote repository.
func (c *Client) Push() error {
	return c.run("push")
}

// PushWithUpstream pushes commits and sets upstream branch.
func (c *Client) PushWithUpstream(remote, branch string) error {
	return c.run("push", "-u", remote, branch)
}

// Pull pulls changes from the remote repository.
func (c *Client) Pull() error {
	return c.run("pull")
}

// RemoteAdd adds a remote repository.
func (c *Client) RemoteAdd(name, url string) error {
	return c.run("remote", "add", name, url)
}

// RemoteGetURL gets the URL of a remote repository.
func (c *Client) RemoteGetURL(name string) (string, error) {
	output, err := c.runOutput("remote", "get-url", name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// HasRemote checks if a remote with the given name exists.
func (c *Client) HasRemote(name string) bool {
	_, err := c.RemoteGetURL(name)
	return err == nil
}

// Status returns the status of the repository.
func (c *Client) Status() (*Status, error) {
	output, err := c.runOutput("status", "--porcelain")
	if err != nil {
		return nil, err
	}

	status := &Status{}
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}
		indicator := line[:2]
		filename := strings.TrimSpace(line[3:])

		switch {
		case strings.Contains(indicator, "A"):
			status.Added = append(status.Added, filename)
		case strings.Contains(indicator, "M"):
			status.Modified = append(status.Modified, filename)
		case strings.Contains(indicator, "D"):
			status.Deleted = append(status.Deleted, filename)
		case strings.HasPrefix(indicator, "??"):
			status.Untracked = append(status.Untracked, filename)
		}
	}

	return status, nil
}

// GetCurrentBranch returns the current branch name.
func (c *Client) GetCurrentBranch() (string, error) {
	output, err := c.runOutput("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// CheckoutBranch switches to or creates a branch.
func (c *Client) CheckoutBranch(branch string, create bool) error {
	if create {
		return c.run("checkout", "-b", branch)
	}
	return c.run("checkout", branch)
}

// Fetch fetches changes from remote.
func (c *Client) Fetch() error {
	return c.run("fetch")
}

// GetDir returns the directory of the git client.
func (c *Client) GetDir() string {
	return c.dir
}

// run executes a git command in the client's directory.
func (c *Client) run(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.dir
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// runOutput executes a git command and returns its output.
func (c *Client) runOutput(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.dir
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), string(exitErr.Stderr))
		}
		return "", err
	}
	return string(output), nil
}
