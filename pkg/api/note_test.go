package api

import (
	"os"
	"path/filepath"
	"strings"
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

func TestUpdateNoteFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "test-project")
	notesDir := filepath.Join(projectDir, "notes")

	// Create an initial note file
	notePath := filepath.Join(notesDir, "2026-02-11-test-note.md")
	initial := "# Test Note\n\nCreated: 2026-02-11\n\nOriginal content.\n"
	tb.WriteFile(notePath, initial)

	// Update it
	updated := "# Test Note\n\nCreated: 2026-02-11\n\nUpdated content.\n"
	if err := UpdateNoteFile(notePath, updated); err != nil {
		t.Fatalf("UpdateNoteFile failed: %v", err)
	}

	// Verify new content
	got, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(got) != updated {
		t.Errorf("Expected updated content, got:\n%s", got)
	}
	if strings.Contains(string(got), "Original content") {
		t.Errorf("Old content still present after update")
	}
}

func TestUpdateNoteFile_NotFound(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	notePath := filepath.Join(tb.ActiveDirPath, "nonexistent", "missing.md")
	err := UpdateNoteFile(notePath, "some content")
	if err == nil {
		t.Error("Expected error for non-existent note, got nil")
	}
}

func TestAppendNoteLink_NoNotesFile(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "notes"), 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	err := AppendNoteLink(projectDir, "2026-02-11", "Weekly Update", "2026-02-11-weekly-update.md")
	if err != nil {
		t.Fatalf("AppendNoteLink failed: %v", err)
	}

	notesIndexPath := filepath.Join(projectDir, "notes.md")
	content, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("notes.md not created: %v", err)
	}

	expected := "- [2026-02-11 Weekly Update](notes/2026-02-11-weekly-update.md)"
	if !strings.Contains(string(content), expected) {
		t.Errorf("Expected link %q in notes.md, got:\n%s", expected, content)
	}
}

func TestAppendNoteLink_NoNotesSection(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "notes"), 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Write a notes.md without a "## Notes" section
	notesIndexPath := filepath.Join(projectDir, "notes.md")
	initialContent := "# My Project Notes\n\nSome introductory text.\n"
	if err := os.WriteFile(notesIndexPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write notes.md: %v", err)
	}

	err := AppendNoteLink(projectDir, "2026-02-11", "Weekly Update", "2026-02-11-weekly-update.md")
	if err != nil {
		t.Fatalf("AppendNoteLink failed: %v", err)
	}

	content, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("Failed to read notes.md: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "## Notes") {
		t.Error("Expected '## Notes' section to be added")
	}

	expected := "- [2026-02-11 Weekly Update](notes/2026-02-11-weekly-update.md)"
	if !strings.Contains(contentStr, expected) {
		t.Errorf("Expected link %q in notes.md, got:\n%s", expected, contentStr)
	}
}

func TestAppendNoteLink_ExistingSection(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "notes"), 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	notesIndexPath := filepath.Join(projectDir, "notes.md")
	initialContent := "# My Project Notes\n\n## Notes\n\n- [2026-01-01 Old Note](notes/2026-01-01-old-note.md)\n"
	if err := os.WriteFile(notesIndexPath, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to write notes.md: %v", err)
	}

	err := AppendNoteLink(projectDir, "2026-02-11", "Weekly Update", "2026-02-11-weekly-update.md")
	if err != nil {
		t.Fatalf("AppendNoteLink failed: %v", err)
	}

	content, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("Failed to read notes.md: %v", err)
	}
	contentStr := string(content)

	newLink := "- [2026-02-11 Weekly Update](notes/2026-02-11-weekly-update.md)"
	oldLink := "- [2026-01-01 Old Note](notes/2026-01-01-old-note.md)"

	if !strings.Contains(contentStr, newLink) {
		t.Errorf("Expected new link %q in notes.md, got:\n%s", newLink, contentStr)
	}
	if !strings.Contains(contentStr, oldLink) {
		t.Errorf("Expected old link %q still in notes.md, got:\n%s", oldLink, contentStr)
	}

	// New link should appear before the old link (newest first)
	newIdx := strings.Index(contentStr, newLink)
	oldIdx := strings.Index(contentStr, oldLink)
	if newIdx > oldIdx {
		t.Errorf("New link should appear before old link (newest first), got:\n%s", contentStr)
	}
}

func TestAppendNoteLink_Idempotent(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, "notes"), 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	filename := "2026-02-11-weekly-update.md"
	err := AppendNoteLink(projectDir, "2026-02-11", "Weekly Update", filename)
	if err != nil {
		t.Fatalf("First AppendNoteLink failed: %v", err)
	}

	err = AppendNoteLink(projectDir, "2026-02-11", "Weekly Update", filename)
	if err != nil {
		t.Fatalf("Second AppendNoteLink failed: %v", err)
	}

	notesIndexPath := filepath.Join(projectDir, "notes.md")
	content, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("Failed to read notes.md: %v", err)
	}
	contentStr := string(content)

	linkStr := "notes/" + filename
	count := strings.Count(contentStr, linkStr)
	if count != 1 {
		t.Errorf("Expected exactly 1 occurrence of the link, got %d:\n%s", count, contentStr)
	}
}

func TestCreateNoteFile_AppendsToNotesIndex(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("my-project")

	filePath, err := CreateNoteFile(projectPath, "Weekly Update", "Content here", "2026-02-11")
	if err != nil {
		t.Fatalf("CreateNoteFile failed: %v", err)
	}

	notesIndexPath := filepath.Join(projectPath, "notes.md")
	content, err := os.ReadFile(notesIndexPath)
	if err != nil {
		t.Fatalf("Failed to read notes.md: %v", err)
	}
	contentStr := string(content)

	expectedFilename := filepath.Base(filePath)
	expectedLink := "- [2026-02-11 Weekly Update](notes/" + expectedFilename + ")"
	if !strings.Contains(contentStr, expectedLink) {
		t.Errorf("Expected link %q in notes.md, got:\n%s", expectedLink, contentStr)
	}
}
