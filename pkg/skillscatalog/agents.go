package skillscatalog

import (
	"fmt"
	"os"
	"path/filepath"
)

// AgentInfo describes a supported AI coding agent.
type AgentInfo struct {
	ID        string // e.g. "claude"
	Name      string // human-readable, e.g. "Claude Code"
	ConfigDir string // absolute path — used for detection
	SkillsDir string // absolute path — install target
}

// KnownAgents returns all supported agents with paths resolved from HOME.
// Paths are resolved at call time so that tests can override $HOME via t.Setenv.
func KnownAgents() []AgentInfo {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}

	return []AgentInfo{
		{
			ID:        "claude",
			Name:      "Claude Code",
			ConfigDir: filepath.Join(home, ".claude"),
			SkillsDir: filepath.Join(home, ".claude", "skills"),
		},
		{
			ID:        "opencode",
			Name:      "OpenCode",
			ConfigDir: filepath.Join(home, ".config", "opencode"),
			SkillsDir: filepath.Join(home, ".config", "opencode", "skills"),
		},
		{
			ID:        "codex",
			Name:      "Codex",
			ConfigDir: filepath.Join(home, ".codex"),
			SkillsDir: filepath.Join(home, ".codex", "skills"),
		},
		{
			ID:        "gemini",
			Name:      "Gemini CLI",
			ConfigDir: filepath.Join(home, ".gemini"),
			SkillsDir: filepath.Join(home, ".gemini", "skills"),
		},
	}
}

// DetectedAgents returns only those agents whose ConfigDir exists on disk.
func DetectedAgents() []AgentInfo {
	var detected []AgentInfo
	for _, a := range KnownAgents() {
		if _, err := os.Stat(a.ConfigDir); err == nil {
			detected = append(detected, a)
		}
	}
	return detected
}

// FindAgent returns the AgentInfo for the given ID, or an error if unknown.
func FindAgent(id string) (AgentInfo, error) {
	for _, a := range KnownAgents() {
		if a.ID == id {
			return a, nil
		}
	}
	return AgentInfo{}, fmt.Errorf("unknown agent %q (valid: claude, codex, gemini, opencode)", id)
}
