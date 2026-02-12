package markdown

import (
	"fmt"
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

func TestExtractPriority(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContent  string
		expectedPriority *int
	}{
		{
			name:             "high priority",
			input:            "Fix critical bug #p:1",
			expectedContent:  "Fix critical bug",
			expectedPriority: intPtr(1),
		},
		{
			name:             "medium priority",
			input:            "Review PR #p:2",
			expectedContent:  "Review PR",
			expectedPriority: intPtr(2),
		},
		{
			name:             "low priority",
			input:            "Update docs #p:3",
			expectedContent:  "Update docs",
			expectedPriority: intPtr(3),
		},
		{
			name:             "no priority",
			input:            "Regular task",
			expectedContent:  "Regular task",
			expectedPriority: nil,
		},
		{
			name:             "invalid priority (0)",
			input:            "Task #p:0",
			expectedContent:  "Task #p:0",
			expectedPriority: nil,
		},
		{
			name:             "invalid priority (4)",
			input:            "Task #p:4",
			expectedContent:  "Task #p:4",
			expectedPriority: nil,
		},
		{
			name:             "priority with timestamp",
			input:            "Fix bug #p:1 #captured:2024-01-21",
			expectedContent:  "Fix bug #captured:2024-01-21",
			expectedPriority: intPtr(1),
		},
		{
			name:             "priority at start",
			input:            "#p:2 Important task",
			expectedContent:  "Important task",
			expectedPriority: intPtr(2),
		},
		{
			name:             "priority with extra spaces",
			input:            "Task  #p:1  extra spaces",
			expectedContent:  "Task  extra spaces",
			expectedPriority: intPtr(1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, priority := ExtractPriority(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if !equalIntPtr(priority, tt.expectedPriority) {
				t.Errorf("Expected priority %v, got %v", formatIntPtr(tt.expectedPriority), formatIntPtr(priority))
			}
		})
	}
}

func TestExtractDueDate(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedDueDate string
	}{
		{
			name:            "due date at end",
			input:           "Fix bug #due:2026-02-15",
			expectedContent: "Fix bug",
			expectedDueDate: "2026-02-15",
		},
		{
			name:            "no due date",
			input:           "Regular task",
			expectedContent: "Regular task",
			expectedDueDate: "",
		},
		{
			name:            "due date with priority",
			input:           "Important task #p:1 #due:2026-03-01",
			expectedContent: "Important task #p:1",
			expectedDueDate: "2026-03-01",
		},
		{
			name:            "due date at start",
			input:           "#due:2026-01-20 Task description",
			expectedContent: "Task description",
			expectedDueDate: "2026-01-20",
		},
		{
			name:            "due date with extra spaces",
			input:           "Task  #due:2026-04-15  extra spaces",
			expectedContent: "Task  extra spaces",
			expectedDueDate: "2026-04-15",
		},
		{
			name:            "invalid due date format (should still extract)",
			input:           "Task #due:not-a-date",
			expectedContent: "Task",
			expectedDueDate: "not-a-date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, dueDate := ExtractDueDate(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if dueDate != tt.expectedDueDate {
				t.Errorf("Expected due date '%s', got '%s'", tt.expectedDueDate, dueDate)
			}
		})
	}
}

func TestExtractTags(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedTags    []string
	}{
		{
			name:            "single tag",
			input:           "Fix bug #bug",
			expectedContent: "Fix bug",
			expectedTags:    []string{"bug"},
		},
		{
			name:            "multiple tags",
			input:           "Fix auth bug #bug #security #urgent",
			expectedContent: "Fix auth bug",
			expectedTags:    []string{"bug", "security", "urgent"},
		},
		{
			name:            "no tags",
			input:           "Regular task",
			expectedContent: "Regular task",
			expectedTags:    []string{},
		},
		{
			name:            "tags with metadata (should ignore metadata)",
			input:           "Important task #p:1 #bug #due:2026-02-15 #feature",
			expectedContent: "Important task #p:1 #due:2026-02-15",
			expectedTags:    []string{"bug", "feature"},
		},
		{
			name:            "duplicate tags (should deduplicate)",
			input:           "Task #bug #feature #bug",
			expectedContent: "Task",
			expectedTags:    []string{"bug", "feature"},
		},
		{
			name:            "tags with hyphens and underscores",
			input:           "Task #high-priority #ui_work #frontend",
			expectedContent: "Task",
			expectedTags:    []string{"high-priority", "ui_work", "frontend"},
		},
		{
			name:            "metadata only (no freeform tags)",
			input:           "Task #p:1 #due:2026-02-15 #captured:2024-01-21",
			expectedContent: "Task #p:1 #due:2026-02-15 #captured:2024-01-21",
			expectedTags:    []string{},
		},
		{
			name:            "tag at start",
			input:           "#bug Fix this issue",
			expectedContent: "Fix this issue",
			expectedTags:    []string{"bug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, tags := ExtractTags(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if len(tags) != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, got %d: %v", len(tt.expectedTags), len(tags), tags)
				return
			}

			// Check each expected tag is present
			tagMap := make(map[string]bool)
			for _, tag := range tags {
				tagMap[tag] = true
			}

			for _, expected := range tt.expectedTags {
				if !tagMap[expected] {
					t.Errorf("Expected tag '%s' not found in %v", expected, tags)
				}
			}
		})
	}
}

func TestExtractPriority_NoHash(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContent  string
		expectedPriority *int
	}{
		{
			name:             "high priority no hash",
			input:            "Fix critical bug p:1",
			expectedContent:  "Fix critical bug",
			expectedPriority: intPtr(1),
		},
		{
			name:             "medium priority no hash",
			input:            "Review PR p:2",
			expectedContent:  "Review PR",
			expectedPriority: intPtr(2),
		},
		{
			name:             "low priority no hash",
			input:            "Update docs p:3",
			expectedContent:  "Update docs",
			expectedPriority: intPtr(3),
		},
		{
			name:             "no-hash priority with due date",
			input:            "Fix bug p:1 due:2026-02-20",
			expectedContent:  "Fix bug due:2026-02-20",
			expectedPriority: intPtr(1),
		},
		{
			name:             "no-hash priority at start",
			input:            "p:2 Important task",
			expectedContent:  "Important task",
			expectedPriority: intPtr(2),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, priority := ExtractPriority(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if !equalIntPtr(priority, tt.expectedPriority) {
				t.Errorf("Expected priority %v, got %v", formatIntPtr(tt.expectedPriority), formatIntPtr(priority))
			}
		})
	}
}

func TestExtractPriority_LegacyHash(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedContent  string
		expectedPriority *int
	}{
		{
			name:             "legacy hash high priority",
			input:            "Fix critical bug #p:1",
			expectedContent:  "Fix critical bug",
			expectedPriority: intPtr(1),
		},
		{
			name:             "legacy hash medium priority",
			input:            "Review PR #p:2",
			expectedContent:  "Review PR",
			expectedPriority: intPtr(2),
		},
		{
			name:             "legacy hash low priority",
			input:            "Update docs #p:3",
			expectedContent:  "Update docs",
			expectedPriority: intPtr(3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, priority := ExtractPriority(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if !equalIntPtr(priority, tt.expectedPriority) {
				t.Errorf("Expected priority %v, got %v", formatIntPtr(tt.expectedPriority), formatIntPtr(priority))
			}
		})
	}
}

func TestExtractDueDate_NoHash(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedDueDate string
	}{
		{
			name:            "no-hash due date at end",
			input:           "Fix bug due:2026-02-15",
			expectedContent: "Fix bug",
			expectedDueDate: "2026-02-15",
		},
		{
			name:            "no-hash due date at start",
			input:           "due:2026-01-20 Task description",
			expectedContent: "Task description",
			expectedDueDate: "2026-01-20",
		},
		{
			name:            "no-hash due date with priority",
			input:           "Important task p:1 due:2026-03-01",
			expectedContent: "Important task p:1",
			expectedDueDate: "2026-03-01",
		},
		{
			name:            "no-hash due date with extra spaces",
			input:           "Task  due:2026-04-15  extra spaces",
			expectedContent: "Task extra spaces",
			expectedDueDate: "2026-04-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, dueDate := ExtractDueDate(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if dueDate != tt.expectedDueDate {
				t.Errorf("Expected due date '%s', got '%s'", tt.expectedDueDate, dueDate)
			}
		})
	}
}

func TestExtractDueDate_LegacyHash(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedContent string
		expectedDueDate string
	}{
		{
			name:            "legacy hash due date at end",
			input:           "Fix bug #due:2026-02-15",
			expectedContent: "Fix bug",
			expectedDueDate: "2026-02-15",
		},
		{
			name:            "legacy hash due date at start",
			input:           "#due:2026-01-20 Task description",
			expectedContent: "Task description",
			expectedDueDate: "2026-01-20",
		},
		{
			name:            "legacy hash due date with priority",
			input:           "Important task #p:1 #due:2026-03-01",
			expectedContent: "Important task #p:1",
			expectedDueDate: "2026-03-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, dueDate := ExtractDueDate(tt.input)

			if content != tt.expectedContent {
				t.Errorf("Expected content '%s', got '%s'", tt.expectedContent, content)
			}

			if dueDate != tt.expectedDueDate {
				t.Errorf("Expected due date '%s', got '%s'", tt.expectedDueDate, dueDate)
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

func TestExtractHTMLComment_Valid(t *testing.T) {
	line := "- [ ] Task content <!-- id:abc123 captured:2026-01-30 -->"
	fields, cleanLine := ExtractHTMLComment(line)

	if fields["id"] != "abc123" {
		t.Errorf("Expected id 'abc123', got '%s'", fields["id"])
	}
	if fields["captured"] != "2026-01-30" {
		t.Errorf("Expected captured '2026-01-30', got '%s'", fields["captured"])
	}
	expected := "- [ ] Task content"
	if cleanLine != expected {
		t.Errorf("Expected clean line %q, got %q", expected, cleanLine)
	}
}

func TestExtractHTMLComment_NoComment(t *testing.T) {
	line := "- [ ] Task content #p:1 #due:2026-02-15"
	fields, cleanLine := ExtractHTMLComment(line)

	if len(fields) != 0 {
		t.Errorf("Expected empty fields map, got %v", fields)
	}
	if cleanLine != line {
		t.Errorf("Expected original line %q, got %q", line, cleanLine)
	}
}

func TestExtractHTMLComment_MultipleFields(t *testing.T) {
	line := "- [x] Done task <!-- id:def456 captured:2026-01-10 done:2026-02-01 -->"
	fields, cleanLine := ExtractHTMLComment(line)

	if fields["id"] != "def456" {
		t.Errorf("Expected id 'def456', got '%s'", fields["id"])
	}
	if fields["captured"] != "2026-01-10" {
		t.Errorf("Expected captured '2026-01-10', got '%s'", fields["captured"])
	}
	if fields["done"] != "2026-02-01" {
		t.Errorf("Expected done '2026-02-01', got '%s'", fields["done"])
	}
	expected := "- [x] Done task"
	if cleanLine != expected {
		t.Errorf("Expected clean line %q, got %q", expected, cleanLine)
	}
}

func TestExtractTimestamp_HTMLComment(t *testing.T) {
	line := "Task content <!-- id:abc123 captured:2026-01-30 -->"
	cleanContent, timestamp := ExtractTimestamp(line)

	if timestamp != "2026-01-30" {
		t.Errorf("Expected timestamp '2026-01-30', got '%s'", timestamp)
	}
	// The clean content should not contain the HTML comment
	if cleanContent == line {
		t.Errorf("Expected content to be cleaned, but got original line")
	}
}

func TestExtractTimestamp_LegacyInline(t *testing.T) {
	// Old format should still work
	line := "Fix bug #captured:2024-01-21"
	cleanContent, timestamp := ExtractTimestamp(line)

	if timestamp != "2024-01-21" {
		t.Errorf("Expected timestamp '2024-01-21', got '%s'", timestamp)
	}
	if cleanContent != "Fix bug" {
		t.Errorf("Expected clean content 'Fix bug', got '%s'", cleanContent)
	}
}

func TestExtractCompletedDate_HTMLComment(t *testing.T) {
	line := "Done task <!-- id:abc123 captured:2026-01-10 done:2026-02-01 -->"
	cleanContent, doneDate := ExtractCompletedDate(line)

	if doneDate != "2026-02-01" {
		t.Errorf("Expected done date '2026-02-01', got '%s'", doneDate)
	}
	// The clean content should not contain the HTML comment
	if cleanContent == line {
		t.Errorf("Expected content to be cleaned, but got original line")
	}
}

// Helper functions for testing

func intPtr(i int) *int {
	return &i
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func formatIntPtr(i *int) string {
	if i == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *i)
}
