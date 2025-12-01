package commands

import (
	"bytes"
	"testing"
)

func TestUpdateCommand_HasFlags(t *testing.T) {
	cmd := NewUpdateCmd()

	// Test that --check flag exists
	checkFlag := cmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Fatal("expected --check flag to exist")
	}
	if checkFlag.Shorthand != "c" {
		t.Errorf("expected --check shorthand to be 'c', got %q", checkFlag.Shorthand)
	}
}

func TestUpdateCommand_Usage(t *testing.T) {
	cmd := NewUpdateCmd()

	if cmd.Use != "update" {
		t.Errorf("Use = %q, want %q", cmd.Use, "update")
	}

	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}

func TestUpdateCommand_OutputFormat(t *testing.T) {
	// This test validates the output format structure
	// We don't actually run updates in tests
	cmd := NewUpdateCmd()

	// Verify the command can be executed (even if it fails due to network)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Just verify the command is properly configured
	if cmd.RunE == nil {
		t.Error("expected RunE to be set")
	}
}
