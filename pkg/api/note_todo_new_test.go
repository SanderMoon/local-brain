package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

// Phase 5: Note Management Tests

func TestCreateNoteFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	filePath, err := CreateNoteFile(projectPath, "Test Note", "Note content here", "2024-01-15")
	if err != nil {
		t.Fatalf("CreateNoteFile failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Note file not created: %v", err)
	}

	// Verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read note file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# Test Note") {
		t.Error("Title not found in note")
	}
	if !strings.Contains(contentStr, "date: 2024-01-15") {
		t.Error("Created date not found in note")
	}
	if !strings.Contains(contentStr, "Note content here") {
		t.Error("Content not found in note")
	}
}

func TestCreateNoteFile_DuplicateNames(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Create two notes with same title and date
	file1, err := CreateNoteFile(projectPath, "Same Title", "Content 1", "2024-01-15")
	if err != nil {
		t.Fatalf("First CreateNoteFile failed: %v", err)
	}

	file2, err := CreateNoteFile(projectPath, "Same Title", "Content 2", "2024-01-15")
	if err != nil {
		t.Fatalf("Second CreateNoteFile failed: %v", err)
	}

	// Should have different paths
	if file1 == file2 {
		t.Error("Duplicate filenames created")
	}

	// Both should exist
	if _, err := os.Stat(file1); err != nil {
		t.Error("First file not created")
	}
	if _, err := os.Stat(file2); err != nil {
		t.Error("Second file not created")
	}
}

func TestReadNoteFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Create a note
	expectedContent := "Test content for reading"
	filePath, _ := CreateNoteFile(projectPath, "Test", expectedContent, "2024-01-15")

	// Read it back
	content, err := ReadNoteFile(filePath)
	if err != nil {
		t.Fatalf("ReadNoteFile failed: %v", err)
	}

	if !strings.Contains(content, expectedContent) {
		t.Errorf("Expected content not found. Got:\n%s", content)
	}
}

func TestReadNoteFile_NotFound(t *testing.T) {
	_, err := ReadNoteFile("/nonexistent/path/note.md")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

// Phase 6: Todo Management Tests

func TestAppendTodoToProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	err := AppendTodoToProject(projectPath, "New task from API")
	if err != nil {
		t.Fatalf("AppendTodoToProject failed: %v", err)
	}

	// Verify task was added
	todoFile := filepath.Join(projectPath, "todo.md")
	content, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read todo.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "New task from API") {
		t.Errorf("Task not found in todo.md:\n%s", contentStr)
	}

	// Verify ID was added (stored in HTML comment: <!-- id:... -->)
	if !strings.Contains(contentStr, "<!-- id:") {
		t.Error("ID not added to task")
	}

	// Verify under Active section
	lines := strings.Split(contentStr, "\n")
	activeIdx := -1
	taskIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "## Active" {
			activeIdx = i
		}
		if strings.Contains(line, "New task from API") {
			taskIdx = i
		}
	}

	if taskIdx <= activeIdx {
		t.Error("Task not under Active section")
	}
}

func TestAppendTodoToProject_Multiple(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Add multiple tasks
	tasks := []string{"Task 1", "Task 2", "Task 3"}
	for _, task := range tasks {
		if err := AppendTodoToProject(projectPath, task); err != nil {
			t.Fatalf("AppendTodoToProject failed: %v", err)
		}
	}

	// Verify all tasks exist
	todoFile := filepath.Join(projectPath, "todo.md")
	content, _ := os.ReadFile(todoFile)
	contentStr := string(content)

	for _, task := range tasks {
		if !strings.Contains(contentStr, task) {
			t.Errorf("Task '%s' not found", task)
		}
	}
}

func TestAppendTodoToProject_MissingFile(t *testing.T) {
	// Create project directory without todo.md
	projectPath := t.TempDir()

	err := AppendTodoToProject(projectPath, "Test task")
	if err != nil {
		t.Fatalf("AppendTodoToProject failed: %v", err)
	}

	// Should create todo.md automatically
	todoFile := filepath.Join(projectPath, "todo.md")
	if _, err := os.Stat(todoFile); err != nil {
		t.Error("todo.md was not created")
	}

	// Verify task was added
	content, _ := os.ReadFile(todoFile)
	if !strings.Contains(string(content), "Test task") {
		t.Error("Task not found in newly created todo.md")
	}
}
