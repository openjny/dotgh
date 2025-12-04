package prompt

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		defaultNo bool
		input     string
		want      bool
		wantErr   bool
	}{
		{
			name:      "yes input with defaultNo=true",
			message:   "Continue?",
			defaultNo: true,
			input:     "y\n",
			want:      true,
		},
		{
			name:      "yes full word with defaultNo=true",
			message:   "Continue?",
			defaultNo: true,
			input:     "yes\n",
			want:      true,
		},
		{
			name:      "no input with defaultNo=true",
			message:   "Continue?",
			defaultNo: true,
			input:     "n\n",
			want:      false,
		},
		{
			name:      "no full word with defaultNo=true",
			message:   "Continue?",
			defaultNo: true,
			input:     "no\n",
			want:      false,
		},
		{
			name:      "empty input with defaultNo=true (default to no)",
			message:   "Continue?",
			defaultNo: true,
			input:     "\n",
			want:      false,
		},
		{
			name:      "empty input with defaultNo=false (default to yes)",
			message:   "Continue?",
			defaultNo: false,
			input:     "\n",
			want:      true,
		},
		{
			name:      "uppercase Y",
			message:   "Continue?",
			defaultNo: true,
			input:     "Y\n",
			want:      true,
		},
		{
			name:      "uppercase YES",
			message:   "Continue?",
			defaultNo: true,
			input:     "YES\n",
			want:      true,
		},
		{
			name:      "uppercase N",
			message:   "Continue?",
			defaultNo: true,
			input:     "N\n",
			want:      false,
		},
		{
			name:      "invalid input treated as no",
			message:   "Continue?",
			defaultNo: true,
			input:     "maybe\n",
			want:      false,
		},
		{
			name:      "invalid input treated as no even with defaultNo=false",
			message:   "Continue?",
			defaultNo: false,
			input:     "invalid\n",
			want:      false,
		},
		{
			name:      "input with spaces",
			message:   "Continue?",
			defaultNo: true,
			input:     "  y  \n",
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader(tt.input)

			got, err := Confirm(tt.message, tt.defaultNo, &out, in)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfirm_PromptFormat(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		defaultNo  bool
		wantPrompt string
	}{
		{
			name:       "defaultNo=true shows [y/N]",
			message:    "Delete files?",
			defaultNo:  true,
			wantPrompt: "Delete files? [y/N]: ",
		},
		{
			name:       "defaultNo=false shows [Y/n]",
			message:    "Continue?",
			defaultNo:  false,
			wantPrompt: "Continue? [Y/n]: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			in := strings.NewReader("y\n")

			_, err := Confirm(tt.message, tt.defaultNo, &out, in)
			require.NoError(t, err)

			assert.Equal(t, tt.wantPrompt, out.String())
		})
	}
}

func TestConfirm_EOFWithoutNewline(t *testing.T) {
	var out bytes.Buffer
	// Input without newline (EOF)
	in := strings.NewReader("y")

	got, err := Confirm("Continue?", true, &out, in)
	require.NoError(t, err)
	assert.True(t, got)
}

func TestConfirm_EmptyEOF(t *testing.T) {
	var out bytes.Buffer
	// Empty input (EOF immediately)
	in := strings.NewReader("")

	got, err := Confirm("Continue?", true, &out, in)
	require.NoError(t, err)
	// Empty input with defaultNo=true should return false
	assert.False(t, got)
}

func TestConfirmWithDefault(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("\n")

	// ConfirmWithDefault always defaults to no
	got, err := ConfirmWithDefault("Continue?", &out, in)
	require.NoError(t, err)
	assert.False(t, got)
	assert.Contains(t, out.String(), "[y/N]")
}
