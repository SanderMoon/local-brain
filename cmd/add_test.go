package cmd

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestRunAdd_StripNewlines(t *testing.T) {
	// Set up test brain environment
	tb := testutil.SetupTestBrain(t)

	// Test cases with newlines
	tests := []struct {
		name           string
		input          []string
		expectedInDump string
	}{
		{
			name:           "single newline in middle",
			input:          []string{"Fix\nbug"},
			expectedInDump: "Fix bug",
		},
		{
			name:           "multiple newlines",
			input:          []string{"Task\nwith\nmultiple\nlines"},
			expectedInDump: "Task with multiple lines",
		},
		{
			name:           "carriage return",
			input:          []string{"Windows\r\nstyle"},
			expectedInDump: "Windows  style", // Two spaces because both \r and \n become spaces
		},
		{
			name:           "leading and trailing newlines",
			input:          []string{"\nTask with spaces\n"},
			expectedInDump: "Task with spaces",
		},
		{
			name:           "no newlines",
			input:          []string{"Normal", "task"},
			expectedInDump: "Normal task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear dump file
			if err := os.WriteFile(tb.DumpPath, []byte("# Dump\n\n"), 0644); err != nil {
				t.Fatalf("Failed to reset dump: %v", err)
			}

			// Run the add command
			cmd := addCmd
			if err := runAdd(cmd, tt.input); err != nil {
				t.Fatalf("runAdd failed: %v", err)
			}

			// Read dump and verify
			content, err := os.ReadFile(tb.DumpPath)
			if err != nil {
				t.Fatalf("Failed to read dump: %v", err)
			}

			contentStr := string(content)

			// Verify the expected text is in the dump
			if !strings.Contains(contentStr, tt.expectedInDump) {
				t.Errorf("Expected dump to contain '%s', got:\n%s", tt.expectedInDump, contentStr)
			}

			// Verify it's a single line (no embedded newlines in the task)
			lines := strings.Split(contentStr, "\n")
			taskLineFound := false
			for _, line := range lines {
				if strings.Contains(line, tt.expectedInDump) {
					taskLineFound = true
					// Verify it's a proper task line
					if !strings.HasPrefix(strings.TrimSpace(line), "- [ ]") {
						t.Errorf("Task line should start with '- [ ]', got: %s", line)
					}
					// Verify it has a timestamp
					if !strings.Contains(line, "#captured:") {
						t.Error("Task line should have #captured: timestamp")
					}
					break
				}
			}

			if !taskLineFound {
				t.Error("Task line not found in dump")
			}
		})
	}
}

func TestRunAdd_NormalTask(t *testing.T) {
	// Set up test brain environment
	tb := testutil.SetupTestBrain(t)

	// Add a normal task
	args := []string{"Fix", "authentication", "bug"}
	if err := runAdd(addCmd, args); err != nil {
		t.Fatalf("runAdd failed: %v", err)
	}

	// Read dump and verify
	content, err := os.ReadFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to read dump: %v", err)
	}

	contentStr := string(content)

	// Verify task is present
	if !strings.Contains(contentStr, "Fix authentication bug") {
		t.Errorf("Expected task 'Fix authentication bug' in dump, got:\n%s", contentStr)
	}

	// Verify format
	today := time.Now().Format("2006-01-02")
	expectedFormat := "- [ ] Fix authentication bug #captured:" + today
	if !strings.Contains(contentStr, expectedFormat) {
		t.Errorf("Expected format '%s' in dump, got:\n%s", expectedFormat, contentStr)
	}
}
