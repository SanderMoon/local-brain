package skillscatalog

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// SkillInfo describes a bundled skill.
type SkillInfo struct {
	Name        string // directory name — canonical identifier
	Description string // parsed from YAML frontmatter
	Content     string // full raw SKILL.md content
}

// ParseFrontmatter extracts name and description from YAML frontmatter.
// It reads lines between the first and second "---" markers.
// Multi-line values are not supported.
func ParseFrontmatter(content string) (name, description string) {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	count := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			count++
			if count == 1 {
				inFrontmatter = true
				continue
			}
			// Second "---" closes frontmatter
			break
		}
		if !inFrontmatter {
			continue
		}
		if strings.HasPrefix(trimmed, "name:") {
			val := strings.TrimPrefix(trimmed, "name:")
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `"'`)
			name = val
		}
		if strings.HasPrefix(trimmed, "description:") {
			val := strings.TrimPrefix(trimmed, "description:")
			val = strings.TrimSpace(val)
			val = strings.Trim(val, `"'`)
			description = val
		}
	}
	return name, description
}

// ListSkills returns all bundled skills from the embedded filesystem.
func ListSkills() ([]SkillInfo, error) {
	entries, err := fs.ReadDir(SkillsFS, "skills")
	if err != nil {
		return nil, fmt.Errorf("reading skills directory: %w", err)
	}

	var skills []SkillInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		content, err := GetSkillContent(entry.Name())
		if err != nil {
			return nil, err
		}
		_, description := ParseFrontmatter(content)
		skills = append(skills, SkillInfo{
			Name:        entry.Name(),
			Description: description,
			Content:     content,
		})
	}
	return skills, nil
}

// GetSkillContent returns the raw SKILL.md content for the named skill.
func GetSkillContent(name string) (string, error) {
	path := filepath.Join("skills", name, "SKILL.md")
	data, err := SkillsFS.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("skill %q not found", name)
	}
	return string(data), nil
}

// IsInstalled reports whether the named skill is already installed for the agent.
func IsInstalled(name string, agent AgentInfo) bool {
	dest := filepath.Join(agent.SkillsDir, name, "SKILL.md")
	return fileutil.FileExists(dest)
}

// InstallSkill writes the skill's SKILL.md into the agent's skills directory.
// If the file already exists and force is false, it returns (false, nil) — skipped.
// Returns (true, nil) on successful install.
func InstallSkill(skill SkillInfo, agent AgentInfo, force bool) (installed bool, err error) {
	dest := filepath.Join(agent.SkillsDir, skill.Name, "SKILL.md")
	if fileutil.FileExists(dest) && !force {
		return false, nil
	}
	dir := filepath.Dir(dest)
	if err := fileutil.EnsureDir(dir); err != nil {
		return false, fmt.Errorf("creating skill directory %s: %w", dir, err)
	}
	if err := os.WriteFile(dest, []byte(skill.Content), 0644); err != nil {
		return false, fmt.Errorf("writing skill file %s: %w", dest, err)
	}
	return true, nil
}

// RemoveSkill removes the named skill from the agent's skills directory.
// Returns (false, nil) if the skill is not installed — not an error.
func RemoveSkill(name string, agent AgentInfo) (removed bool, err error) {
	dest := filepath.Join(agent.SkillsDir, name, "SKILL.md")
	if !fileutil.FileExists(dest) {
		return false, nil
	}
	if err := os.Remove(dest); err != nil {
		return false, fmt.Errorf("removing skill file %s: %w", dest, err)
	}
	// Best-effort: remove the now-empty skill directory
	_ = os.Remove(filepath.Dir(dest))
	return true, nil
}
