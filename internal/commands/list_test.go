package commands

import (
	"bytes"
	"testing"
)

func TestRunList(t *testing.T) {
	tests := []struct {
		name    string
		wantOut string
		wantErr bool
	}{
		{
			name:    "no templates",
			wantOut: "Available templates:\n  (no templates found)\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			listCmd.SetOut(&buf)

			err := runList(listCmd, []string{})
			if (err != nil) != tt.wantErr {
				t.Errorf("runList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got := buf.String(); got != tt.wantOut {
				t.Errorf("output = %q, want %q", got, tt.wantOut)
			}
		})
	}
}
