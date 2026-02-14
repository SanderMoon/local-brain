package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupMCPTest creates a temp HOME with a fake brain-mcp binary and
// optional agent config directories. Returns the temp dir.
func setupMCPTest(t *testing.T, agentDirs ...string) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create a fake brain-mcp binary so FindMCPBinary works via PATH
	binDir := filepath.Join(tmpDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	fakeBinary := filepath.Join(binDir, "brain-mcp")
	if err := os.WriteFile(fakeBinary, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	// Create requested agent config dirs
	for _, dir := range agentDirs {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0755); err != nil {
			t.Fatal(err)
		}
	}
	return tmpDir
}

func TestMCPStatus_NoAgentsDetected(t *testing.T) {
	setupMCPTest(t) // no agent dirs

	var buf bytes.Buffer
	mcpStatusCmd.SetOut(&buf)
	if err := runMCPStatus(mcpStatusCmd, nil); err != nil {
		t.Fatalf("runMCPStatus error: %v", err)
	}

	out := buf.String()
	// All agents should show "-" when not detected
	if !strings.Contains(out, "Claude Code") {
		t.Error("expected Claude Code in output")
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines[2:] { // skip header and separator
		if !strings.Contains(line, "-") {
			t.Errorf("expected '-' for undetected agent, got: %s", line)
		}
	}
}

func TestMCPStatus_MixedState(t *testing.T) {
	tmpDir := setupMCPTest(t, ".claude", ".gemini")

	// Register MCP for Claude only
	claudeConfig := filepath.Join(tmpDir, ".claude.json")
	if err := os.WriteFile(claudeConfig, []byte(`{"mcpServers":{"local-brain":{"type":"stdio","command":"brain-mcp"}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	mcpStatusCmd.SetOut(&buf)
	if err := runMCPStatus(mcpStatusCmd, nil); err != nil {
		t.Fatalf("runMCPStatus error: %v", err)
	}

	out := buf.String()
	lines := strings.Split(strings.TrimSpace(out), "\n")

	found := map[string]string{}
	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Agent name may be two words ("Claude Code", "Gemini CLI")
			name := strings.Join(fields[:len(fields)-1], " ")
			status := fields[len(fields)-1]
			found[name] = status
		}
	}

	if found["Claude Code"] != "registered" {
		t.Errorf("Claude Code: want registered, got %s", found["Claude Code"])
	}
	if found["Gemini CLI"] != "registered" && found["Gemini"] != "registered" {
		// Gemini has no config written, so should be "not"(as part of "not registered")
		// Check for "not" since fields split may break "not registered"
		for _, line := range lines[2:] {
			if strings.Contains(line, "Gemini") && !strings.Contains(line, "not registered") {
				t.Errorf("Gemini CLI should be 'not registered', got line: %s", line)
			}
		}
	}
	if found["Codex"] != "-" {
		t.Errorf("Codex: want -, got %s", found["Codex"])
	}
}

func TestMCPInstall_RegistersNewAgent(t *testing.T) {
	setupMCPTest(t, ".gemini")

	var buf bytes.Buffer
	mcpInstallCmd.SetOut(&buf)

	orig := mcpAgentFlag
	mcpAgentFlag = "all"
	defer func() { mcpAgentFlag = orig }()

	if err := runMCPInstall(mcpInstallCmd, nil); err != nil {
		t.Fatalf("runMCPInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Registered local-brain") {
		t.Errorf("expected registration message, got: %s", out)
	}
	if !strings.Contains(out, "Gemini CLI") {
		t.Errorf("expected Gemini CLI in output, got: %s", out)
	}
}

func TestMCPInstall_SkipsAlreadyRegistered(t *testing.T) {
	tmpDir := setupMCPTest(t, ".gemini")

	// Pre-register
	geminiConfig := filepath.Join(tmpDir, ".gemini", "settings.json")
	if err := os.WriteFile(geminiConfig, []byte(`{"mcpServers":{"local-brain":{"command":"brain-mcp"}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	mcpInstallCmd.SetOut(&buf)

	orig := mcpAgentFlag
	mcpAgentFlag = "all"
	defer func() { mcpAgentFlag = orig }()

	if err := runMCPInstall(mcpInstallCmd, nil); err != nil {
		t.Fatalf("runMCPInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SKIP:") {
		t.Errorf("expected SKIP message for already-registered agent, got: %s", out)
	}
}

func TestMCPInstall_NoAgentsDetected(t *testing.T) {
	setupMCPTest(t) // no agent dirs

	var buf bytes.Buffer
	mcpInstallCmd.SetOut(&buf)

	orig := mcpAgentFlag
	mcpAgentFlag = "all"
	defer func() { mcpAgentFlag = orig }()

	if err := runMCPInstall(mcpInstallCmd, nil); err != nil {
		t.Fatalf("runMCPInstall error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "No AI agents detected") {
		t.Errorf("expected 'No AI agents detected', got: %s", out)
	}
}

func TestMCPRemove_RemovesRegistered(t *testing.T) {
	tmpDir := setupMCPTest(t, ".gemini")

	// Pre-register
	geminiConfig := filepath.Join(tmpDir, ".gemini", "settings.json")
	if err := os.WriteFile(geminiConfig, []byte(`{"mcpServers":{"local-brain":{"command":"brain-mcp"}}}`), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	mcpRemoveCmd.SetOut(&buf)

	orig := mcpAgentFlag
	mcpAgentFlag = "all"
	defer func() { mcpAgentFlag = orig }()

	if err := runMCPRemove(mcpRemoveCmd, nil); err != nil {
		t.Fatalf("runMCPRemove error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "OK:   Removed local-brain from Gemini CLI") {
		t.Errorf("expected removal message, got: %s", out)
	}
}

func TestMCPRemove_SkipsUnregistered(t *testing.T) {
	setupMCPTest(t, ".gemini")

	var buf bytes.Buffer
	mcpRemoveCmd.SetOut(&buf)

	orig := mcpAgentFlag
	mcpAgentFlag = "all"
	defer func() { mcpAgentFlag = orig }()

	if err := runMCPRemove(mcpRemoveCmd, nil); err != nil {
		t.Fatalf("runMCPRemove error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "SKIP:") {
		t.Errorf("expected SKIP message for unregistered agent, got: %s", out)
	}
}

func TestMCPInstall_UnknownAgent(t *testing.T) {
	setupMCPTest(t)

	orig := mcpAgentFlag
	mcpAgentFlag = "nonexistent"
	defer func() { mcpAgentFlag = orig }()

	err := runMCPInstall(mcpInstallCmd, nil)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
	if !strings.Contains(err.Error(), "unknown agent") {
		t.Errorf("expected 'unknown agent' error, got: %v", err)
	}
}
