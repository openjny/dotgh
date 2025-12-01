package commands

import (
	"testing"
)

func TestExecute(t *testing.T) {
	// Test that Execute returns without error for help
	rootCmd.SetArgs([]string{"--help"})
	if err := Execute(); err != nil {
		t.Errorf("Execute() with --help returned error: %v", err)
	}
}

func TestRootCmdHasListSubcommand(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("root command should have 'list' subcommand")
	}
}
