package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

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
	return filepath.Join(brainPath, "02_archive"), nil
}

// GetLinkedRepos returns the list of linked repository paths for a project
func GetLinkedRepos(cfg *Config, projectName string) ([]string, error) {
	projectPath, err := GetProjectPath(cfg, projectName)
	if err != nil {
		return nil, err
	}

	reposFile := filepath.Join(projectPath, ".repos")
	if _, err := os.Stat(reposFile); os.IsNotExist(err) {
		return []string{}, nil
	}

	devDir := filepath.Join(os.Getenv("HOME"), "dev")
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
