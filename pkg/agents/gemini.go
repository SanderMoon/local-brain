package agents

import "path/filepath"

func newGemini(home string) Agent {
	configDir := filepath.Join(home, ".gemini")
	return Agent{
		Info: Info{
			ID:        "gemini",
			Name:      "Gemini CLI",
			ConfigDir: configDir,
			SkillsDir: filepath.Join(configDir, "skills"),
		},
		MCP: &JSONMCPInstaller{
			ConfigPath: filepath.Join(configDir, "settings.json"),
			ServerKey:  "mcpServers",
			BuildEntry: func(binaryPath string) interface{} {
				return map[string]interface{}{
					"command": binaryPath,
				}
			},
		},
	}
}
