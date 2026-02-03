package api

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "special characters",
			input:    "Hello, World! How are you?",
			expected: "hello-world-how-are-you",
		},
		{
			name:     "multiple spaces",
			input:    "Hello    World",
			expected: "hello-world",
		},
		{
			name:     "hyphens already present",
			input:    "hello-world-test",
			expected: "hello-world-test",
		},
		{
			name:     "leading and trailing spaces",
			input:    "  hello world  ",
			expected: "hello-world",
		},
		{
			name:     "leading and trailing hyphens",
			input:    "---hello-world---",
			expected: "hello-world",
		},
		{
			name:     "numbers",
			input:    "test-123-456",
			expected: "test-123-456",
		},
		{
			name:     "mixed alphanumeric",
			input:    "Project 2024 v2.1",
			expected: "project-2024-v21",
		},
		{
			name:     "unicode characters",
			input:    "Héllo Wörld Tëst",
			expected: "hllo-wrld-tst",
		},
		{
			name:     "only special characters",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "long text (over 40 chars)",
			input:    "This is a very long title that exceeds forty characters",
			expected: "this-is-a-very-long-title-that-exceeds-f",
		},
		{
			name:     "exactly 40 chars",
			input:    "1234567890123456789012345678901234567890",
			expected: "1234567890123456789012345678901234567890",
		},
		{
			name:     "underscores",
			input:    "hello_world_test",
			expected: "helloworldtest",
		},
		{
			name:     "dots",
			input:    "hello.world.test",
			expected: "helloworldtest",
		},
		{
			name:     "only hyphens",
			input:    "---",
			expected: "",
		},
		{
			name:     "mixed case",
			input:    "HeLLo WoRLd",
			expected: "hello-world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, expected %q", tt.input, result, tt.expected)
			}

			// Verify slug is filename-safe
			if result != "" {
				// Should not contain special characters
				if strings.ContainsAny(result, " !@#$%^&*()[]{}|\\:;\"'<>?,./") {
					t.Errorf("Slug contains special characters: %q", result)
				}

				// Should not start or end with hyphen
				if strings.HasPrefix(result, "-") || strings.HasSuffix(result, "-") {
					t.Errorf("Slug has leading/trailing hyphens: %q", result)
				}

				// Should not exceed 40 characters
				if len(result) > 40 {
					t.Errorf("Slug exceeds 40 characters: %q (len=%d)", result, len(result))
				}
			}
		})
	}
}

func TestExtractCapturedDate(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedText   string
		expectedDate   string
	}{
		{
			name:           "task with captured date",
			input:          "Fix authentication bug #captured:2024-01-21",
			expectedText:   "Fix authentication bug",
			expectedDate:   "2024-01-21",
		},
		{
			name:           "note title with captured date",
			input:          "Meeting notes #captured:2024-12-15",
			expectedText:   "Meeting notes",
			expectedDate:   "2024-12-15",
		},
		{
			name:           "content without captured date",
			input:          "Some task content",
			expectedText:   "Some task content",
			expectedDate:   "",
		},
		{
			name:           "multiple metadata tags",
			input:          "Task #p:1 #due:2024-02-15 #captured:2024-01-21 #bug",
			expectedText:   "Task #p:1 #due:2024-02-15 #bug",
			expectedDate:   "2024-01-21",
		},
		{
			name:           "captured date in middle",
			input:          "Task #captured:2024-01-21 with more content",
			expectedText:   "Task with more content",
			expectedDate:   "2024-01-21",
		},
		{
			name:           "captured date with space before",
			input:          "Task  #captured:2024-01-21",
			expectedText:   "Task",
			expectedDate:   "2024-01-21",
		},
		{
			name:           "empty string",
			input:          "",
			expectedText:   "",
			expectedDate:   "",
		},
		{
			name:           "only captured tag",
			input:          "#captured:2024-01-21",
			expectedText:   "",
			expectedDate:   "2024-01-21",
		},
		{
			name:           "date with letters",
			input:          "Task #captured:2024-01-21",
			expectedText:   "Task",
			expectedDate:   "2024-01-21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, date := ExtractCapturedDate(tt.input)

			if text != tt.expectedText {
				t.Errorf("ExtractCapturedDate(%q) text = %q, expected %q", tt.input, text, tt.expectedText)
			}

			if date != tt.expectedDate {
				t.Errorf("ExtractCapturedDate(%q) date = %q, expected %q", tt.input, date, tt.expectedDate)
			}
		})
	}
}

func TestSlugify_Idempotent(t *testing.T) {
	// Running slugify twice should produce same result
	input := "Hello World Test"
	first := Slugify(input)
	second := Slugify(first)

	if first != second {
		t.Errorf("Slugify is not idempotent: first=%q, second=%q", first, second)
	}
}

func TestExtractCapturedDate_PreservesOtherMetadata(t *testing.T) {
	// Ensure other metadata tags are not removed
	input := "Task #p:1 #due:2024-02-15 #captured:2024-01-21 #bug #security"
	text, date := ExtractCapturedDate(input)

	if date != "2024-01-21" {
		t.Errorf("Expected date '2024-01-21', got %q", date)
	}

	// Should preserve other tags
	if !strings.Contains(text, "#p:1") {
		t.Error("Priority tag was removed")
	}
	if !strings.Contains(text, "#due:2024-02-15") {
		t.Error("Due date tag was removed")
	}
	if !strings.Contains(text, "#bug") {
		t.Error("Bug tag was removed")
	}
	if !strings.Contains(text, "#security") {
		t.Error("Security tag was removed")
	}

	// Should NOT contain captured tag
	if strings.Contains(text, "#captured") {
		t.Error("Captured tag was not removed")
	}
}
