// Package prompt provides user confirmation prompts for dotgh.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Confirm asks the user for confirmation with the given message.
// If defaultNo is true, the default answer is "no" (pressing Enter = no).
// If defaultNo is false, the default answer is "yes" (pressing Enter = yes).
// Returns true if user confirms, false otherwise.
func Confirm(message string, defaultNo bool, w io.Writer, r io.Reader) (bool, error) {
	var prompt string
	if defaultNo {
		prompt = fmt.Sprintf("%s [y/N]: ", message)
	} else {
		prompt = fmt.Sprintf("%s [Y/n]: ", message)
	}

	_, err := fmt.Fprint(w, prompt)
	if err != nil {
		return false, fmt.Errorf("write prompt: %w", err)
	}

	reader := bufio.NewReader(r)
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read input: %w", err)
	}
	// Note: if err == io.EOF, input contains whatever was read before EOF

	input = strings.TrimSpace(strings.ToLower(input))

	switch input {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	case "":
		// Empty input - use default
		return !defaultNo, nil
	default:
		// Invalid input - treat as no
		return false, nil
	}
}

// ConfirmWithDefault is a convenience function that always defaults to "no".
func ConfirmWithDefault(message string, w io.Writer, r io.Reader) (bool, error) {
	return Confirm(message, true, w, r)
}
