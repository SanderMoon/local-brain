package agents

import "path/filepath"

func newClaude(home string) Agent {
	configDir := filepath.Join(home, ".claude")
	return Agent{
		Info: Info{
			ID:        "claude",
			Name:      "Claude Code",
			ConfigDir: configDir,
			SkillsDir: filepath.Join(configDir, "skills"),
		},
		MCP: &JSONMCPInstaller{
			ConfigPath: filepath.Join(home, ".claude.json"),
			ServerKey:  "mcpServers",
			BuildEntry: func(binaryPath string) interface{} {
				return map[string]interface{}{
					"type":    "stdio",
					"command": binaryPath,
				}
			},
		},
	}
}
