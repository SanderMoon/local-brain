package agents

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAll(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	all := All()
	if len(all) != 4 {
		t.Fatalf("expected 4 agents, got %d", len(all))
	}

	ids := map[string]bool{}
	for _, a := range all {
		ids[a.ID] = true
		if a.Name == "" {
			t.Errorf("agent %s has empty name", a.ID)
		}
		if a.MCP == nil {
			t.Errorf("agent %s has nil MCP installer", a.ID)
		}
	}

	for _, id := range []string{"claude", "codex", "gemini", "opencode"} {
		if !ids[id] {
			t.Errorf("missing agent: %s", id)
		}
	}
}

func TestDetected(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if got := Detected(); len(got) != 0 {
		t.Errorf("expected 0 detected in empty tmpDir, got %d", len(got))
	}

	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatal(err)
	}
	detected := Detected()
	if len(detected) != 1 {
		t.Fatalf("expected 1 detected, got %d", len(detected))
	}
	if detected[0].ID != "claude" {
		t.Errorf("expected claude, got %s", detected[0].ID)
	}
}

func TestFind(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	for _, id := range []string{"claude", "codex", "gemini", "opencode"} {
		a, err := Find(id)
		if err != nil {
			t.Errorf("Find(%q) error: %v", id, err)
		}
		if a.ID != id {
			t.Errorf("Find(%q) returned ID %s", id, a.ID)
		}
	}

	_, err := Find("unknown")
	if err == nil {
		t.Error("Find(unknown) should error")
	}
}

func TestAgentPaths(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	want := map[string]struct {
		configDir string
		skillsDir string
	}{
		"claude":   {filepath.Join(tmpDir, ".claude"), filepath.Join(tmpDir, ".claude", "skills")},
		"opencode": {filepath.Join(tmpDir, ".config", "opencode"), filepath.Join(tmpDir, ".config", "opencode", "skills")},
		"codex":    {filepath.Join(tmpDir, ".codex"), filepath.Join(tmpDir, ".codex", "skills")},
		"gemini":   {filepath.Join(tmpDir, ".gemini"), filepath.Join(tmpDir, ".gemini", "skills")},
	}

	for _, a := range All() {
		exp, ok := want[a.ID]
		if !ok {
			t.Errorf("unexpected agent ID: %s", a.ID)
			continue
		}
		if a.ConfigDir != exp.configDir {
			t.Errorf("agent %s ConfigDir: got %q, want %q", a.ID, a.ConfigDir, exp.configDir)
		}
		if a.SkillsDir != exp.skillsDir {
			t.Errorf("agent %s SkillsDir: got %q, want %q", a.ID, a.SkillsDir, exp.skillsDir)
		}
	}
}
