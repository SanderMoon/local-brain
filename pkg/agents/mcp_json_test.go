package agents

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func newTestJSONInstaller(t *testing.T) (*JSONMCPInstaller, string) {
	t.Helper()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	return &JSONMCPInstaller{
		ConfigPath: configPath,
		ServerKey:  "mcpServers",
		BuildEntry: func(binaryPath string) interface{} {
			return map[string]interface{}{
				"command": binaryPath,
			}
		},
	}, configPath
}

func TestJSONMCPInstaller_Install(t *testing.T) {
	installer, configPath := newTestJSONInstaller(t)

	if err := installer.Install("test-server", "/usr/bin/test"); err != nil {
		t.Fatalf("Install error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("reading config: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("parsing config: %v", err)
	}

	servers, _ := config["mcpServers"].(map[string]interface{})
	if servers == nil {
		t.Fatal("mcpServers key missing")
	}
	entry, _ := servers["test-server"].(map[string]interface{})
	if entry == nil {
		t.Fatal("test-server entry missing")
	}
	if entry["command"] != "/usr/bin/test" {
		t.Errorf("command: got %v, want /usr/bin/test", entry["command"])
	}
}

func TestJSONMCPInstaller_IsInstalled(t *testing.T) {
	installer, _ := newTestJSONInstaller(t)

	installed, err := installer.IsInstalled("test")
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if installed {
		t.Error("expected not installed when file missing")
	}

	if err := installer.Install("test", "/bin/test"); err != nil {
		t.Fatal(err)
	}
	installed, err = installer.IsInstalled("test")
	if err != nil {
		t.Fatalf("IsInstalled error: %v", err)
	}
	if !installed {
		t.Error("expected installed after Install")
	}
}

func TestJSONMCPInstaller_Remove(t *testing.T) {
	installer, _ := newTestJSONInstaller(t)

	// Remove from non-existent file — no error
	if err := installer.Remove("test"); err != nil {
		t.Fatalf("Remove error on missing file: %v", err)
	}

	// Install then remove
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

func TestJSONMCPInstaller_PreservesExistingConfig(t *testing.T) {
	installer, configPath := newTestJSONInstaller(t)

	// Write existing config with another server and a separate key
	existing := map[string]interface{}{
		"someOtherKey": "value",
		"mcpServers": map[string]interface{}{
			"existing-server": map[string]interface{}{
				"command": "/usr/bin/existing",
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Install new server
	if err := installer.Install("new-server", "/usr/bin/new"); err != nil {
		t.Fatal(err)
	}

	readData, _ := os.ReadFile(configPath)
	var config map[string]interface{}
	if err := json.Unmarshal(readData, &config); err != nil {
		t.Fatal(err)
	}

	if config["someOtherKey"] != "value" {
		t.Error("existing config key was lost")
	}
	servers := config["mcpServers"].(map[string]interface{})
	if _, ok := servers["existing-server"]; !ok {
		t.Error("existing server was lost")
	}
	if _, ok := servers["new-server"]; !ok {
		t.Error("new server not added")
	}
}
