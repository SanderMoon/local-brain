package skillscatalog

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseFrontmatter tests frontmatter parsing with various inputs.
func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantName    string
		wantDesc    string
	}{
		{
			name: "valid frontmatter",
			input: `---
name: my-skill
description: Does something useful.
---
# Body`,
			wantName: "my-skill",
			wantDesc: "Does something useful.",
		},
		{
			name: "quoted values",
			input: `---
name: "quoted-skill"
description: "A quoted description."
---`,
			wantName: "quoted-skill",
			wantDesc: "A quoted description.",
		},
		{
			name: "single-quoted values",
			input: `---
name: 'single-quoted'
description: 'Another description.'
---`,
			wantName: "single-quoted",
			wantDesc: "Another description.",
		},
		{
			name: "missing description",
			input: `---
name: no-desc
---`,
			wantName: "no-desc",
			wantDesc: "",
		},
		{
			name: "missing name",
			input: `---
description: No name here.
---`,
			wantName: "",
			wantDesc: "No name here.",
		},
		{
			name:     "no frontmatter",
			input:    `# Just a heading\nNo frontmatter here.`,
			wantName: "",
			wantDesc: "",
		},
		{
			name:     "empty content",
			input:    "",
			wantName: "",
			wantDesc: "",
		},
		{
			name: "extra fields ignored",
			input: `---
name: skill-x
compatibility: Requires something.
description: Short desc.
---`,
			wantName: "skill-x",
			wantDesc: "Short desc.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotName, gotDesc := ParseFrontmatter(tt.input)
			if gotName != tt.wantName {
				t.Errorf("ParseFrontmatter name: got %q, want %q", gotName, tt.wantName)
			}
			if gotDesc != tt.wantDesc {
				t.Errorf("ParseFrontmatter description: got %q, want %q", gotDesc, tt.wantDesc)
			}
		})
	}
}

// TestListSkills verifies all bundled skills are present with valid frontmatter.
func TestListSkills(t *testing.T) {
	skills, err := ListSkills()
	if err != nil {
		t.Fatalf("ListSkills() error: %v", err)
	}

	expectedSkills := []string{
		"brain-capture",
		"brain-daily",
		"brain-focus",
		"brain-plan",
		"brain-review",
		"brain-setup",
		"brain-triage",
	}

	if len(skills) != len(expectedSkills) {
		t.Fatalf("ListSkills() returned %d skills, want %d", len(skills), len(expectedSkills))
	}

	skillMap := make(map[string]SkillInfo)
	for _, s := range skills {
		skillMap[s.Name] = s
	}

	for _, name := range expectedSkills {
		s, ok := skillMap[name]
		if !ok {
			t.Errorf("skill %q not found in list", name)
			continue
		}
		if s.Description == "" {
			t.Errorf("skill %q has empty description", name)
		}
		if s.Content == "" {
			t.Errorf("skill %q has empty content", name)
		}
	}
}

// TestGetSkillContent verifies content retrieval for brain-daily and error on unknown.
func TestGetSkillContent(t *testing.T) {
	content, err := GetSkillContent("brain-daily")
	if err != nil {
		t.Fatalf("GetSkillContent(brain-daily) error: %v", err)
	}
	if content == "" {
		t.Error("GetSkillContent returned empty content for brain-daily")
	}

	_, err = GetSkillContent("nonexistent-skill")
	if err == nil {
		t.Error("GetSkillContent should return error for unknown skill")
	}
}

// TestInstallSkill verifies file creation, skip-on-re-run, and force overwrite.
func TestInstallSkill(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	agent := AgentInfo{
		ID:        "claude",
		Name:      "Claude Code",
		ConfigDir: filepath.Join(tmpDir, ".claude"),
		SkillsDir: filepath.Join(tmpDir, ".claude", "skills"),
	}
	skill := SkillInfo{
		Name:        "test-skill",
		Description: "A test skill.",
		Content:     "# test content",
	}

	// First install — should succeed
	installed, err := InstallSkill(skill, agent, false)
	if err != nil {
		t.Fatalf("InstallSkill error: %v", err)
	}
	if !installed {
		t.Error("expected installed=true on first install")
	}

	dest := filepath.Join(agent.SkillsDir, skill.Name, "SKILL.md")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("SKILL.md not found after install: %v", err)
	}

	// Second install without force — should skip
	installed, err = InstallSkill(skill, agent, false)
	if err != nil {
		t.Fatalf("InstallSkill (skip) error: %v", err)
	}
	if installed {
		t.Error("expected installed=false when file exists and force=false")
	}

	// Install with force — should overwrite
	skill.Content = "# updated content"
	installed, err = InstallSkill(skill, agent, true)
	if err != nil {
		t.Fatalf("InstallSkill (force) error: %v", err)
	}
	if !installed {
		t.Error("expected installed=true when force=true")
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading installed file: %v", err)
	}
	if string(data) != "# updated content" {
		t.Errorf("file content after force install: got %q, want %q", string(data), "# updated content")
	}
}

// TestRemoveSkill verifies removal of an installed skill and no-op on absent skill.
func TestRemoveSkill(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	agent := AgentInfo{
		ID:        "claude",
		Name:      "Claude Code",
		ConfigDir: filepath.Join(tmpDir, ".claude"),
		SkillsDir: filepath.Join(tmpDir, ".claude", "skills"),
	}

	// Pre-create the skill file
	dest := filepath.Join(agent.SkillsDir, "test-skill", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dest, []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Remove it
	removed, err := RemoveSkill("test-skill", agent)
	if err != nil {
		t.Fatalf("RemoveSkill error: %v", err)
	}
	if !removed {
		t.Error("expected removed=true")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("SKILL.md should not exist after removal")
	}

	// Remove again — no error, removed=false
	removed, err = RemoveSkill("test-skill", agent)
	if err != nil {
		t.Fatalf("RemoveSkill (absent) error: %v", err)
	}
	if removed {
		t.Error("expected removed=false when skill not installed")
	}
}

// TestIsInstalled verifies IsInstalled returns false when absent, true after creation.
func TestIsInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	agent := AgentInfo{
		ID:        "claude",
		Name:      "Claude Code",
		ConfigDir: filepath.Join(tmpDir, ".claude"),
		SkillsDir: filepath.Join(tmpDir, ".claude", "skills"),
	}

	if IsInstalled("brain-daily", agent) {
		t.Error("expected IsInstalled=false when directory absent")
	}

	dest := filepath.Join(agent.SkillsDir, "brain-daily", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(dest, []byte("content"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if !IsInstalled("brain-daily", agent) {
		t.Error("expected IsInstalled=true after creating SKILL.md")
	}
}

// TestKnownAgents verifies that agent paths are constructed relative to $HOME.
func TestKnownAgents(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	agents := KnownAgents()
	if len(agents) != 4 {
		t.Fatalf("expected 4 known agents, got %d", len(agents))
	}

	want := map[string]struct {
		configDir string
		skillsDir string
	}{
		"claude":   {filepath.Join(tmpDir, ".claude"), filepath.Join(tmpDir, ".claude", "skills")},
		"opencode": {filepath.Join(tmpDir, ".config", "opencode"), filepath.Join(tmpDir, ".config", "opencode", "skills")},
		"codex":    {filepath.Join(tmpDir, ".codex"), filepath.Join(tmpDir, ".codex", "skills")},
		"gemini":   {filepath.Join(tmpDir, ".gemini"), filepath.Join(tmpDir, ".gemini", "skills")},
	}

	for _, a := range agents {
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

// TestDetectedAgents verifies that only agents whose ConfigDir exists are detected.
func TestDetectedAgents(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Initially no agents detected
	agents := DetectedAgents()
	if len(agents) != 0 {
		t.Errorf("expected 0 detected agents in empty tmpDir, got %d", len(agents))
	}

	// Create ~/.claude
	if err := os.MkdirAll(filepath.Join(tmpDir, ".claude"), 0755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}

	agents = DetectedAgents()
	if len(agents) != 1 {
		t.Errorf("expected 1 detected agent after creating .claude, got %d", len(agents))
	}
	if agents[0].ID != "claude" {
		t.Errorf("expected claude agent, got %s", agents[0].ID)
	}
}

// TestFindAgent verifies all 4 known IDs are findable and unknown IDs return errors.
func TestFindAgent(t *testing.T) {
	known := []string{"claude", "codex", "gemini", "opencode"}
	for _, id := range known {
		a, err := FindAgent(id)
		if err != nil {
			t.Errorf("FindAgent(%q) unexpected error: %v", id, err)
		}
		if a.ID != id {
			t.Errorf("FindAgent(%q) returned wrong ID: %s", id, a.ID)
		}
	}

	_, err := FindAgent("unknown-agent")
	if err == nil {
		t.Error("FindAgent should return error for unknown agent")
	}
}
