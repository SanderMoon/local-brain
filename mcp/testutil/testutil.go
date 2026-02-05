package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandermoonemans/local-brain/mcp/session"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// SetupTestBrain creates a test brain structure in a temporary directory
func SetupTestBrain(t *testing.T) (string, *config.Config) {
	t.Helper()

	tmpDir := t.TempDir()
	brainRoot := filepath.Join(tmpDir, "brains")
	brainPath := filepath.Join(brainRoot, "test-brain")
	configDir := filepath.Join(tmpDir, "config")

	// Set environment variables for test isolation
	t.Setenv("BRAIN_ROOT", brainRoot)
	t.Setenv("BRAIN_CONFIG_DIR", configDir)
	t.Setenv("BRAIN_SYMLINK", filepath.Join(tmpDir, "brain"))

	// Create brain directory structure
	if err := os.MkdirAll(brainPath, 0755); err != nil {
		t.Fatalf("Failed to create brain path: %v", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")
	if err := os.MkdirAll(activeDir, 0755); err != nil {
		t.Fatalf("Failed to create active dir: %v", err)
	}

	// Create dump file
	dumpPath := filepath.Join(brainPath, "00_dump.md")
	dumpContent := `# Inbox

## Tasks
- [ ] Test task #captured:2026-02-01 #id:abc123

## Notes
`
	if err := os.WriteFile(dumpPath, []byte(dumpContent), 0644); err != nil {
		t.Fatalf("Failed to create dump file: %v", err)
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create config manually
	cfg := &config.Config{
		Current: "test-brain",
		Brains: map[string]*config.BrainInfo{
			"test-brain": {
				Path:    brainPath,
				Created: "2026-02-01",
				Focus:   "",
			},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Reload config to get proper initialization
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}

	return brainPath, cfg
}

// CreateTestProject creates a test project with todos and notes
func CreateTestProject(t *testing.T, brainPath, projectName string) string {
	t.Helper()

	projectPath := filepath.Join(brainPath, "01_active", projectName)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		t.Fatalf("Failed to create project dir: %v", err)
	}

	// Create todo.md
	todoPath := filepath.Join(projectPath, "todo.md")
	todoContent := `# Tasks

- [ ] Todo 1 #id:def456
- [x] Todo 2 #id:ghi789 #done:2026-02-01
- [ ] Todo 3 #priority:1 #due:2026-02-15 #id:jkl012
`
	if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
		t.Fatalf("Failed to create todo file: %v", err)
	}

	// Create notes.md
	notesPath := filepath.Join(projectPath, "notes.md")
	notesContent := `# Notes

## Test Note #captured:2026-02-01
Some content here
`
	if err := os.WriteFile(notesPath, []byte(notesContent), 0644); err != nil {
		t.Fatalf("Failed to create notes file: %v", err)
	}

	// Create description
	descPath := filepath.Join(projectPath, "description.txt")
	if err := os.WriteFile(descPath, []byte("Test project description"), 0644); err != nil {
		t.Fatalf("Failed to create description file: %v", err)
	}

	return projectPath
}

// NewTestSession creates a test session with the given config
func NewTestSession(cfg *config.Config) *session.Session {
	return session.NewSession(cfg)
}
