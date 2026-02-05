package api

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterTodosByTemporal(t *testing.T) {
	todos := []TodoItem{
		{ID: "1", Content: "Task 1", Status: "open", CapturedDate: "2026-01-15", CompletedDate: ""},
		{ID: "2", Content: "Task 2", Status: "done", CapturedDate: "2026-01-20", CompletedDate: "2026-01-25"},
		{ID: "3", Content: "Task 3", Status: "open", CapturedDate: "2026-02-01", CompletedDate: ""},
		{ID: "4", Content: "Task 4", Status: "done", CapturedDate: "2026-01-10", CompletedDate: "2026-02-05"},
	}

	testCases := []struct {
		name            string
		createdAfter    string
		createdBefore   string
		completedAfter  string
		completedBefore string
		expectedCount   int
		expectedIDs     []string
	}{
		{
			name:          "Filter by createdAfter",
			createdAfter:  "2026-01-20",
			expectedCount: 2,
			expectedIDs:   []string{"2", "3"},
		},
		{
			name:          "Filter by createdBefore",
			createdBefore: "2026-01-20",
			expectedCount: 3,
			expectedIDs:   []string{"1", "2", "4"},
		},
		{
			name:           "Filter by completedAfter",
			completedAfter: "2026-02-01",
			expectedCount:  1,
			expectedIDs:    []string{"4"},
		},
		{
			name:           "Combined filters",
			createdAfter:   "2026-01-15",
			createdBefore:  "2026-02-01",
			completedAfter: "2026-01-25",
			expectedCount:  1,
			expectedIDs:    []string{"2"},
		},
		{
			name:          "No matches",
			createdAfter:  "2026-03-01",
			expectedCount: 0,
			expectedIDs:   []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filtered := FilterTodosByTemporal(todos, tc.createdAfter, tc.createdBefore, tc.completedAfter, tc.completedBefore, "", "")

			if len(filtered) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(filtered))
			}

			if len(tc.expectedIDs) > 0 {
				for i, todo := range filtered {
					if i >= len(tc.expectedIDs) {
						t.Errorf("Got more results than expected")
						break
					}
					if todo.ID != tc.expectedIDs[i] {
						t.Errorf("Expected todo %s, got %s", tc.expectedIDs[i], todo.ID)
					}
				}
			}
		})
	}
}

func TestUnifiedSearch_TodosOnly(t *testing.T) {
	tmpDir := t.TempDir()
	activeDir := tmpDir

	// Create a project with todos
	projectDir := filepath.Join(activeDir, "test-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project: %v", err)
	}

	// Create todo file
	todoContent := `# Tasks

## Active

- [ ] Important bug to fix #p:1
- [ ] Feature request from user
- [ ] Documentation update

## Completed

- [x] Old completed task
`

	todoFile := filepath.Join(projectDir, "todo.md")
	if err := os.WriteFile(todoFile, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Search for todos only
	results, err := UnifiedSearch(activeDir, "fix", "", true, false, false, "", "", "", "", "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find the bug fix todo
	if len(results) == 0 {
		t.Errorf("Expected to find search results, got 0")
	}

	foundBug := false
	for _, result := range results {
		if result.Type == ResultTypeTodo && result.Todo != nil && result.Todo.Status == "open" {
			foundBug = true
			break
		}
	}

	if !foundBug {
		t.Errorf("Expected to find bug fix todo in results")
	}
}

func TestUnifiedSearch_NotesOnly(t *testing.T) {
	tmpDir := t.TempDir()
	activeDir := tmpDir

	// Create a project with notes
	projectDir := filepath.Join(activeDir, "test-project")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}

	// Create a note file
	noteContent := `# Important Meeting Notes

Created: 2026-01-15

This note contains important information about our product strategy.
`

	noteFile := filepath.Join(notesDir, "2026-01-15-meeting.md")
	if err := os.WriteFile(noteFile, []byte(noteContent), 0644); err != nil {
		t.Fatalf("Failed to create note file: %v", err)
	}

	// Search for notes only with title match
	results, err := UnifiedSearch(activeDir, "meeting", "", false, true, false, "", "", "", "", "", "")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	// Should find the note
	if len(results) == 0 {
		t.Errorf("Expected to find search results, got 0")
	}

	foundNote := false
	for _, result := range results {
		if result.Type == ResultTypeNote && result.Note != nil {
			foundNote = true
			break
		}
	}

	if !foundNote {
		t.Errorf("Expected to find note in results")
	}
}

func TestSearchNotes(t *testing.T) {
	tmpDir := t.TempDir()
	activeDir := tmpDir

	// Create a project with notes
	projectDir := filepath.Join(activeDir, "test-project")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes directory: %v", err)
	}

	// Create note files
	note1 := `# Database Performance

Created: 2026-01-20

Issues with slow queries and indexing.
`

	note2 := `# API Design Notes

Created: 2026-01-25

Design decisions for REST API endpoints.
`

	if err := os.WriteFile(filepath.Join(notesDir, "2026-01-20-db.md"), []byte(note1), 0644); err != nil {
		t.Fatalf("Failed to create note 1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notesDir, "2026-01-25-api.md"), []byte(note2), 0644); err != nil {
		t.Fatalf("Failed to create note 2: %v", err)
	}

	// Test: Search by title
	results, err := SearchNotes(activeDir, "Database", "", false)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Database', got %d", len(results))
	}

	// Test: Search by content (if searchContent is true)
	results, err = SearchNotes(activeDir, "queries", "", true)
	if err != nil {
		t.Fatalf("Content search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'queries' in content, got %d", len(results))
	}

	// Test: Search in specific project
	results, err = SearchNotes(activeDir, "API", "test-project", false)
	if err != nil {
		t.Fatalf("Project filter search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result in test-project, got %d", len(results))
	}
}
