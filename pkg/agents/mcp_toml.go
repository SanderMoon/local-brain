package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TOMLMCPInstaller manages MCP server registration in TOML config files.
// Used by agents that store MCP config in TOML format (Codex).
//
// It uses simple string manipulation rather than a full TOML parser,
// which keeps dependencies minimal while handling the straightforward
// [section.name] format used for MCP server entries.
type TOMLMCPInstaller struct {
	ConfigPath  string // e.g. ~/.codex/config.toml
	SectionName string // e.g. "mcp_servers"
}

func (t *TOMLMCPInstaller) sectionHeader(serverName string) string {
	return fmt.Sprintf("[%s.%s]", t.SectionName, serverName)
}

func (t *TOMLMCPInstaller) Install(serverName, binaryPath string) error {
	content, err := t.readFile()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	header := t.sectionHeader(serverName)
	if strings.Contains(content, header) {
		// Update existing section: remove old, then re-add
		content = t.removeSection(content, serverName)
	}

	section := fmt.Sprintf("\n%s\ncommand = %q\n", header, binaryPath)
	content = strings.TrimRight(content, "\n") + "\n" + section

	return t.writeFile(content)
}

func (t *TOMLMCPInstaller) Remove(serverName string) error {
	content, err := t.readFile()
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if !strings.Contains(content, t.sectionHeader(serverName)) {
		return nil
	}

	content = t.removeSection(content, serverName)
	return t.writeFile(content)
}

func (t *TOMLMCPInstaller) IsInstalled(serverName string) (bool, error) {
	content, err := t.readFile()
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return strings.Contains(content, t.sectionHeader(serverName)), nil
}

// removeSection removes a TOML section and its key-value lines.
func (t *TOMLMCPInstaller) removeSection(content, serverName string) string {
	header := t.sectionHeader(serverName)
	idx := strings.Index(content, header)
	if idx < 0 {
		return content
	}

	// Find end of section: next section header or end of file
	rest := content[idx+len(header):]
	endIdx := strings.Index(rest, "\n[")
	if endIdx < 0 {
		// Last section in file
		return strings.TrimRight(content[:idx], "\n") + "\n"
	}
	return content[:idx] + rest[endIdx+1:]
}

func (t *TOMLMCPInstaller) readFile() (string, error) {
	data, err := os.ReadFile(t.ConfigPath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (t *TOMLMCPInstaller) writeFile(content string) error {
	if err := os.MkdirAll(filepath.Dir(t.ConfigPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(t.ConfigPath, []byte(content), 0644)
}
