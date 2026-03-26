package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"valid simple", "project", false},
		{"valid with hyphen", "project-name", false},
		{"valid with underscore", "project_name", false},
		{"valid with number", "project123", false},
		{"valid mixed", "my-project_v2", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 65), true},
		{"starts with hyphen", "-project", true},
		{"starts with underscore", "_project", true},
		{"starts with dot", ".project", true},
		{"contains space", "my project", true},
		{"contains special char", "project!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateProjectName(tt.input)
			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for '%s'", tt.input)
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for '%s': %v", tt.input, err)
			}
		})
	}
}

func TestCreateProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	projectPath, err := CreateProject(tb.ActiveDirPath, "test-project")
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	expectedPath := filepath.Join(tb.ActiveDirPath, "test-project")
	if projectPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, projectPath)
	}

	// Verify directory structure
	if !tb.DirExists(projectPath) {
		t.Error("Project directory was not created")
	}

	// Verify todo.md
	todoPath := filepath.Join(projectPath, "todo.md")
	if !tb.FileExists(todoPath) {
		t.Error("todo.md was not created")
	}
	todoContent := tb.ReadFile(todoPath)
	if !strings.Contains(todoContent, "## Active") {
		t.Error("todo.md missing Active section")
	}

	// Verify notes.md
	notesPath := filepath.Join(projectPath, "notes.md")
	if !tb.FileExists(notesPath) {
		t.Error("notes.md was not created")
	}

	// Verify notes/ directory
	notesDir := filepath.Join(projectPath, "notes")
	if !tb.DirExists(notesDir) {
		t.Error("notes/ directory was not created")
	}

	// Verify .repos file
	reposPath := filepath.Join(projectPath, ".repos")
	if !tb.FileExists(reposPath) {
		t.Error(".repos file was not created")
	}
}

func TestCreateProject_Duplicate(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create first project
	_, err := CreateProject(tb.ActiveDirPath, "test-project")
	if err != nil {
		t.Fatalf("First CreateProject failed: %v", err)
	}

	// Try to create duplicate
	_, err = CreateProject(tb.ActiveDirPath, "test-project")
	if err == nil {
		t.Error("Expected error for duplicate project")
	}
}

func TestCreateProject_InvalidName(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	_, err := CreateProject(tb.ActiveDirPath, "invalid project!")
	if err == nil {
		t.Error("Expected error for invalid project name")
	}
}

func TestArchiveProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create a project
	projectName := "test-project"
	if _, err := CreateProject(tb.ActiveDirPath, projectName); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Archive it
	err := ArchiveProject(tb.BrainPath, projectName)
	if err != nil {
		t.Fatalf("ArchiveProject failed: %v", err)
	}

	// Verify project no longer in active
	activePath := filepath.Join(tb.ActiveDirPath, projectName)
	if tb.FileExists(activePath) {
		t.Error("Project still exists in active directory")
	}

	// Verify project in archive with timestamp
	archiveDir := filepath.Join(tb.BrainPath, "99_archive")
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("Failed to read archive directory: %v", err)
	}

	found := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), projectName+"-") {
			found = true
			break
		}
	}

	if !found {
		t.Error("Project not found in archive directory")
	}
}

func TestArchiveProject_NotFound(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	err := ArchiveProject(tb.BrainPath, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent project")
	}
}

func TestDeleteProject(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	// Create a project
	projectName := "test-project"
	if _, err := CreateProject(tb.ActiveDirPath, projectName); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	projectPath := filepath.Join(tb.ActiveDirPath, projectName)
	if !tb.DirExists(projectPath) {
		t.Fatal("Project was not created")
	}

	// Delete it
	err := DeleteProject(tb.ActiveDirPath, projectName)
	if err != nil {
		t.Fatalf("DeleteProject failed: %v", err)
	}

	// Verify it's gone
	if tb.FileExists(projectPath) {
		t.Error("Project still exists after deletion")
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	err := DeleteProject(tb.ActiveDirPath, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent project")
	}
}

func TestMoveProjectToBrain(t *testing.T) {
	// Create two test brains
	srcBrain := t.TempDir()
	dstBrain := t.TempDir()

	srcActiveDir := filepath.Join(srcBrain, "01_active")
	dstActiveDir := filepath.Join(dstBrain, "01_active")

	if err := os.MkdirAll(srcActiveDir, 0755); err != nil {
		t.Fatalf("Failed to create srcActiveDir: %v", err)
	}
	if err := os.MkdirAll(dstActiveDir, 0755); err != nil {
		t.Fatalf("Failed to create dstActiveDir: %v", err)
	}

	// Create a project in source brain
	projectName := "test-project"
	if _, err := CreateProject(srcActiveDir, projectName); err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	// Add some content to verify it's copied
	projectPath := filepath.Join(srcActiveDir, projectName)
	testFile := filepath.Join(projectPath, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Move project
	err := MoveProjectToBrain(srcBrain, dstBrain, projectName)
	if err != nil {
		t.Fatalf("MoveProjectToBrain failed: %v", err)
	}

	// Verify project removed from source
	srcPath := filepath.Join(srcActiveDir, projectName)
	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("Project still exists in source brain")
	}

	// Verify project exists in destination
	dstPath := filepath.Join(dstActiveDir, projectName)
	if _, err := os.Stat(dstPath); err != nil {
		t.Error("Project not found in destination brain")
	}

	// Verify content was copied
	dstTestFile := filepath.Join(dstPath, "test.txt")
	content, err := os.ReadFile(dstTestFile)
	if err != nil {
		t.Errorf("Test file not copied: %v", err)
	}
	if string(content) != "test content" {
		t.Error("File content not preserved")
	}
}

func TestMoveProjectToBrain_SourceNotFound(t *testing.T) {
	srcBrain := t.TempDir()
	dstBrain := t.TempDir()

	if err := os.MkdirAll(filepath.Join(srcBrain, "01_active"), 0755); err != nil {
		t.Fatalf("Failed to create source active dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dstBrain, "01_active"), 0755); err != nil {
		t.Fatalf("Failed to create destination active dir: %v", err)
	}

	err := MoveProjectToBrain(srcBrain, dstBrain, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent source project")
	}
}

func TestMoveProjectToBrain_DestinationExists(t *testing.T) {
	srcBrain := t.TempDir()
	dstBrain := t.TempDir()

	srcActiveDir := filepath.Join(srcBrain, "01_active")
	dstActiveDir := filepath.Join(dstBrain, "01_active")

	if err := os.MkdirAll(srcActiveDir, 0755); err != nil {
		t.Fatalf("Failed to create srcActiveDir: %v", err)
	}
	if err := os.MkdirAll(dstActiveDir, 0755); err != nil {
		t.Fatalf("Failed to create dstActiveDir: %v", err)
	}

	// Create project in both brains
	projectName := "test-project"
	if _, err := CreateProject(srcActiveDir, projectName); err != nil {
		t.Fatalf("CreateProject failed in src: %v", err)
	}
	if _, err := CreateProject(dstActiveDir, projectName); err != nil {
		t.Fatalf("CreateProject failed in dst: %v", err)
	}

	err := MoveProjectToBrain(srcBrain, dstBrain, projectName)
	if err == nil {
		t.Error("Expected error for duplicate project in destination")
	}
}
