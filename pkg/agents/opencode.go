package agents

import "path/filepath"

func newOpenCode(home string) Agent {
	configDir := filepath.Join(home, ".config", "opencode")
	return Agent{
		Info: Info{
			ID:        "opencode",
			Name:      "OpenCode",
			ConfigDir: configDir,
			SkillsDir: filepath.Join(configDir, "skills"),
		},
		MCP: &JSONMCPInstaller{
			ConfigPath: filepath.Join(configDir, "opencode.json"),
			ServerKey:  "mcp",
			BuildEntry: func(binaryPath string) interface{} {
				return map[string]interface{}{
					"type":    "local",
					"command": []string{binaryPath},
				}
			},
		},
	}
}
