package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/config"
)

// setupConfigCmdEnv sets up isolated environment variables for config tests.
func setupConfigCmdEnv(t *testing.T) string {
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

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	return tmpDir
}

func TestTildify(t *testing.T) {
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME not set")
	}

	tests := []struct {
		input    string
		expected string
	}{
		{filepath.Join(home, "brains"), "~/brains"},
		{filepath.Join(home, "dev"), "~/dev"},
		{"/tmp/other", "/tmp/other"},
		{home, "~"},
	}

	for _, tt := range tests {
		result := tildify(tt.input)
		if result != tt.expected {
			t.Errorf("tildify(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestConfigSource(t *testing.T) {
	// No config, no env
	result := configSource("", "NONEXISTENT_VAR_12345")
	if result != "(default)" {
		t.Errorf("Expected '(default)', got %q", result)
	}

	// Has config value
	result = configSource("some-value", "")
	if result != "(configured)" {
		t.Errorf("Expected '(configured)', got %q", result)
	}

	// Has env var
	t.Setenv("TEST_CONFIG_SOURCE_VAR", "value")
	result = configSource("", "TEST_CONFIG_SOURCE_VAR")
	if result != "(env: TEST_CONFIG_SOURCE_VAR)" {
		t.Errorf("Expected '(env: TEST_CONFIG_SOURCE_VAR)', got %q", result)
	}
}

func TestRunConfig_ShowNoArgs(t *testing.T) {
	setupConfigCmdEnv(t)

	// Should not error when called with no args (shows table)
	err := runConfig(configCmd, []string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRunConfig_GetEditor(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"editor"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRunConfig_SetEditor(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"editor", "emacs"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify it was saved
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.GetEditor() != "emacs" {
		t.Errorf("Expected editor 'emacs', got %q", cfg.GetEditor())
	}
}

func TestRunConfig_GetDevDir(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"dev_dir"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRunConfig_SetDevDir(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"dev_dir", "~/projects"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	if cfg.GetRawDevDir() != "~/projects" {
		t.Errorf("Expected dev_dir '~/projects', got %q", cfg.GetRawDevDir())
	}
}

func TestRunConfig_GetBrainRoot(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"brain_root"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRunConfig_SetBrainRoot_ReadOnly(t *testing.T) {
	setupConfigCmdEnv(t)

	// Should not error, but should print a message about being read-only
	err := runConfig(configCmd, []string{"brain_root", "/new/path"})
	if err != nil {
		t.Fatalf("Expected no error (read-only message), got: %v", err)
	}
}

func TestRunConfig_GetSymlink(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"symlink"})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestRunConfig_SetSymlink_ReadOnly(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"symlink", "/new/path"})
	if err != nil {
		t.Fatalf("Expected no error (read-only message), got: %v", err)
	}
}

func TestRunConfig_UnknownKey(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfig(configCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("Expected error for unknown key, got nil")
	}
}

func TestRunConfigShow(t *testing.T) {
	setupConfigCmdEnv(t)

	err := runConfigShow(configShowCmd, []string{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}

func TestDetectEditorName(t *testing.T) {
	// Should return something or empty string, never panic
	result := detectEditorName()
	_ = result // just ensure no panic
}
