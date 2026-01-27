package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBrain represents an isolated test brain environment
type TestBrain struct {
	T             *testing.T
	TmpDir        string
	ConfigPath    string
	ConfigDir     string
	BrainPath     string
	SymlinkPath   string
	DumpPath      string
	ActiveDirPath string
}

// SetupTestBrain creates an isolated test environment for brain operations
// All paths are in a temporary directory that is automatically cleaned up after the test
func SetupTestBrain(t *testing.T) *TestBrain {
	t.Helper()

	tmpDir := t.TempDir() // Auto-cleanup on test completion

	// Create isolated paths
	configDir := filepath.Join(tmpDir, "config")
	configPath := filepath.Join(configDir, "config.json")
	brainPath := filepath.Join(tmpDir, "test-brain")
	symlinkPath := filepath.Join(tmpDir, "brain-link")

	// Set environment variables for this test
	t.Setenv("BRAIN_CONFIG_DIR", configDir)
	t.Setenv("BRAIN_CONFIG_PATH", configPath)
	t.Setenv("BRAIN_SYMLINK", symlinkPath)

	// Create brain directory structure
	activeDirPath := filepath.Join(brainPath, "01_active")
	if err := os.MkdirAll(activeDirPath, 0755); err != nil {
		t.Fatalf("Failed to create active directory: %v", err)
	}

	// Create dump file
	dumpPath := filepath.Join(brainPath, "00_dump.md")
	dumpContent := `# Dump

`
	if err := os.WriteFile(dumpPath, []byte(dumpContent), 0644); err != nil {
		t.Fatalf("Failed to create dump file: %v", err)
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Create minimal config as raw JSON (avoid circular dependency with config package)
	configData := map[string]interface{}{
		"current": "test",
		"brains": map[string]interface{}{
			"test": map[string]interface{}{
				"path":    brainPath,
				"created": "2024-01-01",
			},
		},
	}

	// Write config file
	data, err := json.MarshalIndent(configData, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	return &TestBrain{
		T:             t,
		TmpDir:        tmpDir,
		ConfigPath:    configPath,
		ConfigDir:     configDir,
		BrainPath:     brainPath,
		SymlinkPath:   symlinkPath,
		DumpPath:      dumpPath,
		ActiveDirPath: activeDirPath,
	}
}

// AddProject creates a new project directory in the test brain
func (tb *TestBrain) AddProject(name string) string {
	tb.T.Helper()

	projectPath := filepath.Join(tb.ActiveDirPath, name)
	if err := os.MkdirAll(projectPath, 0755); err != nil {
		tb.T.Fatalf("Failed to create project directory: %v", err)
	}

	// Create project todo.md
	todoPath := filepath.Join(projectPath, "todo.md")
	todoContent := `# ` + name + `

## Active

## Completed

`
	if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
		tb.T.Fatalf("Failed to create todo file: %v", err)
	}

	// Create project notes.md
	notesPath := filepath.Join(projectPath, "notes.md")
	notesContent := `# ` + name + ` Notes

`
	if err := os.WriteFile(notesPath, []byte(notesContent), 0644); err != nil {
		tb.T.Fatalf("Failed to create notes file: %v", err)
	}

	return projectPath
}

// AddToDump adds a task or note to the dump file
func (tb *TestBrain) AddToDump(content string) {
	tb.T.Helper()

	f, err := os.OpenFile(tb.DumpPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		tb.T.Fatalf("Failed to open dump file: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content + "\n"); err != nil {
		tb.T.Fatalf("Failed to write to dump file: %v", err)
	}
}

// AddTaskToDump adds a task to the dump file
func (tb *TestBrain) AddTaskToDump(task string, timestamp string) {
	tb.T.Helper()

	taskLine := "- [ ] " + task
	if timestamp != "" {
		taskLine += " #captured:" + timestamp
	}
	tb.AddToDump(taskLine)
}

// AddNoteToDump adds a note to the dump file
func (tb *TestBrain) AddNoteToDump(title string, lines []string, timestamp string) {
	tb.T.Helper()

	noteHeader := "[Note] " + title
	if timestamp != "" {
		noteHeader += " #captured:" + timestamp
	}
	tb.AddToDump(noteHeader)

	for _, line := range lines {
		tb.AddToDump("    " + line)
	}
	tb.AddToDump("") // Empty line after note
}

// ReadDumpFile returns the contents of the dump file
func (tb *TestBrain) ReadDumpFile() string {
	tb.T.Helper()

	data, err := os.ReadFile(tb.DumpPath)
	if err != nil {
		tb.T.Fatalf("Failed to read dump file: %v", err)
	}
	return string(data)
}

// ReadConfigRaw reads the config file as a map (avoids circular dependency)
func (tb *TestBrain) ReadConfigRaw() map[string]interface{} {
	tb.T.Helper()

	data, err := os.ReadFile(tb.ConfigPath)
	if err != nil {
		tb.T.Fatalf("Failed to read config: %v", err)
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(data, &cfg); err != nil {
		tb.T.Fatalf("Failed to parse config: %v", err)
	}
	return cfg
}

// TouchFile updates the modification time of a file (creates if doesn't exist)
func (tb *TestBrain) TouchFile(path string) {
	tb.T.Helper()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create file
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			tb.T.Fatalf("Failed to create file %s: %v", path, err)
		}
	} else {
		// Update mtime
		now := os.FileMode(0644)
		if err := os.Chmod(path, now); err != nil {
			tb.T.Fatalf("Failed to touch file %s: %v", path, err)
		}
	}
}

// FileExists checks if a file exists
func (tb *TestBrain) FileExists(path string) bool {
	tb.T.Helper()

	_, err := os.Stat(path)
	return err == nil
}

// DirExists checks if a directory exists
func (tb *TestBrain) DirExists(path string) bool {
	tb.T.Helper()

	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// WriteFile writes content to a file (creates parent dirs if needed)
func (tb *TestBrain) WriteFile(path string, content string) {
	tb.T.Helper()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		tb.T.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		tb.T.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// ReadFile reads the contents of a file
func (tb *TestBrain) ReadFile(path string) string {
	tb.T.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		tb.T.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(data)
}
