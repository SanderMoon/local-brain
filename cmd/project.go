package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var projectJSONFlag bool

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Project management",
	Long: `Manage projects within the current brain.

Projects are stored in the 01_active directory and can be linked
to git repositories, contain tasks, and have notes.`,
}

var projectListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List active projects",
	RunE:    runProjectList,
}

var projectNewCmd = &cobra.Command{
	Use:   "new <name>",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectNew,
}

var projectSelectCmd = &cobra.Command{
	Use:   "select [name]",
	Short: "Focus on a project",
	Long: `Set the focused project for subsequent commands.

If no name is provided, shows interactive selection.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProjectSelect,
}

var projectCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show currently focused project",
	RunE:  runProjectCurrent,
}

var projectCloneCmd = &cobra.Command{
	Use:   "clone <url> [name]",
	Short: "Import a git repository as a project",
	Long: `Clone a git repository and set it up as a new project.

Creates the project, links the repository, and pulls the code.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runProjectClone,
}

var projectLinkCmd = &cobra.Command{
	Use:   "link <git-url>",
	Short: "Link a git repository to current/focused project",
	Args:  cobra.ExactArgs(1),
	RunE:  runProjectLink,
}

var projectPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Clone/update linked repositories",
	RunE:  runProjectPull,
}

var projectArchiveCmd = &cobra.Command{
	Use:   "archive <name>",
	Short: "Archive a project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProjectArchive,
}

var projectMoveCmd = &cobra.Command{
	Use:   "move <project> <target-brain>",
	Short: "Move project to another brain",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runProjectMove,
}

var projectDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Permanently delete a project",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runProjectDelete,
}

func init() {
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(projectListCmd)
	projectCmd.AddCommand(projectNewCmd)
	projectCmd.AddCommand(projectSelectCmd)
	projectCmd.AddCommand(projectCurrentCmd)
	projectCmd.AddCommand(projectCloneCmd)
	projectCmd.AddCommand(projectLinkCmd)
	projectCmd.AddCommand(projectPullCmd)
	projectCmd.AddCommand(projectArchiveCmd)
	projectCmd.AddCommand(projectMoveCmd)
	projectCmd.AddCommand(projectDeleteCmd)

	projectListCmd.Flags().BoolVar(&projectJSONFlag, "json", false, "Output JSON format")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")
	focusedProject := cfg.GetFocusedProject()

	projects, err := api.ListProjects(activeDir, focusedProject)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if projectJSONFlag {
		data, err := json.MarshalIndent(projects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Println("Active Projects:")
	fmt.Println("----------------")

	if len(projects) == 0 {
		fmt.Println("(No active projects)")
		return nil
	}

	// Sort alphabetically
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	for _, proj := range projects {
		marker := " "
		status := ""
		if proj.Focused {
			marker = "*"
			status = "(selected)"
		}

		fmt.Printf(" %s %-20s %s [Repos: %d, Tasks: %d]\n",
			marker, proj.Name, status, proj.RepoCount, proj.TaskCount)
	}

	fmt.Println("")
	return nil
}

func runProjectNew(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(projectName) {
		return fmt.Errorf("project name can only contain letters, numbers, hyphens, and underscores")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")
	projectDir := filepath.Join(activeDir, projectName)

	if fileutil.FileExists(projectDir) {
		return fmt.Errorf("project '%s' already exists", projectName)
	}

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create notes.md
	notesContent := fmt.Sprintf(`# %s

Created: %s

## Overview

[Description]

## Notes
`, projectName, time.Now().Format("2006-01-02"))

	if err := os.WriteFile(filepath.Join(projectDir, "notes.md"), []byte(notesContent), 0644); err != nil {
		return fmt.Errorf("failed to create notes.md: %w", err)
	}

	// Create todo.md
	todoContent := `# Tasks

## Active

- [ ] Define project goals
- [ ] Set up development environment

## Completed
`

	if err := os.WriteFile(filepath.Join(projectDir, "todo.md"), []byte(todoContent), 0644); err != nil {
		return fmt.Errorf("failed to create todo.md: %w", err)
	}

	// Create empty .repos file
	if err := os.WriteFile(filepath.Join(projectDir, ".repos"), []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create .repos: %w", err)
	}

	fmt.Printf("OK: Created project: %s\n", projectName)

	// Auto-select the new project
	if err := cfg.SetFocusedProject(projectName); err != nil {
		return fmt.Errorf("failed to set focused project: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("OK: Selected project: %s\n", projectName)
	return nil
}

func runProjectSelect(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")

	var projectName string

	if len(args) == 0 {
		// Interactive selection
		if !external.IsFZFAvailable() {
			return fmt.Errorf("fzf not found (required for interactive mode)")
		}

		entries, err := os.ReadDir(activeDir)
		if err != nil {
			return fmt.Errorf("failed to read active directory: %w", err)
		}

		var projects []string
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				projects = append(projects, entry.Name())
			}
		}

		if len(projects) == 0 {
			return fmt.Errorf("no projects found")
		}

		selected, err := external.SelectOne(projects, external.FZFOptions{
			Header: "Select project to focus",
			Prompt: "Project> ",
		})

		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}

		projectName = selected
	} else {
		projectName = args[0]
	}

	// Verify project exists
	projectDir := filepath.Join(activeDir, projectName)
	if !fileutil.FileExists(projectDir) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Set focused project
	if err := cfg.SetFocusedProject(projectName); err != nil {
		return fmt.Errorf("failed to set focused project: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("OK: Selected project: %s\n", projectName)
	return nil
}

func runProjectCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	focused := cfg.GetFocusedProject()
	if focused == "" {
		return fmt.Errorf("no project selected")
	}

	fmt.Println(focused)
	return nil
}

func runProjectClone(cmd *cobra.Command, args []string) error {
	gitURL := args[0]
	var projectName string

	if len(args) > 1 {
		projectName = args[1]
	} else {
		// Extract from URL
		projectName = api.ExtractRepoName(gitURL)
		if projectName == "" {
			return fmt.Errorf("could not determine project name from URL")
		}
	}

	fmt.Printf("🚀 Importing '%s'...\n", projectName)

	// 1. Create new project
	if err := runProjectNew(cmd, []string{projectName}); err != nil {
		return err
	}

	// 2. Link repository
	if err := runProjectLink(cmd, []string{gitURL}); err != nil {
		return err
	}

	// 3. Pull repositories
	if err := runProjectPull(cmd, nil); err != nil {
		return err
	}

	fmt.Println("")
	fmt.Println("✨ Import complete!")
	fmt.Printf("Current project focused: %s\n", projectName)
	fmt.Println("Ready to go? Type: brain go")

	return nil
}

func runProjectLink(cmd *cobra.Command, args []string) error {
	gitURL := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target project (focused or interactive)
	projectName, projectDir, err := resolveTargetProject(cfg, "link repository")
	if err != nil {
		return err
	}

	// Verify remote (optional, with warning)
	fmt.Println("Verifying repository...")
	if err := external.VerifyRemote(gitURL); err != nil {
		fmt.Println("Warning: Could not verify repository.")
		// Ask for confirmation (for now, we'll proceed)
	}

	// Add to .repos file
	if err := api.AddRepoLink(projectDir, gitURL); err != nil {
		return fmt.Errorf("failed to link repository: %w", err)
	}

	fmt.Printf("OK: Linked to %s: %s\n", projectName, gitURL)
	fmt.Println("Run 'brain project pull' to clone.")

	return nil
}

func runProjectPull(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target project
	projectName, projectDir, err := resolveTargetProject(cfg, "pull repositories")
	if err != nil {
		return err
	}

	fmt.Printf("Project: %s\n", projectName)

	devDir := filepath.Join(os.Getenv("HOME"), "dev")
	if err := fileutil.EnsureDir(devDir); err != nil {
		return fmt.Errorf("failed to create dev directory: %w", err)
	}

	repos, err := api.GetLinkedRepos(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get linked repos: %w", err)
	}

	if len(repos) == 0 {
		fmt.Println("No repositories linked.")
		return nil
	}

	// Read .repos file for URLs
	reposFile := filepath.Join(projectDir, ".repos")
	data, err := os.ReadFile(reposFile)
	if err != nil {
		return fmt.Errorf("failed to read .repos: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, gitURL := range lines {
		gitURL = strings.TrimSpace(gitURL)
		if gitURL == "" || strings.HasPrefix(gitURL, "#") {
			continue
		}

		repoName := api.ExtractRepoName(gitURL)
		if repoName == "" {
			continue
		}

		repoPath := filepath.Join(devDir, repoName)
		fmt.Printf("  %s\n", repoName)

		if fileutil.FileExists(repoPath) {
			fmt.Println("    Updating...")
			if err := external.Pull(repoPath); err != nil {
				fmt.Printf("    ERROR: Failed update: %v\n", err)
			}
		} else {
			fmt.Println("    Cloning...")
			if err := external.Clone(gitURL, repoPath); err != nil {
				fmt.Printf("    ERROR: Failed clone: %v\n", err)
			}
		}

		fmt.Println("")
	}

	return nil
}

func runProjectArchive(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		// Use focused project
		projectName = cfg.GetFocusedProject()
		if projectName == "" {
			return fmt.Errorf("no project specified and no focused project")
		}
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")
	archiveDir := filepath.Join(brainPath, "99_archive")
	projectDir := filepath.Join(activeDir, projectName)

	if !fileutil.FileExists(projectDir) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Clear focus if archiving focused project
	if projectName == cfg.GetFocusedProject() {
		if err := cfg.SetFocusedProject(""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	// Create archive directory
	if err := fileutil.EnsureDir(archiveDir); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Create archive name with timestamp
	timestamp := time.Now().Format("20060102")
	archiveName := fmt.Sprintf("%s_%s", projectName, timestamp)
	archivePath := filepath.Join(archiveDir, archiveName)

	// Move project
	if err := os.Rename(projectDir, archivePath); err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	fmt.Printf("OK: Archived: %s\n", projectName)
	return nil
}

func runProjectMove(cmd *cobra.Command, args []string) error {
	projectName := args[0]
	var targetBrain string

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if len(args) > 1 {
		targetBrain = args[1]
	} else {
		// Interactive brain selection
		if !external.IsFZFAvailable() {
			return fmt.Errorf("target brain name required (fzf not available for interactive selection)")
		}

		brains := cfg.ListBrains()
		currentBrain := cfg.GetCurrentBrain()

		// Filter out current brain
		var otherBrains []string
		for _, brain := range brains {
			if brain != currentBrain {
				otherBrains = append(otherBrains, brain)
			}
		}

		if len(otherBrains) == 0 {
			return fmt.Errorf("no other brains available")
		}

		selected, err := external.SelectOne(otherBrains, external.FZFOptions{
			Header: "Select target brain",
			Prompt: "Brain> ",
		})

		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}

		targetBrain = selected
	}

	// Validate target brain exists
	if !cfg.BrainExists(targetBrain) {
		return fmt.Errorf("target brain '%s' does not exist", targetBrain)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	targetBrainPath, err := cfg.GetBrainPath(targetBrain)
	if err != nil {
		return fmt.Errorf("failed to get target brain path: %w", err)
	}

	currentPath := filepath.Join(brainPath, "01_active", projectName)
	targetPath := filepath.Join(targetBrainPath, "01_active", projectName)

	if !fileutil.FileExists(currentPath) {
		return fmt.Errorf("project '%s' not found in current brain", projectName)
	}

	if fileutil.FileExists(targetPath) {
		return fmt.Errorf("project '%s' already exists in '%s'", projectName, targetBrain)
	}

	// Move
	fmt.Printf("Moving '%s' to '%s'...\n", projectName, targetBrain)
	if err := os.Rename(currentPath, targetPath); err != nil {
		return fmt.Errorf("failed to move project: %w", err)
	}

	// Clear focus if moving focused project
	if projectName == cfg.GetFocusedProject() {
		if err := cfg.SetFocusedProject(""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	fmt.Println("OK: Project moved successfully")
	return nil
}

func runProjectDelete(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var projectName string
	if len(args) > 0 {
		projectName = args[0]
	} else {
		// Use focused project
		projectName = cfg.GetFocusedProject()
		if projectName == "" {
			return fmt.Errorf("no project specified and no focused project")
		}
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	projectDir := filepath.Join(brainPath, "01_active", projectName)

	if !fileutil.FileExists(projectDir) {
		return fmt.Errorf("project '%s' not found", projectName)
	}

	// Warning
	fmt.Println("WARNING: WARNING: You are about to PERMANENTLY DELETE project '" + projectName + "'")
	fmt.Printf("  Location: %s\n", projectDir)
	fmt.Println("  This action cannot be undone.")
	fmt.Println("  Consider using 'brain project archive' instead.")
	fmt.Println("")
	fmt.Print("Type the project name to confirm: ")

	var confirmation string
	fmt.Scanln(&confirmation)

	if confirmation != projectName {
		fmt.Println("Aborted")
		return nil
	}

	// Delete
	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Clear focus if deleting focused project
	if projectName == cfg.GetFocusedProject() {
		if err := cfg.SetFocusedProject(""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	fmt.Printf("OK: Deleted project: %s\n", projectName)
	return nil
}

// Helper function to resolve target project (focused or PWD or interactive)
func resolveTargetProject(cfg *config.Config, actionDesc string) (string, string, error) {
	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", "", fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")

	// Check PWD first
	cwd, _ := os.Getwd()
	if strings.HasPrefix(cwd, activeDir+string(filepath.Separator)) {
		projectName := filepath.Base(cwd)
		return projectName, cwd, nil
	}

	// Check focused project
	focused := cfg.GetFocusedProject()
	if focused != "" {
		projectDir := filepath.Join(activeDir, focused)
		if fileutil.FileExists(projectDir) {
			return focused, projectDir, nil
		}
	}

	// Interactive selection
	if !external.IsFZFAvailable() {
		return "", "", fmt.Errorf("cannot resolve project. Install fzf for interactive selection or use 'brain project select <name>'")
	}

	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read active directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, entry.Name())
		}
	}

	if len(projects) == 0 {
		return "", "", fmt.Errorf("no projects found")
	}

	selected, err := external.SelectOne(projects, external.FZFOptions{
		Header: "Select project to " + actionDesc,
		Prompt: "Project> ",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return "", "", fmt.Errorf("no project selected")
		}
		return "", "", err
	}

	projectDir := filepath.Join(activeDir, selected)
	return selected, projectDir, nil
}
