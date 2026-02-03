package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/markdown"
	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestRefileTaskToProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create project
	projectPath := tb.AddProject("test-project")

	// Create a dump item (task)
	item := &markdown.DumpItem{
		Type:      markdown.ItemTypeTodo,
		Content:   "Test task #captured:2024-01-15",
		StartLine: 3,
		EndLine:   3,
		RawLine:   "- [ ] Test task #captured:2024-01-15",
	}

	// Refile task
	err := RefileTaskToProject(projectPath, item)
	if err != nil {
		t.Fatalf("RefileTaskToProject failed: %v", err)
	}

	// Verify task was added to todo.md
	todoFile := filepath.Join(projectPath, "todo.md")
	content, err := os.ReadFile(todoFile)
	if err != nil {
		t.Fatalf("Failed to read todo.md: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Test task") {
		t.Errorf("Task not found in todo.md:\n%s", contentStr)
	}

	// Verify task is under Active section
	lines := strings.Split(contentStr, "\n")
	activeIdx := -1
	taskIdx := -1

	for i, line := range lines {
		if strings.TrimSpace(line) == "## Active" {
			activeIdx = i
		}
		if strings.Contains(line, "Test task") {
			taskIdx = i
		}
	}

	if activeIdx == -1 {
		t.Error("Active section not found")
	}
	if taskIdx == -1 {
		t.Error("Task not found in file")
	}
	if taskIdx <= activeIdx {
		t.Error("Task not under Active section")
	}
}

func TestRefileTaskToProject_MissingTodoFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create project without todo.md
	projectPath := filepath.Join(tb.ActiveDirPath, "test-project")
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project path: %v", err)
	}

	item := &markdown.DumpItem{
		Type:      markdown.ItemTypeTodo,
		Content:   "Test task",
		StartLine: 3,
		EndLine:   3,
		RawLine:   "- [ ] Test task",
	}

	// Should create todo.md automatically
	err := RefileTaskToProject(projectPath, item)
	if err != nil {
		t.Fatalf("RefileTaskToProject failed: %v", err)
	}

	todoFile := filepath.Join(projectPath, "todo.md")
	if _, err := os.Stat(todoFile); err != nil {
		t.Error("todo.md was not created")
	}
}

func TestRefileTaskToProject_WrongType(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Try to refile a note as a task
	item := &markdown.DumpItem{
		Type:    markdown.ItemTypeNote,
		Content: "Note title",
	}

	err := RefileTaskToProject(projectPath, item)
	if err == nil {
		t.Error("Expected error for wrong item type")
	}
}

func TestRefileNoteToProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create project
	projectPath := tb.AddProject("test-project")

	// Add a note to dump
	if err := AddNoteToDump(tb.DumpPath, "Meeting notes", []string{
		"First point",
		"Second point",
	}, "2024-01-15"); err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	// Parse dump to get the note item
	items, err := markdown.ParseDumpFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to parse dump: %v", err)
	}

	var noteItem *markdown.DumpItem
	for i := range items {
		if items[i].Type == markdown.ItemTypeNote {
			noteItem = &items[i]
			break
		}
	}

	if noteItem == nil {
		t.Fatal("Note not found in dump")
	}

	// Refile note
	filePath, err := RefileNoteToProject(tb.DumpPath, projectPath, noteItem)
	if err != nil {
		t.Fatalf("RefileNoteToProject failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("Note file was not created: %v", err)
	}

	// Verify file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read note file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Meeting notes") {
		t.Error("Title not found in note file")
	}
	if !strings.Contains(contentStr, "First point") {
		t.Error("First point not found in note file")
	}
	if !strings.Contains(contentStr, "Second point") {
		t.Error("Second point not found in note file")
	}
}

func TestRefileNoteToProject_DuplicateFilename(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Add two notes with same title
	if err := AddNoteToDump(tb.DumpPath, "Same title", []string{"Content 1"}, "2024-01-15"); err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}
	if err := AddNoteToDump(tb.DumpPath, "Same title", []string{"Content 2"}, "2024-01-15"); err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	items, _ := markdown.ParseDumpFile(tb.DumpPath)

	// Refile both notes
	var filePaths []string
	for i := range items {
		if items[i].Type == markdown.ItemTypeNote {
			filePath, err := RefileNoteToProject(tb.DumpPath, projectPath, &items[i])
			if err != nil {
				t.Fatalf("RefileNoteToProject failed: %v", err)
			}
			filePaths = append(filePaths, filePath)
		}
	}

	// Should have created two different files
	if len(filePaths) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(filePaths))
	}

	if filePaths[0] == filePaths[1] {
		t.Error("Duplicate filenames created")
	}

	// Both files should exist
	for _, path := range filePaths {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("File not created: %s", path)
		}
	}
}

func TestRefileNoteToProject_WrongType(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	projectPath := tb.AddProject("test-project")

	// Try to refile a task as a note
	item := &markdown.DumpItem{
		Type:    markdown.ItemTypeTodo,
		Content: "Task content",
	}

	_, err := RefileNoteToProject(tb.DumpPath, projectPath, item)
	if err == nil {
		t.Error("Expected error for wrong item type")
	}
}

func Test_readNoteContentFromDump(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Add a multi-line note
	if err := AddNoteToDump(tb.DumpPath, "Test note", []string{
		"Line 1",
		"Line 2",
		"Line 3",
	}, "2024-01-15"); err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	// Parse to get line numbers
	items, err := markdown.ParseDumpFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to parse dump: %v", err)
	}

	var noteItem *markdown.DumpItem
	for i := range items {
		if items[i].Type == markdown.ItemTypeNote {
			noteItem = &items[i]
			break
		}
	}

	if noteItem == nil {
		t.Fatal("Note not found")
	}

	// Read content
	content, err := readNoteContentFromDump(tb.DumpPath, noteItem.StartLine, noteItem.EndLine)
	if err != nil {
		t.Fatalf("readNoteContentFromDump failed: %v", err)
	}

	// Verify all lines present (without indentation)
	if !strings.Contains(content, "Line 1") {
		t.Error("Line 1 not found")
	}
	if !strings.Contains(content, "Line 2") {
		t.Error("Line 2 not found")
	}
	if !strings.Contains(content, "Line 3") {
		t.Error("Line 3 not found")
	}

	// Verify no extra indentation
	if strings.Contains(content, "    Line 1") {
		t.Error("Content should not have indentation")
	}
}
