package external

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// FZFOptions configures FZF behavior
type FZFOptions struct {
	Header        string   // Header text
	Prompt        string   // Prompt text
	Preview       string   // Preview command
	PreviewWindow string   // Preview window configuration
	Height        string   // Height (e.g., "40%")
	Multi         bool     // Allow multiple selections
	NoSort        bool     // Disable sorting
	Reverse       bool     // Reverse layout
	ExtraArgs     []string // Additional FZF arguments
}

// Select runs FZF with the given items and options
// Returns the selected item(s) and any error
func Select(items []string, opts FZFOptions) ([]string, error) {
	// Check if fzf is available
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, fmt.Errorf("fzf not found: %w", err)
	}

	// Build FZF arguments
	args := []string{}

	if opts.Header != "" {
		args = append(args, "--header", opts.Header)
	}

	if opts.Prompt != "" {
		args = append(args, "--prompt", opts.Prompt)
	}

	if opts.Preview != "" {
		args = append(args, "--preview", opts.Preview)
	}

	if opts.PreviewWindow != "" {
		args = append(args, "--preview-window", opts.PreviewWindow)
	}

	if opts.Height != "" {
		args = append(args, "--height", opts.Height)
	}

	if opts.Multi {
		args = append(args, "--multi")
	}

	if opts.NoSort {
		args = append(args, "--no-sort")
	}

	if opts.Reverse {
		args = append(args, "--reverse")
	}

	args = append(args, opts.ExtraArgs...)

	// Prepare input
	input := strings.Join(items, "\n")

	// Run FZF
	cmd := exec.Command("fzf", args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Check for cancellation (exit code 130)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 130 {
				// User cancelled (Ctrl-C or Esc)
				return nil, fmt.Errorf("cancelled")
			}
		}
		return nil, fmt.Errorf("fzf failed: %w: %s", err, stderr.String())
	}

	// Parse output
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return []string{}, nil
	}

	// Split by newline for multi-select
	selected := strings.Split(output, "\n")
	return selected, nil
}

// SelectOne runs FZF and returns a single selected item
// Returns error if cancelled or if no item selected
func SelectOne(items []string, opts FZFOptions) (string, error) {
	// Ensure multi is false
	opts.Multi = false

	selected, err := Select(items, opts)
	if err != nil {
		return "", err
	}

	if len(selected) == 0 {
		return "", fmt.Errorf("no item selected")
	}

	return selected[0], nil
}

// SelectWithDefault runs FZF with a default selection (first item)
func SelectWithDefault(items []string, opts FZFOptions) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items to select from")
	}

	return SelectOne(items, opts)
}

// IsFZFAvailable checks if FZF is installed
func IsFZFAvailable() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}
