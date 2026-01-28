package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestLoad(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Current != "test" {
		t.Errorf("Expected current brain 'test', got '%s'", cfg.Current)
	}

	if len(cfg.Brains) != 1 {
		t.Errorf("Expected 1 brain, got %d", len(cfg.Brains))
	}

	brain, exists := cfg.Brains["test"]
	if !exists {
		t.Fatal("Expected brain 'test' to exist")
	}

	if brain.Path != tb.BrainPath {
		t.Errorf("Expected path '%s', got '%s'", tb.BrainPath, brain.Path)
	}
}

func TestLoad_NonExistentConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	t.Setenv("BRAIN_CONFIG_PATH", configPath)
	t.Setenv("BRAIN_CONFIG_DIR", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed on non-existent config: %v", err)
	}

	// Should create empty config
	if cfg.Current != "" {
		t.Errorf("Expected empty current brain, got '%s'", cfg.Current)
	}

	if len(cfg.Brains) != 0 {
		t.Errorf("Expected 0 brains, got %d", len(cfg.Brains))
	}

	// Verify config file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	t.Setenv("BRAIN_CONFIG_PATH", configPath)
	t.Setenv("BRAIN_CONFIG_DIR", tmpDir)

	// Write invalid JSON
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte("invalid json {]"), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err := Load()
	if err == nil {
		t.Error("Expected error on invalid JSON, got nil")
	}
}

func TestSave(t *testing.T) {
	tb := testutil.SetupTestBrain(t)

	cfg := &Config{
		Current: "new-brain",
		Brains: map[string]*BrainInfo{
			"new-brain": {
				Path:    "/path/to/brain",
				Created: "2024-01-15",
				Focus:   "project-x",
			},
		},
	}

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if !tb.FileExists(tb.ConfigPath) {
		t.Error("Config file was not created")
	}

	// Verify contents
	data, err := os.ReadFile(tb.ConfigPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to parse saved config: %v", err)
	}

	if loaded.Current != "new-brain" {
		t.Errorf("Expected current 'new-brain', got '%s'", loaded.Current)
	}

	brain := loaded.Brains["new-brain"]
	if brain == nil {
		t.Fatal("Brain 'new-brain' not found in saved config")
	}

	if brain.Focus != "project-x" {
		t.Errorf("Expected focus 'project-x', got '%s'", brain.Focus)
	}
}

func TestAddBrain(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	newBrainPath := filepath.Join(tb.TmpDir, "new-brain")
	if err := cfg.AddBrain("work", newBrainPath); err != nil {
		t.Fatalf("AddBrain failed: %v", err)
	}

	brain, exists := cfg.Brains["work"]
	if !exists {
		t.Fatal("Brain 'work' not found after adding")
	}

	if brain.Path != newBrainPath {
		t.Errorf("Expected path '%s', got '%s'", newBrainPath, brain.Path)
	}

	if brain.Created == "" {
		t.Error("Created date not set")
	}
}

func TestSetCurrentBrain(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add a second brain
	newBrainPath := filepath.Join(tb.TmpDir, "work")
	if err := cfg.AddBrain("work", newBrainPath); err != nil {
		t.Fatalf("AddBrain failed: %v", err)
	}

	// Switch to it
	if err := cfg.SetCurrentBrain("work"); err != nil {
		t.Fatalf("SetCurrentBrain failed: %v", err)
	}

	if cfg.Current != "work" {
		t.Errorf("Expected current brain 'work', got '%s'", cfg.Current)
	}

	// Try to switch to non-existent brain
	if err := cfg.SetCurrentBrain("nonexistent"); err == nil {
		t.Error("Expected error when switching to non-existent brain")
	}
}

func TestSetFocusedProject(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	if err := cfg.SetFocusedProject("auth-system"); err != nil {
		t.Fatalf("SetFocusedProject failed: %v", err)
	}

	focus := cfg.GetFocusedProject()
	if focus != "auth-system" {
		t.Errorf("Expected focus 'auth-system', got '%s'", focus)
	}

	// Verify it's saved in the brain info
	brain := cfg.Brains["test"]
	if brain.Focus != "auth-system" {
		t.Errorf("Expected brain focus 'auth-system', got '%s'", brain.Focus)
	}
}

func TestGetFocusedProject(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Initially no focus
	focus := cfg.GetFocusedProject()
	if focus != "" {
		t.Errorf("Expected no focus, got '%s'", focus)
	}

	// Set focus
	cfg.Brains["test"].Focus = "project-x"
	focus = cfg.GetFocusedProject()
	if focus != "project-x" {
		t.Errorf("Expected focus 'project-x', got '%s'", focus)
	}
}

func TestRenameBrain(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	newPath := filepath.Join(tb.TmpDir, "renamed-brain")
	if err := cfg.RenameBrain("test", "renamed", newPath); err != nil {
		t.Fatalf("RenameBrain failed: %v", err)
	}

	// Old brain should not exist
	if _, exists := cfg.Brains["test"]; exists {
		t.Error("Old brain 'test' still exists after rename")
	}

	// New brain should exist
	brain, exists := cfg.Brains["renamed"]
	if !exists {
		t.Fatal("New brain 'renamed' not found after rename")
	}

	if brain.Path != newPath {
		t.Errorf("Expected path '%s', got '%s'", newPath, brain.Path)
	}

	// Current should be updated
	if cfg.Current != "renamed" {
		t.Errorf("Expected current brain 'renamed', got '%s'", cfg.Current)
	}
}

func TestRenameBrain_NonExistent(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	err := cfg.RenameBrain("nonexistent", "new", "/path")
	if err == nil {
		t.Error("Expected error when renaming non-existent brain")
	}
}

func TestRenameBrain_NameConflict(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add a second brain
	_ = cfg.AddBrain("work", "/path/to/work")

	// Try to rename test to work (conflict)
	err := cfg.RenameBrain("test", "work", "/new/path")
	if err == nil {
		t.Error("Expected error when renaming to existing brain name")
	}
}

func TestDeleteBrain(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add another brain
	_ = cfg.AddBrain("work", "/path/to/work")

	// Delete test brain
	if err := cfg.DeleteBrain("test"); err != nil {
		t.Fatalf("DeleteBrain failed: %v", err)
	}

	// Should not exist
	if _, exists := cfg.Brains["test"]; exists {
		t.Error("Brain 'test' still exists after deletion")
	}

	// Current should be cleared
	if cfg.Current != "" {
		t.Errorf("Expected current brain to be cleared, got '%s'", cfg.Current)
	}
}

func TestDeleteBrain_NonCurrent(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add another brain
	_ = cfg.AddBrain("work", "/path/to/work")

	// Delete non-current brain
	if err := cfg.DeleteBrain("work"); err != nil {
		t.Fatalf("DeleteBrain failed: %v", err)
	}

	// Current should still be test
	if cfg.Current != "test" {
		t.Errorf("Expected current brain 'test', got '%s'", cfg.Current)
	}
}

func TestDeleteBrain_NonExistent(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	err := cfg.DeleteBrain("nonexistent")
	if err == nil {
		t.Error("Expected error when deleting non-existent brain")
	}
}

func TestListBrains(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	// Add more brains
	_ = cfg.AddBrain("work", "/path/to/work")
	_ = cfg.AddBrain("personal", "/path/to/personal")

	brains := cfg.ListBrains()
	if len(brains) != 3 {
		t.Errorf("Expected 3 brains, got %d", len(brains))
	}

	// Check all brains are present
	brainMap := make(map[string]bool)
	for _, name := range brains {
		brainMap[name] = true
	}

	expected := []string{"test", "work", "personal"}
	for _, name := range expected {
		if !brainMap[name] {
			t.Errorf("Brain '%s' not in list", name)
		}
	}
}

func TestBrainExists(t *testing.T) {
	_ = testutil.SetupTestBrain(t)
	cfg, _ := Load()

	if !cfg.BrainExists("test") {
		t.Error("Brain 'test' should exist")
	}

	if cfg.BrainExists("nonexistent") {
		t.Error("Brain 'nonexistent' should not exist")
	}
}

func TestGetBrain(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	brain, exists := cfg.GetBrain("test")
	if !exists {
		t.Fatal("Brain 'test' should exist")
	}

	if brain.Path != tb.BrainPath {
		t.Errorf("Expected path '%s', got '%s'", tb.BrainPath, brain.Path)
	}

	_, exists = cfg.GetBrain("nonexistent")
	if exists {
		t.Error("Brain 'nonexistent' should not exist")
	}
}

func TestGetBrainPath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	path, err := cfg.GetBrainPath("test")
	if err != nil {
		t.Fatalf("GetBrainPath failed: %v", err)
	}

	if path != tb.BrainPath {
		t.Errorf("Expected path '%s', got '%s'", tb.BrainPath, path)
	}

	_, err = cfg.GetBrainPath("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent brain")
	}
}

func TestGetCurrentBrainPath(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()

	path, err := cfg.GetCurrentBrainPath()
	if err != nil {
		t.Fatalf("GetCurrentBrainPath failed: %v", err)
	}

	if path != tb.BrainPath {
		t.Errorf("Expected path '%s', got '%s'", tb.BrainPath, path)
	}
}

func TestThreadSafety(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	cfg, _ := Load()
	_ = tb // Used below for path generation

	// Add some brains
	for i := 0; i < 5; i++ {
		name := string(rune('a' + i))
		_ = cfg.AddBrain(name, filepath.Join(tb.TmpDir, name))
	}

	// Concurrent operations
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = cfg.GetCurrentBrain()
			_ = cfg.ListBrains()
			_, _ = cfg.GetBrain("a")
			_ = cfg.GetFocusedProject()
		}()
	}

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		brainName := string(rune('a' + i%5))
		go func(name string) {
			defer wg.Done()
			_ = cfg.SetFocusedProject("project-" + name)
		}(brainName)
	}

	wg.Wait()

	// Should not panic and config should be valid
	if len(cfg.Brains) < 5 {
		t.Errorf("Expected at least 5 brains, got %d", len(cfg.Brains))
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	tmpDir := t.TempDir()
	customConfigPath := filepath.Join(tmpDir, "custom-config.json")

	t.Setenv("BRAIN_CONFIG_PATH", customConfigPath)
	t.Setenv("BRAIN_CONFIG_DIR", tmpDir)

	// GetConfigFile should return custom path
	configPath := GetConfigFile()
	if configPath != customConfigPath {
		t.Errorf("Expected config path '%s', got '%s'", customConfigPath, configPath)
	}

	// Load should use custom path
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// File should be created at custom location
	if _, err := os.Stat(customConfigPath); os.IsNotExist(err) {
		t.Error("Config file not created at custom path")
	}

	// Save should use custom path
	cfg.Current = "test-brain"
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify saved to custom path
	data, err := os.ReadFile(customConfigPath)
	if err != nil {
		t.Fatalf("Failed to read custom config: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to parse config: %v", err)
	}

	if loaded.Current != "test-brain" {
		t.Errorf("Expected current 'test-brain', got '%s'", loaded.Current)
	}
}
