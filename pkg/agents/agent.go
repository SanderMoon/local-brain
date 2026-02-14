// Package agents provides a shared registry of AI coding agents.
// It is used by both skill installation and MCP server registration.
//
// To add a new agent:
//  1. Create a new file (e.g., windsurf.go)
//  2. Define a newWindsurf(home string) Agent function
//  3. Add newWindsurf(home) to the All() list below
package agents

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Info holds basic identification for an AI coding agent.
type Info struct {
	ID        string // e.g. "claude"
	Name      string // human-readable, e.g. "Claude Code"
	ConfigDir string // absolute path — used for detection
	SkillsDir string // absolute path — skill install target
}

// MCPInstaller defines how an agent registers and manages MCP servers.
// Implement this interface to add MCP support for a new agent.
type MCPInstaller interface {
	// Install registers an MCP server with the agent.
	Install(serverName, binaryPath string) error
	// Remove unregisters an MCP server from the agent.
	Remove(serverName string) error
	// IsInstalled reports whether the named MCP server is registered.
	IsInstalled(serverName string) (bool, error)
}

// Agent combines basic agent info with MCP management capability.
type Agent struct {
	Info
	MCP MCPInstaller
}

// All returns all known agents with paths resolved from $HOME.
func All() []Agent {
	home := homeDir()
	return []Agent{
		newClaude(home),
		newCodex(home),
		newGemini(home),
		newOpenCode(home),
	}
}

// Detected returns agents whose ConfigDir exists on disk.
func Detected() []Agent {
	var result []Agent
	for _, a := range All() {
		if fi, err := os.Stat(a.ConfigDir); err == nil && fi.IsDir() {
			result = append(result, a)
		}
	}
	return result
}

// Find returns the agent with the given ID, or an error if not found.
func Find(id string) (Agent, error) {
	for _, a := range All() {
		if a.ID == id {
			return a, nil
		}
	}
	return Agent{}, fmt.Errorf("unknown agent %q (valid: claude, codex, gemini, opencode)", id)
}

// FindMCPBinary locates the brain-mcp binary.
// It first checks next to the current executable, then falls back to PATH.
func FindMCPBinary() (string, error) {
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			exe = resolved
		}
		candidate := filepath.Join(filepath.Dir(exe), "brain-mcp")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	if path, err := exec.LookPath("brain-mcp"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("brain-mcp not found; install it with 'make install-mcp' or 'brew install brain'")
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	h, _ := os.UserHomeDir()
	return h
}
