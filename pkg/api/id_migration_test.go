package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMigrateTodoFileIDs(t *testing.T) {
	tmpDir := t.TempDir()
	todoFile := filepath.Join(tmpDir, "todo.md")

	// Create a todo file with some todos without IDs
	content := `# Tasks

## Active

- [ ] First task without ID
- [ ] Second task without ID #p:1
- [ ] Third task already has ID #id:abc123

## Completed

- [x] Completed task without ID
`

	if err := os.WriteFile(todoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run migration
	count, err := MigrateTodoFileIDs(todoFile)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 todos to be migrated, got %d", count)
	}

	// Read the migrated file
	migratedContent, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read migrated file: %v", err)
	}

	migratedStr := string(migratedContent)

	// Check that all todos now have IDs
	lines := strings.Split(migratedStr, "\n")
	var todoLines []string
	for _, line := range lines {
		if strings.Contains(line, "- [ ]") || strings.Contains(line, "- [x]") {
			todoLines = append(todoLines, line)
		}
	}

	if len(todoLines) != 4 {
		t.Errorf("Expected 4 todo lines, found %d", len(todoLines))
	}

	for i, line := range todoLines {
		if !strings.Contains(line, "#id:") {
			t.Errorf("Todo line %d missing ID: %s", i, line)
		}
	}
}

func TestMigrateTodoFileIDs_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	todoFile := filepath.Join(tmpDir, "todo.md")

	content := `# Tasks

## Active

- [ ] Task one
- [ ] Task two
`

	if err := os.WriteFile(todoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// First migration
	count1, err := MigrateTodoFileIDs(todoFile)
	if err != nil {
		t.Fatalf("First migration failed: %v", err)
	}

	if count1 != 2 {
		t.Errorf("First migration: expected 2 todos migrated, got %d", count1)
	}

	// Read the file after first migration
	content1, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read after first migration: %v", err)
	}

	// Second migration (should be idempotent)
	count2, err := MigrateTodoFileIDs(todoFile)
	if err != nil {
		t.Fatalf("Second migration failed: %v", err)
	}

	if count2 != 0 {
		t.Errorf("Second migration (idempotent): expected 0 todos migrated, got %d", count2)
	}

	// Content should be identical
	content2, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read after second migration: %v", err)
	}

	if string(content1) != string(content2) {
		t.Errorf("Second migration changed content (not idempotent)")
	}
}

func TestMigrateDumpFileIDs(t *testing.T) {
	tmpDir := t.TempDir()
	dumpFile := filepath.Join(tmpDir, "00_dump.md")

	// Create a dump file with tasks and notes without IDs
	content := `# Dump

- [ ] Task one without ID
- [ ] Task two without ID

[Note] Note one without ID
    Content of note one

[Note] Note two without ID
    Content of note two
`

	if err := os.WriteFile(dumpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Run migration
	count, err := MigrateDumpFileIDs(dumpFile)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected 4 items to be migrated, got %d", count)
	}

	// Verify all top-level items have IDs
	migratedContent, err := os.ReadFile(dumpFile)
	if err != nil {
		t.Fatalf("Failed to read migrated file: %v", err)
	}

	migratedStr := string(migratedContent)
	if strings.Count(migratedStr, "#id:") < 4 {
		t.Errorf("Expected at least 4 IDs in migrated dump")
	}
}

func TestMigrateAllProjectTodos(t *testing.T) {
	tmpDir := t.TempDir()
	activeDir := tmpDir

	// Create two projects with todos
	project1Dir := filepath.Join(activeDir, "project1")
	project2Dir := filepath.Join(activeDir, "project2")

	if err := os.MkdirAll(project1Dir, 0755); err != nil {
		t.Fatalf("Failed to create project1: %v", err)
	}
	if err := os.MkdirAll(project2Dir, 0755); err != nil {
		t.Fatalf("Failed to create project2: %v", err)
	}

	// Create todo files
	todo1Content := `# Project 1 Tasks

- [ ] P1 Task 1
- [ ] P1 Task 2
`

	todo2Content := `# Project 2 Tasks

- [ ] P2 Task 1
- [ ] P2 Task 2
- [ ] P2 Task 3
`

	if err := os.WriteFile(filepath.Join(project1Dir, "todo.md"), []byte(todo1Content), 0644); err != nil {
		t.Fatalf("Failed to create project1 todo: %v", err)
	}
	if err := os.WriteFile(filepath.Join(project2Dir, "todo.md"), []byte(todo2Content), 0644); err != nil {
		t.Fatalf("Failed to create project2 todo: %v", err)
	}

	// Run migration
	results, err := MigrateAllProjectTodos(activeDir)
	if err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 projects migrated, got %d", len(results))
	}

	if results["project1"] != 2 {
		t.Errorf("Expected project1: 2 todos, got %d", results["project1"])
	}

	if results["project2"] != 3 {
		t.Errorf("Expected project2: 3 todos, got %d", results["project2"])
	}
}
