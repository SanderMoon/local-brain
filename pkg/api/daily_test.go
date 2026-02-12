package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateOrOpenDailyNote_NewNote(t *testing.T) {
	tmpDir := t.TempDir()
	brainPath := filepath.Join(tmpDir, "test-brain")
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain directory: %v", err)
	}

	date := "2026-02-12"
	result, err := CreateOrOpenDailyNote(brainPath, date, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.IsNew {
		t.Error("Expected IsNew=true for a new note")
	}

	if result.Date != date {
		t.Errorf("Expected Date=%s, got %s", date, result.Date)
	}

	expectedPath := filepath.Join(brainPath, "00_daily", date+".md")
	if result.Path != expectedPath {
		t.Errorf("Expected Path=%s, got %s", expectedPath, result.Path)
	}

	// Verify file exists and contains YAML frontmatter
	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "---\n") {
		t.Error("Expected file to start with YAML frontmatter '---'")
	}

	if !strings.Contains(contentStr, "title: "+date) {
		t.Errorf("Expected file to contain 'title: %s'", date)
	}

	if !strings.Contains(contentStr, "date: "+date) {
		t.Errorf("Expected file to contain 'date: %s'", date)
	}

	if !strings.Contains(contentStr, "# "+date) {
		t.Errorf("Expected file to contain heading '# %s'", date)
	}

	if !strings.Contains(contentStr, "## Daily Briefing") {
		t.Error("Expected file to contain '## Daily Briefing' section")
	}

	if !strings.Contains(contentStr, "## Today's Focus") {
		t.Error("Expected file to contain \"## Today's Focus\" section")
	}

	if !strings.Contains(contentStr, "## Notes") {
		t.Error("Expected file to contain '## Notes' section")
	}
}

func TestCreateOrOpenDailyNote_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	brainPath := filepath.Join(tmpDir, "test-brain")
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain directory: %v", err)
	}

	date := "2026-02-12"

	// First call - creates the note
	first, err := CreateOrOpenDailyNote(brainPath, date, nil)
	if err != nil {
		t.Fatalf("First call failed: %v", err)
	}
	if !first.IsNew {
		t.Error("Expected first call to return IsNew=true")
	}

	// Read original content
	originalContent, err := os.ReadFile(first.Path)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	// Second call - should return existing note without modification
	second, err := CreateOrOpenDailyNote(brainPath, date, nil)
	if err != nil {
		t.Fatalf("Second call failed: %v", err)
	}
	if second.IsNew {
		t.Error("Expected second call to return IsNew=false")
	}

	if second.Path != first.Path {
		t.Errorf("Expected same path on second call: %s vs %s", second.Path, first.Path)
	}

	// Verify file content unchanged
	currentContent, err := os.ReadFile(second.Path)
	if err != nil {
		t.Fatalf("Failed to read file after second call: %v", err)
	}

	if string(currentContent) != string(originalContent) {
		t.Error("Expected file content to be unchanged after second call")
	}
}

func TestCreateOrOpenDailyNote_CreatesDailyDir(t *testing.T) {
	tmpDir := t.TempDir()
	brainPath := filepath.Join(tmpDir, "test-brain")
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain directory: %v", err)
	}

	date := "2026-02-12"
	dailyDir := filepath.Join(brainPath, "00_daily")

	// Ensure directory does not exist before
	if _, err := os.Stat(dailyDir); !os.IsNotExist(err) {
		t.Fatal("Expected 00_daily directory to not exist before call")
	}

	_, err := CreateOrOpenDailyNote(brainPath, date, nil)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(dailyDir)
	if err != nil {
		t.Fatalf("Expected 00_daily directory to exist after call: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected 00_daily to be a directory")
	}
}

func TestCreateOrOpenDailyNote_IncludesOverdueTodos(t *testing.T) {
	tmpDir := t.TempDir()
	brainPath := filepath.Join(tmpDir, "test-brain")
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain directory: %v", err)
	}

	date := "2026-02-12"
	overdueTodos := []TodoItem{
		{
			Content: "Fix critical bug",
			Project: "my-project",
			DueDate: "2026-02-10",
			Status:  "open",
		},
	}

	result, err := CreateOrOpenDailyNote(brainPath, date, overdueTodos)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "Fix critical bug") {
		t.Error("Expected overdue todo content to appear in file")
	}
	if !strings.Contains(contentStr, "my-project") {
		t.Error("Expected overdue todo project to appear in file")
	}
	if !strings.Contains(contentStr, "2026-02-10") {
		t.Error("Expected overdue todo due date to appear in file")
	}
	if !strings.Contains(contentStr, "- [ ] [my-project] Fix critical bug (due: 2026-02-10)") {
		t.Error("Expected properly formatted overdue todo line in file")
	}
}

func TestCreateOrOpenDailyNote_EmptyOverdue(t *testing.T) {
	tmpDir := t.TempDir()
	brainPath := filepath.Join(tmpDir, "test-brain")
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain directory: %v", err)
	}

	date := "2026-02-12"

	result, err := CreateOrOpenDailyNote(brainPath, date, []TodoItem{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	content, err := os.ReadFile(result.Path)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if !strings.Contains(string(content), "(no overdue items)") {
		t.Error("Expected '(no overdue items)' when no overdue todos provided")
	}
}
