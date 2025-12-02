package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/openjny/dotgh/internal/config"
	"github.com/openjny/dotgh/internal/editor"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage dotgh configuration",
	Long:  `View and edit the dotgh configuration file.`,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Display current configuration in YAML format",
	Long:  `Display the current configuration settings from the config file or defaults if no file exists.`,
	RunE:  runConfigList,
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in the user's preferred editor",
	Long: `Open the configuration file in the user's preferred editor.
If the config file doesn't exist, it will be created with default values first.`,
	RunE: runConfigEdit,
}

func init() {
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configEditCmd)
}

// NewConfigCmd creates a new config command for testing.
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage dotgh configuration",
		Long:  `View and edit the dotgh configuration file.`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "Display current configuration in YAML format",
	}
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Open configuration file in the user's preferred editor",
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(editCmd)
	return cmd
}

// NewConfigListCmd creates a new config list command with a custom config directory.
func NewConfigListCmd(configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Display current configuration in YAML format",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigListWithDir(cmd, configDir)
		},
	}
	return cmd
}

// NewConfigEditCmd creates a new config edit command with a custom config directory.
func NewConfigEditCmd(configDir string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Open configuration file in the user's preferred editor",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigEditWithDir(cmd, configDir)
		},
	}
	return cmd
}

func runConfigList(cmd *cobra.Command, args []string) error {
	return runConfigListWithDir(cmd, config.GetConfigDir())
}

func runConfigListWithDir(cmd *cobra.Command, configDir string) error {
	configPath := filepath.Join(configDir, "config.yaml")
	w := cmd.OutOrStdout()

	// Print config file path as comment
	fmt.Fprintf(w, "# Config file: %s\n", configPath)

	// Load config (defaults if file doesn't exist)
	cfg, err := config.LoadFromDir(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	fmt.Fprint(w, string(data))
	return nil
}

func runConfigEdit(cmd *cobra.Command, args []string) error {
	return runConfigEditWithDir(cmd, config.GetConfigDir())
}

func runConfigEditWithDir(cmd *cobra.Command, configDir string) error {
	configPath := filepath.Join(configDir, "config.yaml")

	// Ensure config file exists
	if err := ensureConfigExists(configDir); err != nil {
		return fmt.Errorf("ensure config exists: %w", err)
	}

	// Load config to get editor setting
	cfg, err := config.LoadFromDir(configDir)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Build and execute editor command
	editorArgs := buildEditorCommand(cfg.Editor, configPath)
	execCmd := exec.Command(editorArgs[0], editorArgs[1:]...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return execCmd.Run()
}

// ensureConfigExists creates the config file with defaults if it doesn't exist.
func ensureConfigExists(configDir string) error {
	configPath := filepath.Join(configDir, "config.yaml")

	// Check if config file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, nothing to do
		return nil
	}

	// Create config file with defaults
	return config.CreateDefaultConfigFile(configPath)
}

// buildEditorCommand returns the command arguments to launch the editor.
func buildEditorCommand(configEditor, target string) []string {
	editorStr := editor.Detect(configEditor)
	return editor.PrepareCommand(editorStr, target)
}
