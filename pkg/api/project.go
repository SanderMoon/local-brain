package api

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectInfo represents a project in the active directory
type ProjectInfo struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Focused   bool   `json:"focused"`
	RepoCount int    `json:"repo_count"`
	TaskCount int    `json:"task_count"`
}

// ListProjects returns all projects in the active directory
func ListProjects(activeDir, focusedProject string) ([]ProjectInfo, error) {
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read active directory: %w", err)
	}

	var projects []ProjectInfo

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		projectName := entry.Name()
		projectPath := filepath.Join(activeDir, projectName)

		// Count repos
		repoCount := 0
		reposFile := filepath.Join(projectPath, ".repos")
		if data, err := os.ReadFile(reposFile); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && !strings.HasPrefix(line, "#") {
					repoCount++
				}
			}
		}

		// Count tasks
		taskCount := 0
		todoFile := filepath.Join(projectPath, "todo.md")
		if data, err := os.ReadFile(todoFile); err == nil {
			taskPattern := regexp.MustCompile(`^\s*- \[ \]`)
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				if taskPattern.MatchString(line) {
					taskCount++
				}
			}
		}

		projects = append(projects, ProjectInfo{
			Name:      projectName,
			Path:      projectPath,
			Focused:   projectName == focusedProject,
			RepoCount: repoCount,
			TaskCount: taskCount,
		})
	}

	return projects, nil
}

// ExtractRepoName extracts the repository name from a git URL
func ExtractRepoName(gitURL string) string {
	// Remove trailing slash and .git
	url := strings.TrimSuffix(strings.TrimSuffix(gitURL, "/"), ".git")

	// Extract from various URL formats
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`/([^/]+)\.git$`),
		regexp.MustCompile(`/([^/]+)$`),
		regexp.MustCompile(`:([^/]+)\.git$`),
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(gitURL); len(matches) > 1 {
			return matches[1]
		}
	}

	return filepath.Base(url)
}

// GetLinkedRepos returns the list of linked repository paths for a project
func GetLinkedRepos(projectDir string) ([]string, error) {
	reposFile := filepath.Join(projectDir, ".repos")

	if _, err := os.Stat(reposFile); os.IsNotExist(err) {
		return []string{}, nil
	}

	file, err := os.Open(reposFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open .repos file: %w", err)
	}
	defer file.Close()

	devDir := filepath.Join(os.Getenv("HOME"), "dev")
	var repos []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		gitURL := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if gitURL == "" || strings.HasPrefix(gitURL, "#") {
			continue
		}

		// Extract repo name and construct path
		repoName := ExtractRepoName(gitURL)
		if repoName != "" {
			repos = append(repos, filepath.Join(devDir, repoName))
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read .repos file: %w", err)
	}

	return repos, nil
}

// AddRepoLink adds a git URL to the project's .repos file
func AddRepoLink(projectDir, gitURL string) error {
	reposFile := filepath.Join(projectDir, ".repos")

	// Check if already linked
	if data, err := os.ReadFile(reposFile); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) == gitURL {
				// Already linked
				return nil
			}
		}
	}

	// Append to file
	f, err := os.OpenFile(reposFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open .repos file: %w", err)
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, gitURL)
	return err
}
