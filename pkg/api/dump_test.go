package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestParseDumpToJSON(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Add some content to dump
	tb.AddTaskToDump("Task 1", "2024-01-01")
	tb.AddTaskToDump("Task 2", "2024-01-02")
	tb.AddNoteToDump("Note 1", []string{"Line 1", "Line 2"}, "2024-01-03")

	items, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Verify tasks
	if items[0].Type != "todo" {
		t.Errorf("Item 0: expected type 'todo', got '%s'", items[0].Type)
	}
	if !strings.Contains(items[0].Content, "Task 1") {
		t.Errorf("Item 0: unexpected content: %s", items[0].Content)
	}
	if items[0].Timestamp != "2024-01-01" {
		t.Errorf("Item 0: expected timestamp '2024-01-01', got '%s'", items[0].Timestamp)
	}

	// Verify note
	if items[2].Type != "note" {
		t.Errorf("Item 2: expected type 'note', got '%s'", items[2].Type)
	}
	if items[2].Content != "Note 1" {
		t.Errorf("Item 2: unexpected content: %s", items[2].Content)
	}
}

func TestParseDumpToJSONBytes(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Test task", "2024-01-01")

	jsonBytes, err := ParseDumpToJSONBytes(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSONBytes failed: %v", err)
	}

	// Should be valid JSON
	var items []DumpItemJSON
	if err := json.Unmarshal(jsonBytes, &items); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}

	if !strings.Contains(items[0].Content, "Test task") {
		t.Errorf("Unexpected content: %s", items[0].Content)
	}
}

func TestParseDumpToJSONString(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Test task", "2024-01-01")
	tb.AddNoteToDump("Test note", []string{"Content"}, "2024-01-02")

	jsonStr, err := ParseDumpToJSONString(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSONString failed: %v", err)
	}

	// Should be valid JSON string
	var items []DumpItemJSON
	if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
		t.Fatalf("Invalid JSON string: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(items))
	}
}

func TestParseDumpToJSON_EmptyFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Dump file exists but only has header
	items, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed on empty dump: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items from empty dump, got %d", len(items))
	}
}

func TestParseDumpToJSON_NonExistent(t *testing.T) {
	tmpDir := filepath.Join("/tmp", "nonexistent-12345")

	_, err := ParseDumpToJSON(tmpDir)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseDumpToJSONBytes_EmptyFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	jsonBytes, err := ParseDumpToJSONBytes(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSONBytes failed: %v", err)
	}

	// Should be empty JSON array
	var items []DumpItemJSON
	if err := json.Unmarshal(jsonBytes, &items); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected empty array, got %d items", len(items))
	}
}

func TestParseDumpToJSONString_EmptyFile(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	jsonStr, err := ParseDumpToJSONString(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSONString failed: %v", err)
	}

	// Should be empty JSON array as string
	if jsonStr != "[]" && jsonStr != "[\n]" {
		// Allow for formatting variations
		var items []DumpItemJSON
		if err := json.Unmarshal([]byte(jsonStr), &items); err != nil {
			t.Fatalf("Invalid JSON string: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected empty array, got %d items", len(items))
		}
	}
}

func TestAddTaskToDump(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	err := AddTaskToDump(tb.DumpPath, "Test task", "2024-01-15")
	if err != nil {
		t.Fatalf("AddTaskToDump failed: %v", err)
	}

	// Verify task was added
	content, err := os.ReadFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to read dump: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "- [ ] Test task") {
		t.Errorf("Expected task in dump, got:\n%s", contentStr)
	}

	if !strings.Contains(contentStr, "#captured:2024-01-15") {
		t.Errorf("Expected timestamp in dump, got:\n%s", contentStr)
	}

	// Verify ID was added
	if !strings.Contains(contentStr, "#id:") {
		t.Error("Expected #id: tag in task")
	}
}

func TestAddTaskToDump_Multiple(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Add multiple tasks
	tasks := []struct {
		content   string
		timestamp string
	}{
		{"Task 1", "2024-01-01"},
		{"Task 2", "2024-01-02"},
		{"Task 3", "2024-01-03"},
	}

	for _, task := range tasks {
		if err := AddTaskToDump(tb.DumpPath, task.content, task.timestamp); err != nil {
			t.Fatalf("AddTaskToDump failed: %v", err)
		}
	}

	// Parse and verify all tasks
	items, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(items))
	}

	for i, task := range tasks {
		if !strings.Contains(items[i].Content, task.content) {
			t.Errorf("Task %d: expected content '%s', got '%s'", i, task.content, items[i].Content)
		}
		if items[i].Timestamp != task.timestamp {
			t.Errorf("Task %d: expected timestamp '%s', got '%s'", i, task.timestamp, items[i].Timestamp)
		}
	}
}

func TestAddNoteToDump(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	contentLines := []string{
		"First line of note",
		"Second line of note",
		"Third line of note",
	}

	err := AddNoteToDump(tb.DumpPath, "Meeting notes", contentLines, "2024-01-15")
	if err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	// Verify note was added
	content, err := os.ReadFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to read dump: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[Note] Meeting notes") {
		t.Errorf("Expected note header in dump, got:\n%s", contentStr)
	}

	if !strings.Contains(contentStr, "#captured:2024-01-15") {
		t.Errorf("Expected timestamp in dump, got:\n%s", contentStr)
	}

	// Verify content lines are indented
	for _, line := range contentLines {
		expected := "    " + line
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Expected indented line '%s' in dump", expected)
		}
	}

	// Verify ID was added
	if !strings.Contains(contentStr, "#id:") {
		t.Error("Expected #id: tag in note header")
	}
}

func TestAddNoteToDump_EmptyContent(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	err := AddNoteToDump(tb.DumpPath, "Empty note", []string{}, "2024-01-15")
	if err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	// Verify note header was added
	content, err := os.ReadFile(tb.DumpPath)
	if err != nil {
		t.Fatalf("Failed to read dump: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "[Note] Empty note") {
		t.Errorf("Expected note header in dump, got:\n%s", contentStr)
	}
}

func TestRemoveItemFromDump(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Add some items
	tb.AddTaskToDump("Task 1", "2024-01-01")
	tb.AddTaskToDump("Task 2", "2024-01-02")
	tb.AddTaskToDump("Task 3", "2024-01-03")

	// Parse to get line numbers
	items, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items before removal, got %d", len(items))
	}

	// Remove the second task (item[1])
	err = RemoveItemFromDump(tb.DumpPath, items[1].StartLine, items[1].EndLine)
	if err != nil {
		t.Fatalf("RemoveItemFromDump failed: %v", err)
	}

	// Verify only 2 items remain
	itemsAfter, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed after removal: %v", err)
	}

	if len(itemsAfter) != 2 {
		t.Fatalf("Expected 2 items after removal, got %d", len(itemsAfter))
	}

	// Verify the correct item was removed
	content, _ := os.ReadFile(tb.DumpPath)
	contentStr := string(content)

	if !strings.Contains(contentStr, "Task 1") {
		t.Error("Task 1 should still exist")
	}
	if strings.Contains(contentStr, "Task 2") {
		t.Error("Task 2 should be removed")
	}
	if !strings.Contains(contentStr, "Task 3") {
		t.Error("Task 3 should still exist")
	}
}

func TestRemoveItemFromDump_FirstItem(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Task 1", "2024-01-01")
	tb.AddTaskToDump("Task 2", "2024-01-02")

	items, _ := ParseDumpToJSON(tb.DumpPath)

	// Remove first item
	err := RemoveItemFromDump(tb.DumpPath, items[0].StartLine, items[0].EndLine)
	if err != nil {
		t.Fatalf("RemoveItemFromDump failed: %v", err)
	}

	itemsAfter, _ := ParseDumpToJSON(tb.DumpPath)
	if len(itemsAfter) != 1 {
		t.Fatalf("Expected 1 item after removal, got %d", len(itemsAfter))
	}

	if !strings.Contains(itemsAfter[0].Content, "Task 2") {
		t.Error("Wrong item was removed")
	}
}

func TestRemoveItemFromDump_LastItem(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Task 1", "2024-01-01")
	tb.AddTaskToDump("Task 2", "2024-01-02")

	items, _ := ParseDumpToJSON(tb.DumpPath)

	// Remove last item
	err := RemoveItemFromDump(tb.DumpPath, items[1].StartLine, items[1].EndLine)
	if err != nil {
		t.Fatalf("RemoveItemFromDump failed: %v", err)
	}

	itemsAfter, _ := ParseDumpToJSON(tb.DumpPath)
	if len(itemsAfter) != 1 {
		t.Fatalf("Expected 1 item after removal, got %d", len(itemsAfter))
	}

	if !strings.Contains(itemsAfter[0].Content, "Task 1") {
		t.Error("Wrong item was removed")
	}
}

func TestRemoveItemFromDump_MultiLineNote(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Task 1", "2024-01-01")
	tb.AddNoteToDump("Note 1", []string{"Line 1", "Line 2", "Line 3"}, "2024-01-02")
	tb.AddTaskToDump("Task 2", "2024-01-03")

	items, _ := ParseDumpToJSON(tb.DumpPath)

	// Remove the multi-line note (item[1])
	err := RemoveItemFromDump(tb.DumpPath, items[1].StartLine, items[1].EndLine)
	if err != nil {
		t.Fatalf("RemoveItemFromDump failed: %v", err)
	}

	// Verify only tasks remain
	itemsAfter, _ := ParseDumpToJSON(tb.DumpPath)
	if len(itemsAfter) != 2 {
		t.Fatalf("Expected 2 items after removal, got %d", len(itemsAfter))
	}

	content, _ := os.ReadFile(tb.DumpPath)
	contentStr := string(content)

	// Note should be completely gone (including indented content)
	if strings.Contains(contentStr, "Note 1") {
		t.Error("Note should be removed")
	}
	if strings.Contains(contentStr, "Line 1") {
		t.Error("Note content should be removed")
	}
}

func TestRemoveItemFromDump_InvalidRange(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Task 1", "2024-01-01")

	// Test invalid line ranges
	tests := []struct {
		name      string
		startLine int
		endLine   int
	}{
		{"negative start", -1, 5},
		{"zero start", 0, 5},
		{"end before start", 5, 3},
		{"start beyond file", 1000, 1001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RemoveItemFromDump(tb.DumpPath, tt.startLine, tt.endLine)
			if err == nil {
				t.Error("Expected error for invalid range")
			}
		})
	}
}

func TestFindDumpItemByID(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Add items using API functions (which add IDs)
	if err := AddTaskToDump(tb.DumpPath, "Task 1", "2024-01-01"); err != nil {
		t.Fatalf("AddTaskToDump failed: %v", err)
	}
	if err := AddTaskToDump(tb.DumpPath, "Task 2", "2024-01-02"); err != nil {
		t.Fatalf("AddTaskToDump failed: %v", err)
	}
	if err := AddNoteToDump(tb.DumpPath, "Note 1", []string{"Content"}, "2024-01-03"); err != nil {
		t.Fatalf("AddNoteToDump failed: %v", err)
	}

	// Get all items to find their IDs
	items, err := ParseDumpToJSON(tb.DumpPath)
	if err != nil {
		t.Fatalf("ParseDumpToJSON failed: %v", err)
	}

	// Find second task by ID
	targetID := items[1].ID
	foundItem, err := FindDumpItemByID(tb.DumpPath, targetID)
	if err != nil {
		t.Fatalf("FindDumpItemByID failed: %v", err)
	}

	if foundItem == nil {
		t.Fatal("Expected to find item, got nil")
	}

	if !strings.Contains(foundItem.Content, "Task 2") {
		t.Errorf("Found wrong item: %s", foundItem.Content)
	}
}

func TestFindDumpItemByID_NotFound(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	tb.AddTaskToDump("Task 1", "2024-01-01")

	// Search for non-existent ID
	foundItem, err := FindDumpItemByID(tb.DumpPath, "nonexistent123")
	if err != nil {
		t.Fatalf("FindDumpItemByID failed: %v", err)
	}

	if foundItem != nil {
		t.Error("Expected nil for non-existent ID, got item")
	}
}

func TestFindDumpItemByID_EmptyDump(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Search in empty dump
	foundItem, err := FindDumpItemByID(tb.DumpPath, "some-id")
	if err != nil {
		t.Fatalf("FindDumpItemByID failed: %v", err)
	}

	if foundItem != nil {
		t.Error("Expected nil for empty dump, got item")
	}
}
