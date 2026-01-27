package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestParseAllTodos(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create a project with todos
	tb.AddProject("test-project")
	todoFile := filepath.Join(tb.ActiveDirPath, "test-project", "todo.md")

	todoContent := `# Test Project

## Active

- [ ] Task 1
- [ ] Task 2 with details
- [x] Completed task

## Completed

- [x] Old completed task
`
	tb.WriteFile(todoFile, todoContent)

	// Parse todos (excluding completed)
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Should find 2 open tasks
	if len(todos) != 2 {
		t.Fatalf("Expected 2 open tasks, got %d", len(todos))
	}

	// Verify first task
	if todos[0].Content != "Task 1" {
		t.Errorf("Expected content 'Task 1', got '%s'", todos[0].Content)
	}
	if todos[0].Status != "open" {
		t.Errorf("Expected status 'open', got '%s'", todos[0].Status)
	}
	if todos[0].Project != "test-project" {
		t.Errorf("Expected project 'test-project', got '%s'", todos[0].Project)
	}
	if todos[0].Line != 5 {
		t.Errorf("Expected line 5, got %d", todos[0].Line)
	}

	// Verify ID is generated
	if todos[0].ID == "" {
		t.Error("Task ID is empty")
	}
	if len(todos[0].ID) != 6 {
		t.Errorf("Expected 6-char ID, got %d chars", len(todos[0].ID))
	}
}

func TestParseAllTodos_IncludeCompleted(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("project-a")
	todoFile := filepath.Join(tb.ActiveDirPath, "project-a", "todo.md")

	todoContent := `# Project A

## Active

- [ ] Open task
- [x] Done task

## Completed

- [x] Completed task
`
	tb.WriteFile(todoFile, todoContent)

	// Parse with completed tasks
	todos, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Should find 1 open + 2 completed = 3 tasks
	if len(todos) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(todos))
	}

	// Count statuses
	openCount := 0
	doneCount := 0
	for _, todo := range todos {
		if todo.Status == "open" {
			openCount++
		} else if todo.Status == "done" {
			doneCount++
		}
	}

	if openCount != 1 {
		t.Errorf("Expected 1 open task, got %d", openCount)
	}
	if doneCount != 2 {
		t.Errorf("Expected 2 done tasks, got %d", doneCount)
	}
}

func TestParseAllTodos_MultipleProjects(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create multiple projects
	projects := []string{"project-a", "project-b", "project-c"}
	for _, proj := range projects {
		tb.AddProject(proj)
		todoFile := filepath.Join(tb.ActiveDirPath, proj, "todo.md")
		content := `# ` + proj + `

## Active

- [ ] Task from ` + proj + `
`
		tb.WriteFile(todoFile, content)
	}

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Should find 3 tasks (one per project)
	if len(todos) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(todos))
	}

	// Verify each project is represented
	projectsSeen := make(map[string]bool)
	for _, todo := range todos {
		projectsSeen[todo.Project] = true
	}

	for _, proj := range projects {
		if !projectsSeen[proj] {
			t.Errorf("Project '%s' not found in todos", proj)
		}
	}
}

func TestParseAllTodos_NoProjects(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(todos))
	}
}

func TestParseAllTodos_NonExistentDir(t *testing.T) {
	tmpDir := os.TempDir()
	nonExistentDir := filepath.Join(tmpDir, "nonexistent-dir-12345")

	_, err := ParseAllTodos(nonExistentDir, false)
	if err == nil {
		t.Error("Expected error for non-existent directory")
	}
}

func TestParseAllTodos_SkipsHiddenDirs(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create a hidden directory
	hiddenDir := filepath.Join(tb.ActiveDirPath, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden dir: %v", err)
	}

	todoFile := filepath.Join(hiddenDir, "todo.md")
	content := `# Hidden

- [ ] Should not be found
`
	tb.WriteFile(todoFile, content)

	// Also create a normal project
	tb.AddProject("visible")
	visibleTodo := filepath.Join(tb.ActiveDirPath, "visible", "todo.md")
	visibleContent := `# Visible

- [ ] Should be found
`
	tb.WriteFile(visibleTodo, visibleContent)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Should only find the visible project's task
	if len(todos) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(todos))
	}

	if todos[0].Project != "visible" {
		t.Errorf("Expected project 'visible', got '%s'", todos[0].Project)
	}
}

func TestParseAllTodos_MissingTodoFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create project without todo.md
	projDir := filepath.Join(tb.ActiveDirPath, "no-todo")
	if err := os.MkdirAll(projDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Should not error, just return empty
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(todos))
	}
}

func TestParseTodoFile_IndentedTasks(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("indented")
	todoFile := filepath.Join(tb.ActiveDirPath, "indented", "todo.md")

	content := `# Project

  - [ ] Indented task
    - [ ] More indented
- [ ] Not indented
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// All should be found (regex allows leading whitespace)
	if len(todos) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(todos))
	}
}

func TestParseTodoFile_VariousCheckboxFormats(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("formats")
	todoFile := filepath.Join(tb.ActiveDirPath, "formats", "todo.md")

	content := `# Test

- [ ] lowercase x
- [x] lowercase x done
- [X] uppercase X done
- [] empty (should not match)
- [o] other char (should not match)
`
	tb.WriteFile(todoFile, content)

	// Without completed
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 1 {
		t.Fatalf("Expected 1 open task, got %d", len(todos))
	}

	// With completed
	todosAll, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todosAll) != 3 {
		t.Fatalf("Expected 3 total tasks (1 open + 2 done), got %d", len(todosAll))
	}
}

func TestTodoItem_IDConsistency(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("consistency")
	todoFile := filepath.Join(tb.ActiveDirPath, "consistency", "todo.md")

	content := `# Test

- [ ] Task 1
- [ ] Task 2
`
	tb.WriteFile(todoFile, content)

	// Parse twice
	todos1, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("First parse failed: %v", err)
	}

	todos2, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// IDs should be the same (deterministic)
	if len(todos1) != len(todos2) {
		t.Fatalf("Length mismatch: %d vs %d", len(todos1), len(todos2))
	}

	for i := range todos1 {
		if todos1[i].ID != todos2[i].ID {
			t.Errorf("Task %d ID mismatch: %s vs %s", i, todos1[i].ID, todos2[i].ID)
		}
	}
}

func TestTodoItem_UniqueIDs(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("unique")
	todoFile := filepath.Join(tb.ActiveDirPath, "unique", "todo.md")

	content := `# Test

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// IDs should be unique
	seen := make(map[string]bool)
	for _, todo := range todos {
		if seen[todo.ID] {
			t.Errorf("Duplicate ID found: %s", todo.ID)
		}
		seen[todo.ID] = true
	}
}

func TestFindTodoByID(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("find-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "find-test", "todo.md")

	content := `# Test

- [ ] Task 1
- [ ] Task 2
- [ ] Task 3
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Find by ID
	if len(todos) < 2 {
		t.Fatalf("Expected at least 2 todos, got %d", len(todos))
	}

	targetID := todos[1].ID
	found := FindTodoByID(todos, targetID)

	if found == nil {
		t.Fatal("Todo not found by ID")
	}

	if found.ID != targetID {
		t.Errorf("Found wrong todo: %s != %s", found.ID, targetID)
	}

	if !strings.Contains(found.Content, "Task 2") {
		t.Errorf("Found todo has wrong content: %s", found.Content)
	}
}

func TestFindTodoByID_NotFound(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("find-test2")
	todoFile := filepath.Join(tb.ActiveDirPath, "find-test2", "todo.md")

	content := `# Test

- [ ] Task 1
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	found := FindTodoByID(todos, "nonexistent-id")
	if found != nil {
		t.Error("Expected nil for non-existent ID")
	}
}

func TestFindTodoByPattern(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("pattern-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "pattern-test", "todo.md")

	content := `# Test

- [ ] Fix authentication bug
- [ ] Update documentation
- [ ] Fix login bug
- [ ] Add tests
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Find todos with "bug"
	matches := FindTodoByPattern(todos, "bug")
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches for 'bug', got %d", len(matches))
	}

	// Should match case-insensitively
	for _, match := range matches {
		if !strings.Contains(strings.ToLower(match.Content), "bug") {
			t.Errorf("Match doesn't contain 'bug': %s", match.Content)
		}
	}

	// Find todos with "Fix"
	matches = FindTodoByPattern(todos, "Fix")
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches for 'Fix', got %d", len(matches))
	}

	// Find todos with pattern that doesn't exist
	matches = FindTodoByPattern(todos, "nonexistent")
	if len(matches) != 0 {
		t.Errorf("Expected 0 matches for 'nonexistent', got %d", len(matches))
	}
}
