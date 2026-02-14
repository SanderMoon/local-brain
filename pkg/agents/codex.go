package agents

import "path/filepath"

func newCodex(home string) Agent {
	configDir := filepath.Join(home, ".codex")
	return Agent{
		Info: Info{
			ID:        "codex",
			Name:      "Codex",
			ConfigDir: configDir,
			SkillsDir: filepath.Join(configDir, "skills"),
		},
		MCP: &TOMLMCPInstaller{
			ConfigPath:  filepath.Join(configDir, "config.toml"),
			SectionName: "mcp_servers",
		},
	}
}
