package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// MigrateNoteToFrontmatter tests
// ---------------------------------------------------------------------------

func TestMigrateNoteToFrontmatter_AddsFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "test-project", "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	notePath := filepath.Join(notesDir, "2026-01-15-meeting.md")
	legacyContent := `# Meeting Notes

Created: 2026-01-15

Some content here.
`
	if err := os.WriteFile(notePath, []byte(legacyContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateNoteToFrontmatter(notePath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if result.Change != "added frontmatter" {
		t.Errorf("Expected Change='added frontmatter', got: %q", result.Change)
	}

	// Verify file now has frontmatter
	content, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("Failed to read migrated file: %v", err)
	}
	contentStr := string(content)

	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("Expected file to start with frontmatter delimiter '---'")
	}
	if !strings.Contains(contentStr, "title: Meeting Notes") {
		t.Errorf("Expected frontmatter to contain title, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "date: 2026-01-15") {
		t.Errorf("Expected frontmatter to contain date, got:\n%s", contentStr)
	}
	if !strings.Contains(contentStr, "project: test-project") {
		t.Errorf("Expected frontmatter to contain project, got:\n%s", contentStr)
	}
}

func TestMigrateNoteToFrontmatter_AlreadyHasFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "my-project", "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	notePath := filepath.Join(notesDir, "already-migrated.md")
	content := `---
title: Already migrated
date: 2026-01-10
project: my-project
tags: []
---

# Already migrated

Content here.
`
	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateNoteToFrontmatter(notePath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Changed {
		t.Error("Expected Changed=false for already-migrated note")
	}
	if result.Change != "already has frontmatter" {
		t.Errorf("Expected Change='already has frontmatter', got: %q", result.Change)
	}

	// Verify file is unchanged
	read, _ := os.ReadFile(notePath)
	if string(read) != content {
		t.Error("File should not have been modified")
	}
}

func TestMigrateNoteToFrontmatter_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	notesDir := filepath.Join(tmpDir, "proj", "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	notePath := filepath.Join(notesDir, "dry-run-note.md")
	original := `# Dry Run Test

Created: 2026-02-01

Content.
`
	if err := os.WriteFile(notePath, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateNoteToFrontmatter(notePath, true /* dryRun */)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true in dry run")
	}
	if result.Change != "would add frontmatter" {
		t.Errorf("Expected Change='would add frontmatter', got: %q", result.Change)
	}

	// File must remain unchanged
	read, _ := os.ReadFile(notePath)
	if string(read) != original {
		t.Error("Dry run must not modify the file")
	}
}

// ---------------------------------------------------------------------------
// MigrateTodoToHTMLComments tests
// ---------------------------------------------------------------------------

func TestMigrateTodoToHTMLComments_MovesTags(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	content := `# Tasks

- [ ] Buy groceries #id:abc123 #captured:2026-01-30
- [ ] Write tests #id:def456 #captured:2026-01-31 #done:2026-02-01
- [ ] No metadata here
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateTodoToHTMLComments(todoPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true")
	}

	// Verify file content
	migrated, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read migrated file: %v", err)
	}
	migratedStr := string(migrated)

	if strings.Contains(migratedStr, "#id:abc123") {
		t.Error("Inline #id: tag should have been removed from first todo")
	}
	if !strings.Contains(migratedStr, "<!-- id:abc123 captured:2026-01-30 -->") {
		t.Errorf("Expected HTML comment for first todo, got:\n%s", migratedStr)
	}
	if !strings.Contains(migratedStr, "<!-- id:def456 captured:2026-01-31 done:2026-02-01 -->") {
		t.Errorf("Expected HTML comment with done date for second todo, got:\n%s", migratedStr)
	}
	// Line with no metadata should be unchanged
	if !strings.Contains(migratedStr, "- [ ] No metadata here") {
		t.Error("Line without metadata should be preserved unchanged")
	}
}

func TestMigrateTodoToHTMLComments_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	// Already-migrated content (HTML comment format)
	content := `# Tasks

- [ ] Buy groceries <!-- id:abc123 captured:2026-01-30 -->
- [ ] Write tests <!-- id:def456 captured:2026-01-31 done:2026-02-01 -->
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateTodoToHTMLComments(todoPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Changed {
		t.Error("Expected Changed=false for already-migrated todos")
	}
	if result.Change != "no inline metadata found" {
		t.Errorf("Expected Change='no inline metadata found', got: %q", result.Change)
	}

	// File should be unchanged
	read, _ := os.ReadFile(todoPath)
	if string(read) != content {
		t.Error("File should not have been modified (idempotent)")
	}
}

func TestMigrateTodoToHTMLComments_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	original := `# Tasks

- [ ] Task with inline tags #id:aabbcc #captured:2026-02-10
`
	if err := os.WriteFile(todoPath, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigrateTodoToHTMLComments(todoPath, true /* dryRun */)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true in dry run")
	}
	if !strings.Contains(result.Change, "would move") {
		t.Errorf("Expected Change to mention 'would move', got: %q", result.Change)
	}

	// File must remain unchanged
	read, _ := os.ReadFile(todoPath)
	if string(read) != original {
		t.Error("Dry run must not modify the file")
	}
}

// ---------------------------------------------------------------------------
// MigrateNotesIndex tests
// ---------------------------------------------------------------------------

func TestMigrateNotesIndex_AddsLinks(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "my-project")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	// Create two note files with frontmatter
	note1 := `---
title: First Note
date: 2026-01-20
project: my-project
tags: []
---

# First Note

Content.
`
	note2 := `---
title: Second Note
date: 2026-01-25
project: my-project
tags: []
---

# Second Note

Content.
`
	if err := os.WriteFile(filepath.Join(notesDir, "2026-01-20-first-note.md"), []byte(note1), 0644); err != nil {
		t.Fatalf("Failed to write note1: %v", err)
	}
	if err := os.WriteFile(filepath.Join(notesDir, "2026-01-25-second-note.md"), []byte(note2), 0644); err != nil {
		t.Fatalf("Failed to write note2: %v", err)
	}

	// notes.md is empty / doesn't reference the files yet
	notesIndexPath := filepath.Join(projectDir, "notes.md")
	if err := os.WriteFile(notesIndexPath, []byte("# my-project Notes\n\n"), 0644); err != nil {
		t.Fatalf("Failed to write notes.md: %v", err)
	}

	result, err := MigrateNotesIndex(projectDir, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true")
	}
	if !strings.Contains(result.Change, "added 2 links") {
		t.Errorf("Expected 'added 2 links', got: %q", result.Change)
	}

	// Verify notes.md now contains the links
	indexContent, _ := os.ReadFile(notesIndexPath)
	indexStr := string(indexContent)
	if !strings.Contains(indexStr, "notes/2026-01-20-first-note.md") {
		t.Errorf("Expected first note link in notes.md, got:\n%s", indexStr)
	}
	if !strings.Contains(indexStr, "notes/2026-01-25-second-note.md") {
		t.Errorf("Expected second note link in notes.md, got:\n%s", indexStr)
	}
}

func TestMigrateNotesIndex_SkipsAlreadyLinked(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "proj")
	notesDir := filepath.Join(projectDir, "notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		t.Fatalf("Failed to create notes dir: %v", err)
	}

	noteFile := "2026-01-10-existing.md"
	noteContent := `---
title: Existing Note
date: 2026-01-10
project: proj
tags: []
---

# Existing Note
`
	if err := os.WriteFile(filepath.Join(notesDir, noteFile), []byte(noteContent), 0644); err != nil {
		t.Fatalf("Failed to write note: %v", err)
	}

	// notes.md already contains the link
	notesIndexContent := "# proj Notes\n\n## Notes\n\n- [2026-01-10 Existing Note](notes/" + noteFile + ")\n"
	notesIndexPath := filepath.Join(projectDir, "notes.md")
	if err := os.WriteFile(notesIndexPath, []byte(notesIndexContent), 0644); err != nil {
		t.Fatalf("Failed to write notes.md: %v", err)
	}

	result, err := MigrateNotesIndex(projectDir, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Changed {
		t.Error("Expected Changed=false when all notes already linked")
	}
	if result.Change != "all notes already linked" {
		t.Errorf("Expected 'all notes already linked', got: %q", result.Change)
	}

	// File should be unchanged
	read, _ := os.ReadFile(notesIndexPath)
	if string(read) != notesIndexContent {
		t.Error("notes.md should not have been modified")
	}
}

// ---------------------------------------------------------------------------
// MigratePriorityDueTags tests
// ---------------------------------------------------------------------------

func TestMigratePriorityDueTags_ConvertsHashTags(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	content := `# Tasks

- [ ] Buy groceries #p:2 #due:2026-02-20 <!-- id:abc123 captured:2026-01-30 -->
- [ ] Write tests #p:1 <!-- id:def456 captured:2026-01-31 -->
- [ ] No priority or due date <!-- id:ghi789 captured:2026-02-01 -->
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigratePriorityDueTags(todoPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true")
	}

	// Verify file content
	migrated, err := os.ReadFile(todoPath)
	if err != nil {
		t.Fatalf("Failed to read migrated file: %v", err)
	}
	migratedStr := string(migrated)

	if strings.Contains(migratedStr, "#p:") {
		t.Errorf("Legacy #p: tags should have been converted, got:\n%s", migratedStr)
	}
	if strings.Contains(migratedStr, "#due:") {
		t.Errorf("Legacy #due: tags should have been converted, got:\n%s", migratedStr)
	}
	if !strings.Contains(migratedStr, "p:2") {
		t.Errorf("Expected p:2 (no hash) in migrated content, got:\n%s", migratedStr)
	}
	if !strings.Contains(migratedStr, "due:2026-02-20") {
		t.Errorf("Expected due:2026-02-20 (no hash) in migrated content, got:\n%s", migratedStr)
	}
	if !strings.Contains(migratedStr, "p:1") {
		t.Errorf("Expected p:1 (no hash) in migrated content, got:\n%s", migratedStr)
	}
	// Line with no priority/due should be unchanged
	if !strings.Contains(migratedStr, "- [ ] No priority or due date") {
		t.Error("Line without priority/due should be preserved unchanged")
	}
}

func TestMigratePriorityDueTags_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	// Already in no-hash format
	content := `# Tasks

- [ ] Buy groceries p:2 due:2026-02-20 <!-- id:abc123 captured:2026-01-30 -->
- [ ] Write tests p:1 <!-- id:def456 captured:2026-01-31 -->
`
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigratePriorityDueTags(todoPath, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Changed {
		t.Error("Expected Changed=false for already no-hash format todos")
	}
	if result.Change != "no legacy priority/due tags found" {
		t.Errorf("Expected Change='no legacy priority/due tags found', got: %q", result.Change)
	}

	// File should be unchanged
	read, _ := os.ReadFile(todoPath)
	if string(read) != content {
		t.Error("File should not have been modified (idempotent)")
	}
}

func TestMigratePriorityDueTags_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	todoPath := filepath.Join(tmpDir, "todo.md")

	original := `# Tasks

- [ ] Task with hash tags #p:2 #due:2026-02-20 <!-- id:aabbcc captured:2026-02-10 -->
`
	if err := os.WriteFile(todoPath, []byte(original), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := MigratePriorityDueTags(todoPath, true /* dryRun */)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Changed {
		t.Error("Expected Changed=true in dry run")
	}
	if !strings.Contains(result.Change, "would convert") {
		t.Errorf("Expected Change to mention 'would convert', got: %q", result.Change)
	}

	// File must remain unchanged
	read, _ := os.ReadFile(todoPath)
	if string(read) != original {
		t.Error("Dry run must not modify the file")
	}
}

// ---------------------------------------------------------------------------
// MigrateAllProjects tests
// ---------------------------------------------------------------------------

func TestMigrateAllProjects_Summary(t *testing.T) {
	tmpDir := t.TempDir()
	activeDir := filepath.Join(tmpDir, "01_active")

	// Project 1: has a legacy note and a todo with inline tags
	proj1Dir := filepath.Join(activeDir, "project-one")
	notes1Dir := filepath.Join(proj1Dir, "notes")
	if err := os.MkdirAll(notes1Dir, 0755); err != nil {
		t.Fatalf("Failed to create proj1 notes dir: %v", err)
	}

	legacyNote := `# Legacy Note

Created: 2026-01-05

Body text.
`
	if err := os.WriteFile(filepath.Join(notes1Dir, "2026-01-05-legacy.md"), []byte(legacyNote), 0644); err != nil {
		t.Fatalf("Failed to write legacy note: %v", err)
	}

	todo1 := `# project-one

- [ ] Task one #id:111aaa #captured:2026-01-05
`
	if err := os.WriteFile(filepath.Join(proj1Dir, "todo.md"), []byte(todo1), 0644); err != nil {
		t.Fatalf("Failed to write todo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(proj1Dir, "notes.md"), []byte("# project-one Notes\n\n"), 0644); err != nil {
		t.Fatalf("Failed to write notes.md: %v", err)
	}

	// Project 2: already clean (frontmatter present, no inline tags, notes indexed)
	proj2Dir := filepath.Join(activeDir, "project-two")
	notes2Dir := filepath.Join(proj2Dir, "notes")
	if err := os.MkdirAll(notes2Dir, 0755); err != nil {
		t.Fatalf("Failed to create proj2 notes dir: %v", err)
	}

	cleanNote := `---
title: Clean Note
date: 2026-01-10
project: project-two
tags: []
---

# Clean Note
`
	if err := os.WriteFile(filepath.Join(notes2Dir, "2026-01-10-clean.md"), []byte(cleanNote), 0644); err != nil {
		t.Fatalf("Failed to write clean note: %v", err)
	}
	todo2 := `# project-two

- [ ] No inline tags <!-- id:222bbb captured:2026-01-10 -->
`
	if err := os.WriteFile(filepath.Join(proj2Dir, "todo.md"), []byte(todo2), 0644); err != nil {
		t.Fatalf("Failed to write todo2: %v", err)
	}
	notesIndex2 := "# project-two Notes\n\n## Notes\n\n- [2026-01-10 Clean Note](notes/2026-01-10-clean.md)\n"
	if err := os.WriteFile(filepath.Join(proj2Dir, "notes.md"), []byte(notesIndex2), 0644); err != nil {
		t.Fatalf("Failed to write notes2.md: %v", err)
	}

	// Run migration (apply, not dry-run)
	results, err := MigrateAllProjects(activeDir, false)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 project results, got %d", len(results))
	}

	// Find results by project name
	var proj1Result, proj2Result *ProjectMigrateResult
	for i := range results {
		if results[i].ProjectName == "project-one" {
			proj1Result = &results[i]
		} else if results[i].ProjectName == "project-two" {
			proj2Result = &results[i]
		}
	}

	if proj1Result == nil {
		t.Fatal("Expected result for project-one")
	}
	if proj2Result == nil {
		t.Fatal("Expected result for project-two")
	}

	// project-one should have a changed note (legacy -> frontmatter)
	hasChangedNote := false
	for _, nr := range proj1Result.Notes {
		if nr.Changed {
			hasChangedNote = true
		}
	}
	if !hasChangedNote {
		t.Error("Expected project-one note to be migrated")
	}

	// project-one should have a changed todo (inline tags -> HTML comment)
	hasChangedTodo := false
	for _, tr := range proj1Result.Todos {
		if tr.Changed {
			hasChangedTodo = true
		}
	}
	if !hasChangedTodo {
		t.Error("Expected project-one todo to be migrated")
	}

	// project-one notes.md should have a new link
	hasChangedLink := false
	for _, lr := range proj1Result.Links {
		if lr.Changed {
			hasChangedLink = true
		}
	}
	if !hasChangedLink {
		t.Error("Expected project-one notes.md to have a new link")
	}

	// project-two should have no changes
	for _, nr := range proj2Result.Notes {
		if nr.Changed {
			t.Errorf("Expected no changes in project-two notes, but got: %s", nr.Change)
		}
	}
	for _, tr := range proj2Result.Todos {
		if tr.Changed {
			t.Errorf("Expected no changes in project-two todos, but got: %s", tr.Change)
		}
	}
	for _, lr := range proj2Result.Links {
		if lr.Changed {
			t.Errorf("Expected no changes in project-two links, but got: %s", lr.Change)
		}
	}
}
