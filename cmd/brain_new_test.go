package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

// setupNewCmdEnv sets up isolated environment variables for brain new tests.
// Returns the brainRoot path (i.e., BRAIN_ROOT) and a cleanup function.
func setupNewCmdEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	brainRoot := filepath.Join(tmpDir, "brains")
	if err := os.MkdirAll(brainRoot, 0755); err != nil {
		t.Fatalf("Failed to create brainRoot: %v", err)
	}

	configDir := filepath.Join(tmpDir, "config")
	configPath := filepath.Join(configDir, "config.json")
	symlinkPath := filepath.Join(tmpDir, "brain-link")

	t.Setenv("BRAIN_ROOT", brainRoot)
	t.Setenv("BRAIN_CONFIG_DIR", configDir)
	t.Setenv("BRAIN_CONFIG_PATH", configPath)
	t.Setenv("BRAIN_SYMLINK", symlinkPath)

	// Create config directory (config.Load() needs it to exist or will create it)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	return brainRoot
}

func TestRunNew_CreatesTemplatesDir(t *testing.T) {
	brainRoot := setupNewCmdEnv(t)

	if err := runNew(newCmd, []string{"testbrain"}); err != nil {
		t.Fatalf("runNew failed: %v", err)
	}

	templatesDir := filepath.Join(brainRoot, "testbrain", "_templates")
	info, err := os.Stat(templatesDir)
	if err != nil {
		t.Fatalf("Expected _templates/ directory to exist, got error: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected _templates to be a directory")
	}
}

func TestRunNew_TemplateFilesExist(t *testing.T) {
	brainRoot := setupNewCmdEnv(t)

	if err := runNew(newCmd, []string{"testbrain"}); err != nil {
		t.Fatalf("runNew failed: %v", err)
	}

	templatesDir := filepath.Join(brainRoot, "testbrain", "_templates")

	expectedFiles := []string{
		"new-note.md",
		"new-project.md",
		"daily-note.md",
	}

	for _, filename := range expectedFiles {
		filePath := filepath.Join(templatesDir, filename)

		content, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("Expected template file %s to exist, got error: %v", filename, err)
			continue
		}

		if len(content) == 0 {
			t.Errorf("Expected template file %s to have non-empty content", filename)
		}
	}
}

func TestRunNew_AdoptionPathNoTemplates(t *testing.T) {
	brainRoot := setupNewCmdEnv(t)

	// Pre-create a directory that looks like an existing brain (triggers adoption path)
	brainLocation := filepath.Join(brainRoot, "existingbrain")
	activeDir := filepath.Join(brainLocation, "01_active")
	dumpFile := filepath.Join(brainLocation, "00_dump.md")

	if err := os.MkdirAll(activeDir, 0755); err != nil {
		t.Fatalf("Failed to create active dir: %v", err)
	}
	if err := os.WriteFile(dumpFile, []byte("# Dump\n\n"), 0644); err != nil {
		t.Fatalf("Failed to create dump file: %v", err)
	}

	// Ensure _templates does NOT exist before adoption
	templatesDir := filepath.Join(brainLocation, "_templates")
	if _, err := os.Stat(templatesDir); err == nil {
		t.Fatalf("_templates dir should not exist before adoption test")
	}

	// Run brain new — should adopt, not create new structure
	if err := runNew(newCmd, []string{"existingbrain"}); err != nil {
		t.Fatalf("runNew failed on adoption path: %v", err)
	}

	// Verify _templates was NOT created by the adoption path
	if _, err := os.Stat(templatesDir); err == nil {
		t.Errorf("Adoption path should NOT create _templates/ directory")
	}
}
