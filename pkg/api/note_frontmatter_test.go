package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestParseFrontmatter_Valid(t *testing.T) {
	content := "---\ntitle: My Note\ndate: 2024-03-10\nproject: my-project\ntags: []\n---\n\n# My Note\n\nSome content.\n"

	title, date, project, ok := parseFrontmatter(content)
	if !ok {
		t.Fatal("Expected ok=true for valid frontmatter, got false")
	}
	if title != "My Note" {
		t.Errorf("Expected title 'My Note', got '%s'", title)
	}
	if date != "2024-03-10" {
		t.Errorf("Expected date '2024-03-10', got '%s'", date)
	}
	if project != "my-project" {
		t.Errorf("Expected project 'my-project', got '%s'", project)
	}
}

func TestParseFrontmatter_NoFrontmatter(t *testing.T) {
	content := "# My Note\n\nCreated: 2024-03-10\n\nSome content.\n"

	_, _, _, ok := parseFrontmatter(content)
	if ok {
		t.Error("Expected ok=false for content without frontmatter, got true")
	}
}

func TestParseFrontmatter_NoClosingDelimiter(t *testing.T) {
	content := "---\ntitle: My Note\ndate: 2024-03-10\nproject: my-project\n"

	_, _, _, ok := parseFrontmatter(content)
	if ok {
		t.Error("Expected ok=false when opening '---' has no closing '---', got true")
	}
}

func TestCreateNoteFile_YAMLFrontmatter(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("my-project")

	filePath, err := CreateNoteFile(projectPath, "My Title", "Some body text", "2024-05-01")
	if err != nil {
		t.Fatalf("CreateNoteFile failed: %v", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read note file: %v", err)
	}
	contentStr := string(content)

	if !strings.Contains(contentStr, "---") {
		t.Error("Expected '---' frontmatter delimiter in created note")
	}
	if !strings.Contains(contentStr, "title: My Title") {
		t.Error("Expected 'title: My Title' in frontmatter")
	}
	if !strings.Contains(contentStr, "date: 2024-05-01") {
		t.Error("Expected 'date: 2024-05-01' in frontmatter")
	}
	if !strings.Contains(contentStr, "project: my-project") {
		t.Error("Expected 'project: my-project' in frontmatter")
	}
	if !strings.Contains(contentStr, "tags: []") {
		t.Error("Expected 'tags: []' in frontmatter")
	}
	if !strings.Contains(contentStr, "# My Title") {
		t.Error("Expected '# My Title' heading in note body")
	}
	if !strings.Contains(contentStr, "Some body text") {
		t.Error("Expected body content in note")
	}
}

func TestParseNoteFile_Frontmatter(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "fm-project")
	notesDir := filepath.Join(projectDir, "notes")

	noteFile := filepath.Join(notesDir, "2024-06-01-my-note.md")
	noteContent := "---\ntitle: Frontmatter Note\ndate: 2024-06-01\nproject: fm-project\ntags: []\n---\n\n# Frontmatter Note\n\nContent here.\n"
	tb.WriteFile(noteFile, noteContent)

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	note := notes[0]
	if note.Title != "Frontmatter Note" {
		t.Errorf("Expected title 'Frontmatter Note', got '%s'", note.Title)
	}
	if note.Created != "2024-06-01" {
		t.Errorf("Expected created '2024-06-01', got '%s'", note.Created)
	}
}

func TestParseNoteFile_LegacyFallback(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectDir := filepath.Join(tb.ActiveDirPath, "legacy-project")
	notesDir := filepath.Join(projectDir, "notes")

	noteFile := filepath.Join(notesDir, "2024-01-10-old-note.md")
	// Old format: no frontmatter, uses "# Title" and "Created:" line
	noteContent := "# Legacy Note\n\nCreated: 2024-01-10\n\nOld content.\n"
	tb.WriteFile(noteFile, noteContent)

	notes, err := ListNotes(projectDir)
	if err != nil {
		t.Fatalf("ListNotes failed: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("Expected 1 note, got %d", len(notes))
	}

	note := notes[0]
	if note.Title != "Legacy Note" {
		t.Errorf("Expected title 'Legacy Note', got '%s'", note.Title)
	}
	if note.Created != "2024-01-10" {
		t.Errorf("Expected created '2024-01-10', got '%s'", note.Created)
	}
}
