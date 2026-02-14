package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupSkillTest creates a temp HOME with optional agent config directories.
func setupSkillTest(t *testing.T, agentDirs ...string) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	for _, dir := range agentDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return tmpDir
}

func TestSkillList_OutputsAllSkills(t *testing.T) {
	setupSkillTest(t)

	var buf bytes.Buffer
	skillListCmd.SetOut(&buf)
	if err := runSkillList(skillListCmd, nil); err != nil {
		t.Fatalf("runSkillList error: %v", err)
	}

	out := buf.String()
	expected := []string{
		"brain-capture", "brain-daily", "brain-focus",
		"brain-plan", "brain-review", "brain-setup", "brain-triage",
	}
	for _, name := range expected {
		if !strings.Contains(out, name) {
			t.Errorf("expected %q in list output", name)
		}
	}
}

func TestSkillInstall_InstallsToDetectedAgent(t *testing.T) {
	setupSkillTest(t, ".claude")

	var buf bytes.Buffer
	skillInstallCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	origForce := skillForceFlag
	skillAgentFlag = "all"
	skillForceFlag = false
	defer func() { skillAgentFlag = origAgent; skillForceFlag = origForce }()

	if err := runSkillInstall(skillInstallCmd, []string{"brain-daily"}); err != nil {
		t.Fatalf("runSkillInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Installed brain-daily") {
		t.Errorf("expected install message, got: %s", out)
	}
	if !strings.Contains(out, "Claude Code") {
		t.Errorf("expected Claude Code in output, got: %s", out)
	}
}

func TestSkillInstall_SkipsExisting(t *testing.T) {
	tmpDir := setupSkillTest(t, ".claude")

	// Pre-install a skill
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "brain-daily")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("existing"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	skillInstallCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	origForce := skillForceFlag
	skillAgentFlag = "all"
	skillForceFlag = false
	defer func() { skillAgentFlag = origAgent; skillForceFlag = origForce }()

	if err := runSkillInstall(skillInstallCmd, []string{"brain-daily"}); err != nil {
		t.Fatalf("runSkillInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SKIP:") {
		t.Errorf("expected SKIP message, got: %s", out)
	}
	if !strings.Contains(out, "--force") {
		t.Errorf("expected --force hint, got: %s", out)
	}
}

func TestSkillInstall_ForceOverwrites(t *testing.T) {
	tmpDir := setupSkillTest(t, ".claude")

	// Pre-install with old content
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "brain-daily")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	skillInstallCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	origForce := skillForceFlag
	skillAgentFlag = "all"
	skillForceFlag = true
	defer func() { skillAgentFlag = origAgent; skillForceFlag = origForce }()

	if err := runSkillInstall(skillInstallCmd, []string{"brain-daily"}); err != nil {
		t.Fatalf("runSkillInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Installed") {
		t.Errorf("expected install message with --force, got: %s", out)
	}

	// Verify content was actually overwritten
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) == "old content" {
		t.Error("file was not overwritten with --force")
	}
}

func TestSkillInstall_NoAgentsDetected(t *testing.T) {
	setupSkillTest(t) // no agent dirs

	var buf bytes.Buffer
	skillInstallCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	skillAgentFlag = "all"
	defer func() { skillAgentFlag = origAgent }()

	if err := runSkillInstall(skillInstallCmd, nil); err != nil {
		t.Fatalf("runSkillInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No AI agents detected") {
		t.Errorf("expected 'No AI agents detected', got: %s", out)
	}
}

func TestSkillInstall_UnknownSkill(t *testing.T) {
	setupSkillTest(t, ".claude")

	origAgent := skillAgentFlag
	skillAgentFlag = "all"
	defer func() { skillAgentFlag = origAgent }()

	err := runSkillInstall(skillInstallCmd, []string{"nonexistent-skill"})
	if err == nil {
		t.Fatal("expected error for unknown skill")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestSkillInstall_UnknownAgent(t *testing.T) {
	setupSkillTest(t)

	origAgent := skillAgentFlag
	skillAgentFlag = "nonexistent"
	defer func() { skillAgentFlag = origAgent }()

	err := runSkillInstall(skillInstallCmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected 'unknown agent' error, got: %v", err)
	}
}

func TestSkillRemove_RemovesInstalled(t *testing.T) {
	tmpDir := setupSkillTest(t, ".claude")

	// Pre-install
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "brain-daily")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	skillRemoveCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	skillAgentFlag = "all"
	defer func() { skillAgentFlag = origAgent }()

	if err := runSkillRemove(skillRemoveCmd, []string{"brain-daily"}); err != nil {
		t.Fatalf("runSkillRemove error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Removed brain-daily from Claude Code") {
		t.Errorf("expected removal message, got: %s", out)
	}
}

func TestSkillRemove_SkipsNotInstalled(t *testing.T) {
	setupSkillTest(t, ".claude")

	var buf bytes.Buffer
	skillRemoveCmd.SetOut(&buf)

	origAgent := skillAgentFlag
	skillAgentFlag = "all"
	defer func() { skillAgentFlag = origAgent }()

	if err := runSkillRemove(skillRemoveCmd, []string{"brain-daily"}); err != nil {
		t.Fatalf("runSkillRemove error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SKIP:") {
		t.Errorf("expected SKIP message, got: %s", out)
	}
}

func TestSkillUpgrade_UpgradesInstalled(t *testing.T) {
	tmpDir := setupSkillTest(t, ".claude")

	// Pre-install with old content
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "brain-daily")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("old version"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	skillUpgradeCmd.SetOut(&buf)

	origAgent := upgradeAgentFlag
	upgradeAgentFlag = "all"
	defer func() { upgradeAgentFlag = origAgent }()

	if err := runSkillUpgrade(skillUpgradeCmd, nil); err != nil {
		t.Fatalf("runSkillUpgrade error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Upgraded brain-daily") {
		t.Errorf("expected upgrade message, got: %s", out)
	}
	if !strings.Contains(out, "local modifications") {
		t.Errorf("expected overwrite warning, got: %s", out)
	}

	// Verify content was updated
	data, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if string(data) == "old version" {
		t.Error("file was not upgraded")
	}
}

func TestSkillUpgrade_NothingInstalled(t *testing.T) {
	setupSkillTest(t, ".claude")

	var buf bytes.Buffer
	skillUpgradeCmd.SetOut(&buf)

	origAgent := upgradeAgentFlag
	upgradeAgentFlag = "all"
	defer func() { upgradeAgentFlag = origAgent }()

	if err := runSkillUpgrade(skillUpgradeCmd, nil); err != nil {
		t.Fatalf("runSkillUpgrade error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No installed skills found") {
		t.Errorf("expected 'no installed skills' message, got: %s", out)
	}
}

func TestSkillStatus_MixedState(t *testing.T) {
	tmpDir := setupSkillTest(t, ".claude")

	// Install one skill for Claude
	skillDir := filepath.Join(tmpDir, ".claude", "skills", "brain-daily")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	skillStatusCmd.SetOut(&buf)
	if err := runSkillStatus(skillStatusCmd, nil); err != nil {
		t.Fatalf("runSkillStatus error: %v", err)
	}

	out := buf.String()
	// Header should have Skill and agent names
	if !strings.Contains(out, "Skill") {
		t.Error("expected header with 'Skill'")
	}
	// brain-daily should be installed for Claude
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "brain-daily") {
			if !strings.Contains(line, "installed") {
				t.Errorf("brain-daily should show 'installed' for Claude, got: %s", line)
			}
			break
		}
	}
	// Codex is not detected, should show "-"
	for _, line := range lines {
		if strings.Contains(line, "brain-capture") {
			// Codex column should be "-", Claude should be "not installed"
			if !strings.Contains(line, "-") {
				t.Errorf("undetected agent should show '-', got: %s", line)
			}
			break
		}
	}
}
