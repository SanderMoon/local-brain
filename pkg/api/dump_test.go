package api

import (
	"encoding/json"
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
