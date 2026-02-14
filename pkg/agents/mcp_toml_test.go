package agents

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func newTestTOMLInstaller(t *testing.T) (*TOMLMCPInstaller, string) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	return &TOMLMCPInstaller{
		ConfigPath:  configPath,
		SectionName: "mcp_servers",
	}, configPath
}

func TestTOMLMCPInstaller_Install(t *testing.T) {
	installer, configPath := newTestTOMLInstaller(t)

	if err := installer.Install("test-server", "/usr/bin/test"); err != nil {
		t.Fatalf("Install error: %v", err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)

	if !strings.Contains(content, `[mcp_servers.test-server]`) {
		t.Error("section header missing")
	}
	if !strings.Contains(content, `command = "/usr/bin/test"`) {
		t.Error("command line missing")
	}
}

func TestTOMLMCPInstaller_IsInstalled(t *testing.T) {
	installer, _ := newTestTOMLInstaller(t)

	installed, _ := installer.IsInstalled("test")
	if installed {
		t.Error("expected not installed when file missing")
	}

	if err := installer.Install("test", "/bin/test"); err != nil {
		t.Fatal(err)
	}
	installed, _ = installer.IsInstalled("test")
	if !installed {
		t.Error("expected installed after Install")
	}
}

func TestTOMLMCPInstaller_Remove(t *testing.T) {
	installer, _ := newTestTOMLInstaller(t)

	// Remove from non-existent file — no error
	if err := installer.Remove("test"); err != nil {
		t.Fatalf("Remove error on missing file: %v", err)
	}

	if err := installer.Install("test", "/bin/test"); err != nil {
		t.Fatal(err)
	}
	if err := installer.Remove("test"); err != nil {
		t.Fatalf("Remove error: %v", err)
	}

	installed, _ := installer.IsInstalled("test")
	if installed {
		t.Error("expected not installed after Remove")
	}
}

func TestTOMLMCPInstaller_PreservesExisting(t *testing.T) {
	installer, configPath := newTestTOMLInstaller(t)

	existing := "[some_section]\nkey = \"value\"\n\n[mcp_servers.existing]\ncommand = \"/usr/bin/existing\"\n"
	if err := os.WriteFile(configPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	if err := installer.Install("new-server", "/usr/bin/new"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)

	if !strings.Contains(content, `[some_section]`) {
		t.Error("existing section was lost")
	}
	if !strings.Contains(content, `[mcp_servers.existing]`) {
		t.Error("existing server was lost")
	}
	if !strings.Contains(content, `[mcp_servers.new-server]`) {
		t.Error("new server not added")
	}
}

func TestTOMLMCPInstaller_UpdateExisting(t *testing.T) {
	installer, configPath := newTestTOMLInstaller(t)

	if err := installer.Install("test", "/old/path"); err != nil {
		t.Fatal(err)
	}
	if err := installer.Install("test", "/new/path"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(configPath)
	content := string(data)

	if strings.Contains(content, "/old/path") {
		t.Error("old path should have been replaced")
	}
	if !strings.Contains(content, `command = "/new/path"`) {
		t.Error("new path not found")
	}
	if strings.Count(content, "[mcp_servers.test]") != 1 {
		t.Error("duplicate sections")
	}
}
