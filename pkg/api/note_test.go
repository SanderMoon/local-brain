package api

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestListNotes(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "test-project")
	notesDir := filepath.Join(projectDir, "notes")

	// Create note files
	note1 := filepath.Join(notesDir, "note1.md")
	note1Content := `# First Note

Created: 2024-01-15

This is the first note.
`
	tb.WriteFile(note1, note1Content)

	time.Sleep(10 * time.Millisecond) // Ensure different mtime

	note2 := filepath.Join(notesDir, "note2.md")
	note2Content := `# Second Note

Created: 2024-01-20

This is the second note.
`
	tb.WriteFile(note2, note2Content)

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}

	// Should be sorted by mtime (newest first)
	// note2 was created later, should be first
	if notes[0].Filename != "note2.md" {
		t.Errorf("Expected newest note first, got %s", notes[0].Filename)
	}

	// Verify titles
	if notes[0].Title != "Second Note" {
		t.Errorf("Expected title 'Second Note', got '%s'", notes[0].Title)
	}

	// Verify created dates
	if notes[0].Created != "2024-01-20" {
		t.Errorf("Expected created '2024-01-20', got '%s'", notes[0].Created)
	}

	// Verify project
	if notes[0].Project != "test-project" {
		t.Errorf("Expected project 'test-project', got '%s'", notes[0].Project)
	}
}

func TestListNotes_NoNotesDir(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "no-notes")

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Expected 0 notes, got %d", len(notes))
	}
}

func TestListNotes_EmptyNotesDir(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "empty-notes")
	notesDir := filepath.Join(projectDir, "notes")
	tb.WriteFile(filepath.Join(notesDir, ".gitkeep"), "")

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 0 {
		t.Errorf("Expected 0 notes, got %d", len(notes))
	}
}

func TestListNotes_NoTitle(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "no-title")
	notesDir := filepath.Join(projectDir, "notes")

	noteFile := filepath.Join(notesDir, "plain.md")
	noteContent := `This is a plain file without a markdown title.

Just some content.
`
	tb.WriteFile(noteFile, noteContent)

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	// Title should be first line
	if notes[0].Title != "This is a plain file without a markdown title." {
		t.Errorf("Unexpected title: '%s'", notes[0].Title)
	}
}

func TestListNotes_NoCreatedDate(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "no-date")
	notesDir := filepath.Join(projectDir, "notes")

	noteFile := filepath.Join(notesDir, "note.md")
	noteContent := `# Note without date

Some content here.
`
	tb.WriteFile(noteFile, noteContent)

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	// Created should be empty
	if notes[0].Created != "" {
		t.Errorf("Expected empty created date, got '%s'", notes[0].Created)
	}
}

func TestListNotes_SortByModTime(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "sorted")
	notesDir := filepath.Join(projectDir, "notes")

	// Create notes with different times
	for i := 1; i <= 3; i++ {
		noteFile := filepath.Join(notesDir, string(rune('a'+i-1))+".md")
		content := `# Note ` + string(rune('0'+i)) + "\n"
		tb.WriteFile(noteFile, content)
		time.Sleep(5 * time.Millisecond) // Ensure different mtimes
	}

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	if len(notes) != 3 {
		t.Fatalf("Expected 3 notes, got %d", len(notes))
	}

	// Should be in reverse order (newest first)
	expectedOrder := []string{"c.md", "b.md", "a.md"}
	for i, note := range notes {
		if note.Filename != expectedOrder[i] {
			t.Errorf("Position %d: expected %s, got %s", i, expectedOrder[i], note.Filename)
		}
	}

	// Verify mtimes are in descending order
	for i := 0; i < len(notes)-1; i++ {
		if notes[i].ModTime.Before(notes[i+1].ModTime) {
			t.Errorf("Notes not sorted by mtime: %v < %v", notes[i].ModTime, notes[i+1].ModTime)
		}
	}
}

func TestListNotes_OnlyMarkdownFiles(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "mixed")
	notesDir := filepath.Join(projectDir, "notes")

	// Create markdown files
	tb.WriteFile(filepath.Join(notesDir, "note1.md"), "# Note 1\n")
	tb.WriteFile(filepath.Join(notesDir, "note2.md"), "# Note 2\n")

	// Create non-markdown files (should be ignored)
	tb.WriteFile(filepath.Join(notesDir, "readme.txt"), "text file\n")
	tb.WriteFile(filepath.Join(notesDir, "data.json"), "{}\n")

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}

	// Should only find .md files
	if len(notes) != 2 {
		t.Fatalf("Expected 2 notes, got %d", len(notes))
	}
}
