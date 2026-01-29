package api

import (
	"fmt"
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

func TestParseTodoFile_WithPriority(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("priority-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "priority-test", "todo.md")

	content := `# Priority Test

- [ ] High priority task #p:1
- [ ] Medium priority task #p:2
- [ ] Low priority task #p:3
- [ ] No priority task
- [x] Completed high priority #p:1
`
	tb.WriteFile(todoFile, content)

	// Parse without completed
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 4 {
		t.Fatalf("Expected 4 open tasks, got %d", len(todos))
	}

	// Verify high priority task
	if todos[0].Content != "High priority task" {
		t.Errorf("Expected content 'High priority task', got '%s'", todos[0].Content)
	}
	if todos[0].Priority == nil || *todos[0].Priority != 1 {
		t.Errorf("Expected priority 1, got %v", formatPriority(todos[0].Priority))
	}

	// Verify medium priority task
	if todos[1].Priority == nil || *todos[1].Priority != 2 {
		t.Errorf("Expected priority 2, got %v", formatPriority(todos[1].Priority))
	}

	// Verify low priority task
	if todos[2].Priority == nil || *todos[2].Priority != 3 {
		t.Errorf("Expected priority 3, got %v", formatPriority(todos[2].Priority))
	}

	// Verify no priority task
	if todos[3].Priority != nil {
		t.Errorf("Expected nil priority, got %v", formatPriority(todos[3].Priority))
	}

	// Parse with completed
	todosAll, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todosAll) != 5 {
		t.Fatalf("Expected 5 total tasks, got %d", len(todosAll))
	}

	// Verify completed task has priority extracted
	completedTask := todosAll[4]
	if completedTask.Status != "done" {
		t.Errorf("Expected status 'done', got '%s'", completedTask.Status)
	}
	if completedTask.Content != "Completed high priority" {
		t.Errorf("Expected content 'Completed high priority', got '%s'", completedTask.Content)
	}
	if completedTask.Priority == nil || *completedTask.Priority != 1 {
		t.Errorf("Expected priority 1, got %v", formatPriority(completedTask.Priority))
	}
}

func TestSetTodoPriority(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("prio-set-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "prio-set-test", "todo.md")

	content := `# Priority Set Test

- [ ] Task without priority
- [ ] Task with priority #p:2
- [x] Completed task
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Test 1: Set priority on task without priority
	task1 := &todos[0]
	priority1 := 1
	err = SetTodoPriority(task1, &priority1)
	if err != nil {
		t.Fatalf("SetTodoPriority failed: %v", err)
	}

	// Verify the file was updated
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "Task without priority #p:1") {
		t.Errorf("Priority not set correctly. File content:\n%s", updatedContent)
	}

	// Test 2: Change priority on task with priority
	task2 := &todos[1]
	priority3 := 3
	err = SetTodoPriority(task2, &priority3)
	if err != nil {
		t.Fatalf("SetTodoPriority failed: %v", err)
	}

	// Verify the priority was changed
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "Task with priority #p:3") {
		t.Errorf("Priority not changed correctly. File content:\n%s", updatedContent)
	}
	// Old priority should be removed
	if strings.Contains(string(updatedContent), "#p:2") {
		t.Errorf("Old priority tag not removed. File content:\n%s", updatedContent)
	}

	// Test 3: Clear priority
	err = SetTodoPriority(task2, nil)
	if err != nil {
		t.Fatalf("SetTodoPriority failed: %v", err)
	}

	// Verify the priority was cleared
	updatedContent, _ = os.ReadFile(todoFile)
	if strings.Contains(string(updatedContent), "#p:3") {
		t.Errorf("Priority not cleared. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "- [ ] Task with priority") {
		t.Errorf("Task content corrupted. File content:\n%s", updatedContent)
	}

	// Test 4: Set priority on completed task
	task3 := &todos[2]
	priority2 := 2
	err = SetTodoPriority(task3, &priority2)
	if err != nil {
		t.Fatalf("SetTodoPriority failed on completed task: %v", err)
	}

	// Verify the priority was set on completed task
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [x] Completed task #p:2") {
		t.Errorf("Priority not set on completed task. File content:\n%s", updatedContent)
	}
}

func TestSetTodoPriority_InvalidPriority(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("invalid-prio")
	todoFile := filepath.Join(tb.ActiveDirPath, "invalid-prio", "todo.md")

	content := `# Test

- [ ] Task 1
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	task := &todos[0]

	// Test invalid priority values
	invalid := []int{0, 4, 5, -1, 10}
	for _, p := range invalid {
		prio := p
		err := SetTodoPriority(task, &prio)
		if err == nil {
			t.Errorf("Expected error for invalid priority %d, got nil", p)
		}
	}
}

func TestSetTodoPriority_InvalidLineNumber(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("invalid-line")
	todoFile := filepath.Join(tb.ActiveDirPath, "invalid-line", "todo.md")

	content := `# Test

- [ ] Task 1
`
	tb.WriteFile(todoFile, content)

	// Create a TodoItem with invalid line number
	task := &TodoItem{
		File:    todoFile,
		Line:    999,
		Content: "Fake task",
	}

	priority := 1
	err := SetTodoPriority(task, &priority)
	if err == nil {
		t.Error("Expected error for invalid line number, got nil")
	}
}

func TestParseTodoFile_WithStates(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("state-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "state-test", "todo.md")

	content := `# State Test

- [ ] Open task
- [>] In progress task
- [-] Blocked task
- [x] Done task
`
	tb.WriteFile(todoFile, content)

	// Parse without completed
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 3 {
		t.Fatalf("Expected 3 open/in-progress/blocked tasks, got %d", len(todos))
	}

	// Verify statuses
	expectedStatuses := []string{"open", "in-progress", "blocked"}
	for i, expected := range expectedStatuses {
		if todos[i].Status != expected {
			t.Errorf("Task %d: expected status '%s', got '%s'", i, expected, todos[i].Status)
		}
	}

	// Parse with completed
	todosAll, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todosAll) != 4 {
		t.Fatalf("Expected 4 total tasks, got %d", len(todosAll))
	}

	// Verify done status
	if todosAll[3].Status != "done" {
		t.Errorf("Expected done status, got '%s'", todosAll[3].Status)
	}
}

func TestSetTodoStatus(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("status-set-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "status-set-test", "todo.md")

	content := `# Status Set Test

- [ ] Task 1
- [>] Task 2
- [-] Task 3
- [x] Task 4
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Test 1: Change open to in-progress
	task1 := &todos[0]
	err = SetTodoStatus(task1, "in-progress")
	if err != nil {
		t.Fatalf("SetTodoStatus failed: %v", err)
	}

	// Verify the file was updated
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [>] Task 1") {
		t.Errorf("Status not set correctly. File content:\n%s", updatedContent)
	}

	// Test 2: Change in-progress to blocked
	task2 := &todos[1]
	err = SetTodoStatus(task2, "blocked")
	if err != nil {
		t.Fatalf("SetTodoStatus failed: %v", err)
	}

	// Verify the status was changed
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [-] Task 2") {
		t.Errorf("Status not changed correctly. File content:\n%s", updatedContent)
	}

	// Test 3: Change blocked to open
	task3 := &todos[2]
	err = SetTodoStatus(task3, "open")
	if err != nil {
		t.Fatalf("SetTodoStatus failed: %v", err)
	}

	// Verify the status was changed
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [ ] Task 3") {
		t.Errorf("Status not changed correctly. File content:\n%s", updatedContent)
	}

	// Test 4: Change done to open
	task4 := &todos[3]
	err = SetTodoStatus(task4, "open")
	if err != nil {
		t.Fatalf("SetTodoStatus failed on completed task: %v", err)
	}

	// Verify the status was changed
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [ ] Task 4") {
		t.Errorf("Status not changed correctly. File content:\n%s", updatedContent)
	}
}

func TestSetTodoStatus_InvalidStatus(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("invalid-status")
	todoFile := filepath.Join(tb.ActiveDirPath, "invalid-status", "todo.md")

	content := `# Test

- [ ] Task 1
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	task := &todos[0]

	// Test invalid status values
	invalid := []string{"invalid", "pending", "completed", ""}
	for _, s := range invalid {
		err := SetTodoStatus(task, s)
		if err == nil {
			t.Errorf("Expected error for invalid status %s, got nil", s)
		}
	}
}

func TestSetTodoStatus_PreservesContent(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("preserve-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "preserve-test", "todo.md")

	content := `# Test

- [ ] Task with #p:1 and other metadata #captured:2024-01-21
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	task := &todos[0]
	err = SetTodoStatus(task, "in-progress")
	if err != nil {
		t.Fatalf("SetTodoStatus failed: %v", err)
	}

	// Verify metadata is preserved
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "#p:1") {
		t.Errorf("Priority tag not preserved. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "#captured:2024-01-21") {
		t.Errorf("Captured tag not preserved. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "- [>]") {
		t.Errorf("Status not set correctly. File content:\n%s", updatedContent)
	}
}

func TestParseTodoFile_WithDueDate(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("duedate-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "duedate-test", "todo.md")

	content := `# Due Date Test

- [ ] Task with due date #due:2026-02-15
- [ ] Task with priority and due date #p:1 #due:2026-03-01
- [ ] Task without due date
- [x] Completed task with due date #due:2026-01-20
`
	tb.WriteFile(todoFile, content)

	// Parse without completed
	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 3 {
		t.Fatalf("Expected 3 open tasks, got %d", len(todos))
	}

	// Verify first task has due date
	if todos[0].Content != "Task with due date" {
		t.Errorf("Expected content 'Task with due date', got '%s'", todos[0].Content)
	}
	if todos[0].DueDate != "2026-02-15" {
		t.Errorf("Expected due date '2026-02-15', got '%s'", todos[0].DueDate)
	}

	// Verify second task has both priority and due date
	if todos[1].Content != "Task with priority and due date" {
		t.Errorf("Expected content 'Task with priority and due date', got '%s'", todos[1].Content)
	}
	if todos[1].Priority == nil || *todos[1].Priority != 1 {
		t.Errorf("Expected priority 1, got %v", formatPriority(todos[1].Priority))
	}
	if todos[1].DueDate != "2026-03-01" {
		t.Errorf("Expected due date '2026-03-01', got '%s'", todos[1].DueDate)
	}

	// Verify third task has no due date
	if todos[2].DueDate != "" {
		t.Errorf("Expected no due date, got '%s'", todos[2].DueDate)
	}

	// Parse with completed
	todosAll, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Verify completed task has due date
	if todosAll[3].DueDate != "2026-01-20" {
		t.Errorf("Expected due date '2026-01-20', got '%s'", todosAll[3].DueDate)
	}
}

func TestSetTodoDueDate(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("duedate-set-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "duedate-set-test", "todo.md")

	content := `# Due Date Set Test

- [ ] Task without due date
- [ ] Task with due date #due:2026-02-15
- [x] Completed task
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, true)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Test 1: Set due date on task without due date
	task1 := &todos[0]
	err = SetTodoDueDate(task1, "2026-03-01")
	if err != nil {
		t.Fatalf("SetTodoDueDate failed: %v", err)
	}

	// Verify the file was updated
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "Task without due date #due:2026-03-01") {
		t.Errorf("Due date not set correctly. File content:\n%s", updatedContent)
	}

	// Test 2: Change due date on task with due date
	task2 := &todos[1]
	err = SetTodoDueDate(task2, "2026-04-15")
	if err != nil {
		t.Fatalf("SetTodoDueDate failed: %v", err)
	}

	// Verify the due date was changed
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "Task with due date #due:2026-04-15") {
		t.Errorf("Due date not changed correctly. File content:\n%s", updatedContent)
	}
	// Old due date should be removed
	if strings.Contains(string(updatedContent), "#due:2026-02-15") {
		t.Errorf("Old due date tag not removed. File content:\n%s", updatedContent)
	}

	// Test 3: Clear due date
	err = SetTodoDueDate(task2, "clear")
	if err != nil {
		t.Fatalf("SetTodoDueDate failed: %v", err)
	}

	// Verify the due date was cleared
	updatedContent, _ = os.ReadFile(todoFile)
	if strings.Contains(string(updatedContent), "#due:2026-04-15") {
		t.Errorf("Due date not cleared. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "- [ ] Task with due date") {
		t.Errorf("Task content corrupted. File content:\n%s", updatedContent)
	}

	// Test 4: Set due date on completed task
	task3 := &todos[2]
	err = SetTodoDueDate(task3, "2026-05-01")
	if err != nil {
		t.Fatalf("SetTodoDueDate failed on completed task: %v", err)
	}

	// Verify the due date was set on completed task
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "- [x] Completed task #due:2026-05-01") {
		t.Errorf("Due date not set on completed task. File content:\n%s", updatedContent)
	}
}

func TestSetTodoDueDate_InvalidDate(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("invalid-date")
	todoFile := filepath.Join(tb.ActiveDirPath, "invalid-date", "todo.md")

	content := `# Test

- [ ] Task 1
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	task := &todos[0]

	// Test invalid date formats
	invalid := []string{"2026-13-01", "not-a-date", "01-20-2026", "2026/02/15"}
	for _, d := range invalid {
		err := SetTodoDueDate(task, d)
		if err == nil {
			t.Errorf("Expected error for invalid date %s, got nil", d)
		}
	}
}

func TestSetTodoDueDate_PreservesOtherMetadata(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("preserve-metadata")
	todoFile := filepath.Join(tb.ActiveDirPath, "preserve-metadata", "todo.md")

	content := `# Test

- [ ] Task with #p:1 and other metadata #captured:2024-01-21
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	task := &todos[0]
	err = SetTodoDueDate(task, "2026-02-15")
	if err != nil {
		t.Fatalf("SetTodoDueDate failed: %v", err)
	}

	// Verify other metadata is preserved
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "#p:1") {
		t.Errorf("Priority tag not preserved. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "#captured:2024-01-21") {
		t.Errorf("Captured tag not preserved. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "#due:2026-02-15") {
		t.Errorf("Due date not set correctly. File content:\n%s", updatedContent)
	}
}

func TestParseTodoFile_WithTags(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("tags-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "tags-test", "todo.md")

	content := `# Tags Test

- [ ] Fix bug #bug #security
- [ ] Add feature #feature #ui #frontend
- [ ] Task without tags
- [ ] Task with tags and metadata #p:1 #bug #due:2026-02-15 #feature
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	if len(todos) != 4 {
		t.Fatalf("Expected 4 tasks, got %d", len(todos))
	}

	// Verify first task has tags
	if todos[0].Content != "Fix bug" {
		t.Errorf("Expected content 'Fix bug', got '%s'", todos[0].Content)
	}
	expectedTags := []string{"bug", "security"}
	if !equalStringSlices(todos[0].Tags, expectedTags) {
		t.Errorf("Expected tags %v, got %v", expectedTags, todos[0].Tags)
	}

	// Verify second task has multiple tags
	expectedTags2 := []string{"feature", "ui", "frontend"}
	if !equalStringSlices(todos[1].Tags, expectedTags2) {
		t.Errorf("Expected tags %v, got %v", expectedTags2, todos[1].Tags)
	}

	// Verify third task has no tags
	if len(todos[2].Tags) != 0 {
		t.Errorf("Expected no tags, got %v", todos[2].Tags)
	}

	// Verify fourth task has tags extracted separately from metadata
	if todos[3].Priority == nil || *todos[3].Priority != 1 {
		t.Errorf("Expected priority 1, got %v", formatPriority(todos[3].Priority))
	}
	if todos[3].DueDate != "2026-02-15" {
		t.Errorf("Expected due date '2026-02-15', got '%s'", todos[3].DueDate)
	}
	expectedTags3 := []string{"bug", "feature"}
	if !equalStringSlices(todos[3].Tags, expectedTags3) {
		t.Errorf("Expected tags %v, got %v", expectedTags3, todos[3].Tags)
	}
}

func TestAddTodoTags(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("add-tags-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "add-tags-test", "todo.md")

	content := `# Add Tags Test

- [ ] Task without tags
- [ ] Task with tags #bug
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Test 1: Add tags to task without tags
	task1 := &todos[0]
	err = AddTodoTags(task1, []string{"feature", "ui"})
	if err != nil {
		t.Fatalf("AddTodoTags failed: %v", err)
	}

	// Verify the file was updated
	updatedContent, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "Task without tags #feature #ui") {
		t.Errorf("Tags not added correctly. File content:\n%s", updatedContent)
	}

	// Test 2: Add tags to task with existing tags
	task2 := &todos[1]
	err = AddTodoTags(task2, []string{"security", "urgent"})
	if err != nil {
		t.Fatalf("AddTodoTags failed: %v", err)
	}

	// Verify the tags were added
	updatedContent, _ = os.ReadFile(todoFile)
	if !strings.Contains(string(updatedContent), "#security #urgent") {
		t.Errorf("Tags not added correctly. File content:\n%s", updatedContent)
	}

	// Test 3: Try to add duplicate tag (should not add)
	todos2, _ := ParseAllTodos(tb.ActiveDirPath, false)
	task3 := &todos2[1]
	err = AddTodoTags(task3, []string{"bug"}) // Already has #bug
	if err != nil {
		t.Fatalf("AddTodoTags failed: %v", err)
	}

	// Count occurrences of #bug (should only be one)
	updatedContent, _ = os.ReadFile(todoFile)
	bugCount := strings.Count(string(updatedContent), "#bug")
	if bugCount != 1 {
		t.Errorf("Expected 1 #bug tag, found %d. File content:\n%s", bugCount, updatedContent)
	}
}

func TestRemoveTodoTags(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddProject("remove-tags-test")
	todoFile := filepath.Join(tb.ActiveDirPath, "remove-tags-test", "todo.md")

	content := `# Remove Tags Test

- [ ] Task with multiple tags #bug #security #urgent
- [ ] Task with one tag #feature
`
	tb.WriteFile(todoFile, content)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	// Test 1: Remove one tag from multiple
	task1 := &todos[0]
	err = RemoveTodoTags(task1, []string{"security"})
	if err != nil {
		t.Fatalf("RemoveTodoTags failed: %v", err)
	}

	// Verify the tag was removed
	updatedContent, _ := os.ReadFile(todoFile)
	if strings.Contains(string(updatedContent), "#security") {
		t.Errorf("Tag not removed. File content:\n%s", updatedContent)
	}
	if !strings.Contains(string(updatedContent), "#bug") || !strings.Contains(string(updatedContent), "#urgent") {
		t.Errorf("Other tags were removed incorrectly. File content:\n%s", updatedContent)
	}

	// Test 2: Remove all remaining tags
	todos2, _ := ParseAllTodos(tb.ActiveDirPath, false)
	task1Updated := &todos2[0]
	err = RemoveTodoTags(task1Updated, []string{"bug", "urgent"})
	if err != nil {
		t.Fatalf("RemoveTodoTags failed: %v", err)
	}

	// Verify all tags were removed
	updatedContent, _ = os.ReadFile(todoFile)
	firstLine := strings.Split(string(updatedContent), "\n")[2]
	if strings.Contains(firstLine, "#bug") || strings.Contains(firstLine, "#urgent") {
		t.Errorf("Tags not fully removed. Line: %s", firstLine)
	}
}

func TestListAllTags(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create multiple projects with tags
	tb.AddProject("project-a")
	todoFileA := filepath.Join(tb.ActiveDirPath, "project-a", "todo.md")
	contentA := `# Project A
- [ ] Task 1 #bug #security
- [ ] Task 2 #bug #feature
`
	tb.WriteFile(todoFileA, contentA)

	tb.AddProject("project-b")
	todoFileB := filepath.Join(tb.ActiveDirPath, "project-b", "todo.md")
	contentB := `# Project B
- [ ] Task 3 #feature #ui
- [ ] Task 4 #bug
`
	tb.WriteFile(todoFileB, contentB)

	todos, err := ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("ParseAllTodos failed: %v", err)
	}

	tagCounts := ListAllTags(todos)

	// Verify counts
	expectedCounts := map[string]int{
		"bug":      3,
		"security": 1,
		"feature":  2,
		"ui":       1,
	}

	for tag, expectedCount := range expectedCounts {
		if count, ok := tagCounts[tag]; !ok {
			t.Errorf("Expected tag '%s' not found in counts", tag)
		} else if count != expectedCount {
			t.Errorf("Expected count %d for tag '%s', got %d", expectedCount, tag, count)
		}
	}
}

// Helper function to format priority for test output
func formatPriority(p *int) string {
	if p == nil {
		return "nil"
	}
	return fmt.Sprintf("%d", *p)
}

// Helper function to compare string slices (order doesn't matter)
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]bool)
	for _, s := range a {
		aMap[s] = true
	}
	for _, s := range b {
		if !aMap[s] {
			return false
		}
	}
	return true
}
