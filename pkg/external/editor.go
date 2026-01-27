package external

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Editor represents an editor configuration
type Editor struct {
	command string
	args    []string
}

// DetectEditor detects the best available editor
// Priority: nvim > vim > $EDITOR environment variable
func DetectEditor() (*Editor, error) {
	// Try nvim first
	if _, err := exec.LookPath("nvim"); err == nil {
		return &Editor{command: "nvim"}, nil
	}

	// Try vim
	if _, err := exec.LookPath("vim"); err == nil {
		return &Editor{command: "vim"}, nil
	}

	// Try $EDITOR environment variable
	if editorEnv := os.Getenv("EDITOR"); editorEnv != "" {
		return &Editor{command: editorEnv}, nil
	}

	return nil, fmt.Errorf("no editor found (tried nvim, vim, $EDITOR)")
}

// Open opens a file in the editor
func (e *Editor) Open(filePath string) error {
	cmd := exec.Command(e.command, append(e.args, filePath)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// OpenAtLine opens a file at a specific line number
// Uses +line syntax supported by vim/nvim
func (e *Editor) OpenAtLine(filePath string, lineNum int) error {
	lineArg := fmt.Sprintf("+%d", lineNum)
	cmd := exec.Command(e.command, lineArg, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// EditTemp creates a temporary file, opens it in the editor, and returns the content
func (e *Editor) EditTemp(initialContent string) (string, error) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "brain-edit-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write initial content
	if initialContent != "" {
		if _, err := tmpFile.WriteString(initialContent); err != nil {
			tmpFile.Close()
			return "", fmt.Errorf("failed to write initial content: %w", err)
		}
	}

	tmpFile.Close()

	// Open editor
	if err := e.Open(tmpPath); err != nil {
		return "", fmt.Errorf("editor failed: %w", err)
	}

	// Read result
	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to read edited content: %w", err)
	}

	return string(content), nil
}

// OpenFile is a convenience function that detects the editor and opens a file
func OpenFile(filePath string) error {
	editor, err := DetectEditor()
	if err != nil {
		return err
	}

	return editor.Open(filePath)
}

// OpenFileAtLine is a convenience function that opens a file at a specific line
func OpenFileAtLine(filePath string, lineNum int) error {
	editor, err := DetectEditor()
	if err != nil {
		return err
	}

	return editor.OpenAtLine(filePath, lineNum)
}

// EditTempFile is a convenience function for editing temporary content
func EditTempFile(initialContent string) (string, error) {
	editor, err := DetectEditor()
	if err != nil {
		return "", err
	}

	return editor.EditTemp(initialContent)
}

// OpenFileAtLineFromString opens a file at a line number specified as a string
func OpenFileAtLineFromString(filePath string, lineNumStr string) error {
	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		return fmt.Errorf("invalid line number: %s", lineNumStr)
	}

	return OpenFileAtLine(filePath, lineNum)
}
