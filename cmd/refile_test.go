package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/markdown"
)

func TestListProjects(t *testing.T) {
	// Create temporary directory with projects
	tmpDir := t.TempDir()

	// Create projects in non-alphabetical order
	projectNames := []string{"zebra", "alpha", "charlie", "bravo"}
	for _, name := range projectNames {
		projectDir := filepath.Join(tmpDir, name)
		if err := os.Mkdir(projectDir, 0755); err != nil {
			t.Fatalf("Failed to create project dir: %v", err)
		}
	}

	// List projects
	projects, err := listProjects(tmpDir)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	// Verify count
	if len(projects) != 4 {
		t.Fatalf("Expected 4 projects, got %d", len(projects))
	}

	// Verify alphabetical order
	expected := []string{"alpha", "bravo", "charlie", "zebra"}
	for i, name := range expected {
		if projects[i] != name {
			t.Errorf("Index %d: expected '%s', got '%s'", i, name, projects[i])
		}
	}
}

func TestListProjects_HiddenDirectories(t *testing.T) {
	// Create temporary directory with projects
	tmpDir := t.TempDir()

	// Create visible and hidden directories
	visibleProjects := []string{"project1", "project2"}
	hiddenDirs := []string{".hidden", ".git"}

	for _, name := range visibleProjects {
		if err := os.Mkdir(filepath.Join(tmpDir, name), 0755); err != nil {
			t.Fatalf("Failed to create visible project: %v", err)
		}
	}

	for _, name := range hiddenDirs {
		if err := os.Mkdir(filepath.Join(tmpDir, name), 0755); err != nil {
			t.Fatalf("Failed to create hidden dir: %v", err)
		}
	}

	// List projects
	projects, err := listProjects(tmpDir)
	if err != nil {
		t.Fatalf("listProjects failed: %v", err)
	}

	// Verify only visible projects are returned
	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(projects))
	}

	// Verify hidden dirs are not included
	for _, project := range projects {
		if strings.HasPrefix(project, ".") {
			t.Errorf("Hidden directory '%s' should not be in results", project)
		}
	}
}

func TestRefileTask_ActiveSection(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create todo.md with Active and Completed sections
	todoFile := filepath.Join(projectDir, "todo.md")
	initialContent := `# Tasks

## Active

- [ ] Existing task 1
- [ ] Existing task 2

## Completed

- [x] Done task
`
	if err := os.WriteFile(todoFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create todo.md: %v", err)
	}

	// Create a dump item to refile
	item := &markdown.DumpItem{
		Type:    markdown.ItemTypeTodo,
		Content: "New refiled task #captured:2026-01-30",
	}

	// Refile the task
	if err := refileTask(item, projectDir); err != nil {
		t.Fatalf("refileTask failed: %v", err)
	}

	// Read the updated file
	content, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read todo.md: %v", err)
	}

	lines := strings.Split(string(content), "\n")

	// Find the Active section and verify new task is inserted there
	activeIdx := -1
	newTaskIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "## Active" {
			activeIdx = i
		}
		if strings.Contains(line, "New refiled task") {
			newTaskIdx = i
		}
	}

	if activeIdx == -1 {
		t.Fatal("Active section not found")
	}

	if newTaskIdx == -1 {
		t.Fatal("New task not found in todo.md")
	}

	// Verify new task is after Active section but before Completed section
	completedIdx := -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "## Completed" {
			completedIdx = i
			break
		}
	}

	if newTaskIdx <= activeIdx {
		t.Errorf("New task (line %d) should be after Active section (line %d)", newTaskIdx, activeIdx)
	}

	if completedIdx != -1 && newTaskIdx >= completedIdx {
		t.Errorf("New task (line %d) should be before Completed section (line %d)", newTaskIdx, completedIdx)
	}

	// Verify task content is correct (should include ID tag now)
	taskLine := strings.TrimSpace(lines[newTaskIdx])
	if !strings.Contains(taskLine, "- [~] New refiled task #captured:2026-01-30") {
		t.Errorf("Task should contain '- [~] New refiled task #captured:2026-01-30', got '%s'", taskLine)
	}
	// Verify ID tag was added
	if !strings.Contains(taskLine, "#id:") {
		t.Errorf("Task should have #id: tag, got '%s'", taskLine)
	}
}

func TestRefileTask_NoActiveSection(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create todo.md without sections (edge case)
	todoFile := filepath.Join(projectDir, "todo.md")
	initialContent := `# Tasks

- [ ] Existing task
`
	if err := os.WriteFile(todoFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create todo.md: %v", err)
	}

	// Create a dump item to refile
	item := &markdown.DumpItem{
		Type:    markdown.ItemTypeTodo,
		Content: "New task",
	}

	// Refile the task (should fallback to append)
	if err := refileTask(item, projectDir); err != nil {
		t.Fatalf("refileTask failed: %v", err)
	}

	// Read the updated file
	content, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read todo.md: %v", err)
	}

	// Verify task was added (even without Active section)
	if !strings.Contains(string(content), "New task") {
		t.Error("New task should be added to file")
	}
}

func TestRefileTask_CreatesTodoFile(t *testing.T) {
	// Create temporary project directory
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "test-project")
	if err := os.Mkdir(projectDir, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Don't create todo.md - let refileTask create it

	// Create a dump item to refile
	item := &markdown.DumpItem{
		Type:    markdown.ItemTypeTodo,
		Content: "First task",
	}

	// Refile the task
	if err := refileTask(item, projectDir); err != nil {
		t.Fatalf("refileTask failed: %v", err)
	}

	// Verify todo.md was created
	todoFile := filepath.Join(projectDir, "todo.md")
	if _, err := os.Stat(todoFile); os.IsNotExist(err) {
		t.Fatal("todo.md should be created")
	}

	// Verify it has the correct structure
	content, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read todo.md: %v", err)
	}

	contentStr := string(content)

	// Should have sections
	if !strings.Contains(contentStr, "## Active") {
		t.Error("Should contain Active section")
	}

	if !strings.Contains(contentStr, "## Completed") {
		t.Error("Should contain Completed section")
	}

	// Should have the task
	if !strings.Contains(contentStr, "First task") {
		t.Error("Should contain the refiled task")
	}
}
