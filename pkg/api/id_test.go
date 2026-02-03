package api

import (
	"regexp"
	"strings"
	"testing"
)

// Tests for stable ID system (v2)
// IDs are now stored as #id:xxxxxx tags in the markdown and remain stable across modifications

func TestExtractID(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "task with ID",
			line:     "- [ ] Fix bug #id:abc123 #captured:2024-01-21",
			expected: "abc123",
		},
		{
			name:     "task with ID and other tags",
			line:     "- [ ] Fix bug #p:1 #id:def456 #security",
			expected: "def456",
		},
		{
			name:     "task without ID",
			line:     "- [ ] Fix bug #captured:2024-01-21",
			expected: "",
		},
		{
			name:     "note with ID",
			line:     "[Note] Meeting notes #id:789abc #captured:2024-01-21",
			expected: "789abc",
		},
		{
			name:     "ID at end of line",
			line:     "- [ ] Task content #id:ffffff",
			expected: "ffffff",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractID(tt.line)
			if result != tt.expected {
				t.Errorf("ExtractID(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestGenerateNewID(t *testing.T) {
	// Test ID format
	for i := 0; i < 100; i++ {
		id := GenerateNewID()

		// Check length
		if len(id) != 6 {
			t.Errorf("Expected 6-char ID, got %d chars: %s", len(id), id)
		}

		// Check all hex characters
		matched, err := regexp.MatchString("^[a-f0-9]{6}$", id)
		if err != nil {
			t.Fatalf("Regex error: %v", err)
		}
		if !matched {
			t.Errorf("ID contains non-hex characters: %s", id)
		}
	}

	// Test uniqueness (should be very unlikely to generate duplicates)
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateNewID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGetOrGenerateID(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		expectNew bool
	}{
		{
			name:      "line with existing ID",
			line:      "- [ ] Task #id:abc123 #captured:2024-01-21",
			expectNew: false,
		},
		{
			name:      "line without ID",
			line:      "- [ ] Task #captured:2024-01-21",
			expectNew: true,
		},
		{
			name:      "note with ID",
			line:      "[Note] Notes #id:def456",
			expectNew: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := GetOrGenerateID(tt.line)

			// Check format
			if len(id) != 6 {
				t.Errorf("Expected 6-char ID, got %d chars: %s", len(id), id)
			}

			matched, _ := regexp.MatchString("^[a-f0-9]{6}$", id)
			if !matched {
				t.Errorf("ID contains non-hex characters: %s", id)
			}

			// If line had an ID, verify it was extracted correctly
			if !tt.expectNew {
				extracted := ExtractID(tt.line)
				if id != extracted {
					t.Errorf("GetOrGenerateID returned %s, but ExtractID returned %s", id, extracted)
				}
			}
		})
	}
}

func TestAddIDToLine(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectModified bool
	}{
		{
			name:           "task without ID",
			line:           "- [ ] Fix bug #captured:2024-01-21",
			expectModified: true,
		},
		{
			name:           "task with existing ID",
			line:           "- [ ] Fix bug #id:abc123 #captured:2024-01-21",
			expectModified: false,
		},
		{
			name:           "note without ID",
			line:           "[Note] Meeting notes #captured:2024-01-21",
			expectModified: true,
		},
		{
			name:           "empty line",
			line:           "",
			expectModified: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newLine, id := AddIDToLine(tt.line)

			// Check ID format
			if len(id) != 6 {
				t.Errorf("Expected 6-char ID, got %d chars: %s", len(id), id)
			}

			matched, _ := regexp.MatchString("^[a-f0-9]{6}$", id)
			if !matched {
				t.Errorf("ID contains non-hex characters: %s", id)
			}

			// Check line modification
			if tt.expectModified {
				if newLine == tt.line {
					t.Errorf("Expected line to be modified, but it wasn't")
				}
				if !strings.Contains(newLine, "#id:"+id) {
					t.Errorf("Modified line doesn't contain #id:%s: %s", id, newLine)
				}
			} else {
				if newLine != tt.line {
					t.Errorf("Expected line to be unchanged, but it was modified: %s -> %s", tt.line, newLine)
				}
			}

			// Verify we can extract the ID back
			extractedID := ExtractID(newLine)
			if extractedID != id {
				t.Errorf("Could not extract ID back. Added %s, extracted %s", id, extractedID)
			}
		})
	}
}

func TestRemoveIDFromContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "content with ID",
			content:  "Fix bug #id:abc123 #captured:2024-01-21",
			expected: "Fix bug  #captured:2024-01-21",
		},
		{
			name:     "content with ID and other tags",
			content:  "Fix bug #p:1 #id:def456 #security",
			expected: "Fix bug #p:1  #security",
		},
		{
			name:     "content without ID",
			content:  "Fix bug #captured:2024-01-21",
			expected: "Fix bug #captured:2024-01-21",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "ID at end",
			content:  "Task content #id:ffffff",
			expected: "Task content ",
		},
		{
			name:     "multiple spaces preserved",
			content:  "Task  content  #id:123456  #tag",
			expected: "Task  content    #tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveIDFromContent(tt.content)
			if result != tt.expected {
				t.Errorf("RemoveIDFromContent(%q) = %q, want %q", tt.content, result, tt.expected)
			}
		})
	}
}

func TestIDStability(t *testing.T) {
	// Test that IDs remain stable through typical modification workflows
	originalLine := "- [ ] Original task #captured:2024-01-21"

	// Add ID
	line1, id1 := AddIDToLine(originalLine)
	t.Logf("Added ID: %s", id1)

	// Simulate adding priority
	line2 := strings.Replace(line1, "#captured", "#p:1 #captured", 1)
	id2 := GetOrGenerateID(line2)
	if id1 != id2 {
		t.Errorf("ID changed after adding priority: %s -> %s", id1, id2)
	}

	// Simulate adding due date
	line3 := line2 + " #due:2024-02-01"
	id3 := GetOrGenerateID(line3)
	if id1 != id3 {
		t.Errorf("ID changed after adding due date: %s -> %s", id1, id3)
	}

	// Simulate adding tags
	line4 := line3 + " #bug #urgent"
	id4 := GetOrGenerateID(line4)
	if id1 != id4 {
		t.Errorf("ID changed after adding tags: %s -> %s", id1, id4)
	}

	// Simulate changing status
	line5 := strings.Replace(line4, "- [ ]", "- [>]", 1)
	id5 := GetOrGenerateID(line5)
	if id1 != id5 {
		t.Errorf("ID changed after changing status: %s -> %s", id1, id5)
	}

	t.Logf("ID remained stable through all modifications: %s", id1)
}
