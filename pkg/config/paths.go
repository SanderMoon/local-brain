package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
)

// GetBrainRoot returns the root directory for all brains
// Can be overridden with BRAIN_ROOT environment variable
// In dev mode (brain-dev binary), uses ~/brains-dev by default
func GetBrainRoot() string {
	if brainRoot := os.Getenv("BRAIN_ROOT"); brainRoot != "" {
		expanded, err := fileutil.ExpandPath(brainRoot)
		if err == nil {
			return expanded
		}
	}

	// Use dev-specific path if running brain-dev
	suffix := ""
	if isDevMode() {
		suffix = "-dev"
	}

	return filepath.Join(os.Getenv("HOME"), "brains"+suffix)
}

// GetDumpPath returns the path to the dump file for the current brain
func GetDumpPath(cfg *Config) (string, error) {
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(brainPath, "00_dump.md"), nil
}

// GetProjectsPath returns the path to the active projects directory
func GetProjectsPath(cfg *Config) (string, error) {
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(brainPath, "01_active"), nil
}

// ValidSections lists all valid PARA section names
var ValidSections = []string{"01_active", "02_areas", "03_resources"}

// ValidateSection returns an error if section is not a valid PARA section
func ValidateSection(section string) error {
	for _, s := range ValidSections {
		if s == section {
			return nil
		}
	}
	return fmt.Errorf("invalid section %q: must be one of %v", section, ValidSections)
}

// GetSectionPath returns the path to a specific PARA section directory.
// If section is empty, defaults to "01_active".
func GetSectionPath(cfg *Config, section string) (string, error) {
	if section == "" {
		section = "01_active"
	}
	if err := ValidateSection(section); err != nil {
		return "", err
	}
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(brainPath, section), nil
}

// GetProjectPath returns the path to a specific project
func GetProjectPath(cfg *Config, projectName string) (string, error) {
	projectsPath, err := GetProjectsPath(cfg)
	if err != nil {
		return "", err
	}
	return filepath.Join(projectsPath, projectName), nil
}

// GetArchivePath returns the path to the archive directory
func GetArchivePath(cfg *Config) (string, error) {
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(brainPath, "99_archive"), nil
}

// GetLinkedRepos returns the list of linked repository paths for a project
func GetLinkedRepos(cfg *Config, projectName string, devDir string) ([]string, error) {
	projectPath, err := GetProjectPath(cfg, projectName)
	if err != nil {
		return nil, err
	}

	reposFile := filepath.Join(projectPath, ".repos")
	if _, err := os.Stat(reposFile); os.IsNotExist(err) {
		return []string{}, nil
	}

	var repos []string

	file, err := os.Open(reposFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open .repos file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		gitURL := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if gitURL == "" || strings.HasPrefix(gitURL, "#") {
			continue
		}

		// Extract repo name from URL
		repoName := extractRepoName(gitURL)
		if repoName != "" {
			repos = append(repos, filepath.Join(devDir, repoName))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read .repos file: %w", err)
	}

	return repos, nil
}

// extractRepoName extracts the repository name from a git URL
func extractRepoName(gitURL string) string {
	// Pattern: /repo-name.git or /repo-name at end of URL
	re1 := regexp.MustCompile(`/([^/]+)\.git$`)
	if matches := re1.FindStringSubmatch(gitURL); len(matches) > 1 {
		return matches[1]
	}

	re2 := regexp.MustCompile(`/([^/]+)$`)
	if matches := re2.FindStringSubmatch(gitURL); len(matches) > 1 {
		return matches[1]
	}

	// Pattern: :repo-name.git for SSH URLs
	re3 := regexp.MustCompile(`:([^/]+)\.git$`)
	if matches := re3.FindStringSubmatch(gitURL); len(matches) > 1 {
		return matches[1]
	}

	return ""
}
