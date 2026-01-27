package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDumpFile(t *testing.T) {
	// Create a temporary dump file
	tmpDir := t.TempDir()
	dumpFile := filepath.Join(tmpDir, "00_dump.md")

	content := `# Dump

- [ ] Fix authentication bug #captured:2024-01-21
- [ ] Review PR #123

[Note] Meeting notes #captured:2024-01-22
    Discussed project timeline
    Need to follow up with team

- [ ] Another task

[Note] Ideas for next sprint
    Feature A
    Feature B
    Feature C
`

	err := os.WriteFile(dumpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the dump file
	items, err := ParseDumpFile(dumpFile)
	if err != nil {
		t.Fatalf("ParseDumpFile failed: %v", err)
	}

	// Verify number of items
	expectedCount := 5 // 3 tasks + 2 notes
	if len(items) != expectedCount {
		t.Fatalf("Expected %d items, got %d", expectedCount, len(items))
	}

	// Verify first task
	if items[0].Type != ItemTypeTodo {
		t.Errorf("Item 0: expected type %s, got %s", ItemTypeTodo, items[0].Type)
	}
	if items[0].Content != "Fix authentication bug #captured:2024-01-21" {
		t.Errorf("Item 0: unexpected content: %s", items[0].Content)
	}
	if items[0].StartLine != 3 || items[0].EndLine != 3 {
		t.Errorf("Item 0: expected lines 3-3, got %d-%d", items[0].StartLine, items[0].EndLine)
	}

	// Verify second task
	if items[1].Type != ItemTypeTodo {
		t.Errorf("Item 1: expected type %s, got %s", ItemTypeTodo, items[1].Type)
	}

	// Verify first note
	if items[2].Type != ItemTypeNote {
		t.Errorf("Item 2: expected type %s, got %s", ItemTypeNote, items[2].Type)
	}
	if items[2].Content != "Meeting notes #captured:2024-01-22" {
		t.Errorf("Item 2: unexpected content: %s", items[2].Content)
	}
	// Note starts at line 6, ends at line 8 (last indented line)
	if items[2].StartLine != 6 {
		t.Errorf("Item 2: expected start line 6, got %d", items[2].StartLine)
	}

	// Verify third task
	if items[3].Type != ItemTypeTodo {
		t.Errorf("Item 3: expected type %s, got %s", ItemTypeTodo, items[3].Type)
	}

	// Verify second note
	if items[4].Type != ItemTypeNote {
		t.Errorf("Item 4: expected type %s, got %s", ItemTypeNote, items[4].Type)
	}
	if items[4].Content != "Ideas for next sprint" {
		t.Errorf("Item 4: unexpected content: %s", items[4].Content)
	}
}

func TestExtractTimestamp(t *testing.T) {
	tests := []struct {
		input             string
		expectedContent   string
		expectedTimestamp string
	}{
		{
			input:             "Fix bug #captured:2024-01-21",
			expectedContent:   "Fix bug",
			expectedTimestamp: "2024-01-21",
		},
		{
			input:             "No timestamp here",
			expectedContent:   "No timestamp here",
			expectedTimestamp: "",
		},
		{
			input:             "Task with #captured:2024-12-31",
			expectedContent:   "Task with",
			expectedTimestamp: "2024-12-31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			content, timestamp := ExtractTimestamp(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if timestamp != tt.expectedTimestamp {
				t.Errorf("Expected timestamp '%s', got '%s'", tt.expectedTimestamp, timestamp)
			}
		})
	}
}

func TestIsEmptyOrWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", true},
		{"   ", true},
		{"\t\n", true},
		{"hello", false},
		{"  hello  ", false},
	}

	for _, tt := range tests {
		result := IsEmptyOrWhitespace(tt.input)
		if result != tt.expected {
			t.Errorf("IsEmptyOrWhitespace(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}
