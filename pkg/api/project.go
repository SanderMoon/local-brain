package api

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/fileutil"
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

// ReadProjectDescription reads the description.md file for a project
// Returns empty string if file doesn't exist (not an error)
func ReadProjectDescription(projectDir string) (string, error) {
	descPath := filepath.Join(projectDir, "description.md")

	data, err := os.ReadFile(descPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read description.md: %w", err)
	}

	return string(data), nil
}

// WriteProjectDescription writes content to the description.md file
// Creates the file if it doesn't exist
func WriteProjectDescription(projectDir, content string) error {
	descPath := filepath.Join(projectDir, "description.md")

	// Ensure single trailing newline
	content = strings.TrimRight(content, "\n") + "\n"

	if err := os.WriteFile(descPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write description.md: %w", err)
	}

	return nil
}

// ProjectDescriptionExists checks if a project has a description.md file
func ProjectDescriptionExists(projectDir string) bool {
	descPath := filepath.Join(projectDir, "description.md")
	_, err := os.Stat(descPath)
	return err == nil
}

// ValidateProjectName validates a project name against naming rules
// Rules:
// - Must start with alphanumeric character
// - Can contain letters, numbers, hyphens, and underscores
// - Length must be 1-64 characters
// - Cannot start with a dot
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 64 {
		return fmt.Errorf("project name too long (max 64 characters)")
	}

	// Must start with alphanumeric
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	if !validPattern.MatchString(name) {
		return fmt.Errorf("project name must start with a letter or number and contain only letters, numbers, hyphens, and underscores")
	}

	return nil
}

// CreateProject creates a new project with directory structure and initial files
// Returns the full path to the created project directory
func CreateProject(activeDir, projectName string) (string, error) {
	// Validate project name
	if err := ValidateProjectName(projectName); err != nil {
		return "", err
	}

	projectPath := filepath.Join(activeDir, projectName)

	// Check if project already exists
	if fileutil.FileExists(projectPath) {
		return "", fmt.Errorf("project '%s' already exists", projectName)
	}

	// Create project directory
	if err := fileutil.EnsureDir(projectPath); err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create todo.md
	todoContent := `# Tasks

## Active

## Completed
`
	todoPath := filepath.Join(projectPath, "todo.md")
	if err := os.WriteFile(todoPath, []byte(todoContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create todo.md: %w", err)
	}

	// Create notes.md
	notesContent := fmt.Sprintf(`# %s Notes

`, projectName)
	notesPath := filepath.Join(projectPath, "notes.md")
	if err := os.WriteFile(notesPath, []byte(notesContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create notes.md: %w", err)
	}

	// Create empty notes/ directory
	notesDir := filepath.Join(projectPath, "notes")
	if err := fileutil.EnsureDir(notesDir); err != nil {
		return "", fmt.Errorf("failed to create notes directory: %w", err)
	}

	// Create empty .repos file
	reposPath := filepath.Join(projectPath, ".repos")
	if err := os.WriteFile(reposPath, []byte(""), 0644); err != nil {
		return "", fmt.Errorf("failed to create .repos file: %w", err)
	}

	return projectPath, nil
}

// ArchiveProject moves a project from active to archive directory with timestamp
// Format: {brainPath}/02_archive/{projectName}-YYYYMMDD/
func ArchiveProject(brainPath, projectName string) error {
	srcPath := filepath.Join(brainPath, "01_active", projectName)
	archiveDir := filepath.Join(brainPath, "02_archive")

	// Check if project exists
	if !fileutil.FileExists(srcPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Ensure archive directory exists
	if err := fileutil.EnsureDir(archiveDir); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Create archive name with timestamp
	timestamp := time.Now().Format("20060102")
	archiveName := fmt.Sprintf("%s-%s", projectName, timestamp)
	dstPath := filepath.Join(archiveDir, archiveName)

	// Handle duplicate archive names (append counter)
	counter := 1
	originalDstPath := dstPath
	for fileutil.FileExists(dstPath) {
		dstPath = fmt.Sprintf("%s-%d", originalDstPath, counter)
		counter++
	}

	// Move project atomically
	if err := os.Rename(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	return nil
}

// DeleteProject permanently removes a project directory
func DeleteProject(activeDir, projectName string) error {
	projectPath := filepath.Join(activeDir, projectName)

	// Check if project exists
	if !fileutil.FileExists(projectPath) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Remove project directory
	if err := os.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// MoveProjectToBrain moves a project from one brain to another
// Copies the entire project directory and then removes the source
func MoveProjectToBrain(srcBrainPath, dstBrainPath, projectName string) error {
	srcPath := filepath.Join(srcBrainPath, "01_active", projectName)
	dstActiveDir := filepath.Join(dstBrainPath, "01_active")
	dstPath := filepath.Join(dstActiveDir, projectName)

	// Check if source project exists
	if !fileutil.FileExists(srcPath) {
		return fmt.Errorf("project '%s' not found in source brain", projectName)
	}

	// Check if destination brain exists
	if !fileutil.FileExists(dstBrainPath) {
		return fmt.Errorf("destination brain path does not exist: %s", dstBrainPath)
	}

	// Ensure destination active directory exists
	if err := fileutil.EnsureDir(dstActiveDir); err != nil {
		return fmt.Errorf("failed to create destination active directory: %w", err)
	}

	// Check if project already exists in destination
	if fileutil.FileExists(dstPath) {
		return fmt.Errorf("project '%s' already exists in destination brain", projectName)
	}

	// Copy project directory recursively
	if err := copyDir(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to copy project: %w", err)
	}

	// Remove source project after successful copy
	if err := os.RemoveAll(srcPath); err != nil {
		return fmt.Errorf("failed to remove source project: %w", err)
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each entry
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Get source file info for permissions
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

// ProjectContext provides complete project details in a single call
// Optimized for MCP server to minimize round-trips
type ProjectContext struct {
	Name        string     `json:"name"`
	Path        string     `json:"path"`
	Description string     `json:"description"`
	Focused     bool       `json:"focused"`
	Todos       []TodoItem `json:"todos"`
	LinkedRepos []string   `json:"linked_repos"`
	NoteFiles   []NoteFile `json:"note_files"`
}

// GetProjectContext returns complete project context including todos, notes, repos, and description
// projectPath: absolute path to the project directory
// projectName: name of the project
// focusedProject: name of the currently focused project (empty string if none)
// includeCompleted: whether to include completed todos
func GetProjectContext(projectPath, projectName, focusedProject string, includeCompleted bool) (*ProjectContext, error) {
	// Check if project exists
	if !fileutil.FileExists(projectPath) {
		return nil, fmt.Errorf("project not found: %s", projectName)
	}

	// Get description
	description, err := ReadProjectDescription(projectPath)
	if err != nil {
		// Non-fatal, continue with empty description
		description = ""
	}

	// Get todos for this project
	todoFile := filepath.Join(projectPath, "todo.md")
	var todos []TodoItem
	if fileutil.FileExists(todoFile) {
		// Parse just this project's todos (not all projects)
		projectTodos, err := parseTodoFile(todoFile, projectName, includeCompleted)
		if err == nil {
			todos = projectTodos
		}
	}

	// Get linked repos
	linkedRepos, err := GetLinkedRepos(projectPath)
	if err != nil {
		// Non-fatal, continue with empty list
		linkedRepos = []string{}
	}

	// Get note files
	noteFiles, err := ListNotes(projectPath)
	if err != nil {
		// Non-fatal, continue with empty list
		noteFiles = []NoteFile{}
	}

	return &ProjectContext{
		Name:        projectName,
		Path:        projectPath,
		Description: description,
		Focused:     projectName == focusedProject,
		Todos:       todos,
		LinkedRepos: linkedRepos,
		NoteFiles:   noteFiles,
	}, nil
}
