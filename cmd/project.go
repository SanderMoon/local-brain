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
var projectSectionFlag string

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

var projectLinkPickFlag bool

var projectLinkCmd = &cobra.Command{
	Use:   "link [git-url|.|owner/repo]",
	Short: "Link a git repository to current/focused project",
	Long: `Link a git repository to the current or focused project.

Supports multiple input formats:
  brain project link <full-git-url>     Link by full URL
  brain project link .                  Detect remote URL from current directory
  brain project link /path/to/repo      Detect remote URL from local path
  brain project link owner/repo         Expand GitHub shorthand to full URL
  brain project link --pick             Scan ~/dev for repos and pick with fzf`,
	Args: cobra.MaximumNArgs(1),
	RunE: runProjectLink,
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

var projectDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Edit or show project description",
	Long: `Edit or display the project description.

By default, opens description.md in your editor.
Use --show to display the current description.`,
	RunE: runProjectDescribe,
}

var projectDescribeShowFlag bool

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
	projectCmd.AddCommand(projectDescribeCmd)

	projectListCmd.Flags().BoolVar(&projectJSONFlag, "json", false, "Output JSON format")
	projectListCmd.Flags().StringVar(&projectSectionFlag, "section", "01_active", "PARA section (01_active, 02_areas, 03_resources)")
	projectNewCmd.Flags().StringVar(&projectSectionFlag, "section", "01_active", "PARA section (01_active, 02_areas, 03_resources)")
	projectLinkCmd.Flags().BoolVar(&projectLinkPickFlag, "pick", false, "Scan ~/dev for git repos and pick with fzf")
	projectDescribeCmd.Flags().BoolVar(&projectDescribeShowFlag, "show", false, "Display description instead of editing")
	projectDescribeCmd.Flags().BoolVar(&projectJSONFlag, "json", false, "Output JSON format (with --show)")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	activeDir, err := config.GetSectionPath(cfg, projectSectionFlag)
	if err != nil {
		return fmt.Errorf("failed to get section path: %w", err)
	}

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

	activeDir, err := config.GetSectionPath(cfg, projectSectionFlag)
	if err != nil {
		return fmt.Errorf("failed to get section path: %w", err)
	}

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
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var gitURL string

	if projectLinkPickFlag {
		// --pick mode: scan dev directory for git repos and pick with fzf
		if !external.IsFZFAvailable() {
			return fmt.Errorf("fzf not found (required for --pick mode)")
		}

		devDir := cfg.GetDevDir()

		entries, err := os.ReadDir(devDir)
		if err != nil {
			return fmt.Errorf("failed to read dev directory (%s): %w", devDir, err)
		}

		var repos []string
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			repoPath := filepath.Join(devDir, entry.Name())
			if external.IsGitRepo(repoPath) {
				repos = append(repos, entry.Name())
			}
		}

		if len(repos) == 0 {
			return fmt.Errorf("no git repositories found in %s", devDir)
		}

		selected, err := external.SelectOne(repos, external.FZFOptions{
			Header: "Select repository to link",
			Prompt: "Repo> ",
		})
		if err != nil {
			if err.Error() == "cancelled" {
				return nil
			}
			return err
		}

		repoPath := filepath.Join(devDir, selected)
		remoteURL, err := external.GetRemoteURL(repoPath)
		if err != nil {
			return fmt.Errorf("selected repo '%s' has no remote origin: %w", selected, err)
		}

		gitURL = remoteURL
		fmt.Printf("Detected URL: %s\n", gitURL)

	} else if len(args) == 0 {
		return fmt.Errorf("requires a git URL, path, or owner/repo argument (or use --pick)")
	} else {
		arg := args[0]

		// Determine mode based on argument
		githubShorthand := regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)

		if arg == "." || strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "./") {
			// Local directory mode: detect git remote URL
			dirPath := arg
			if arg == "." || strings.HasPrefix(arg, "./") {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("failed to get working directory: %w", err)
				}
				if arg == "." {
					dirPath = cwd
				} else {
					dirPath = filepath.Join(cwd, arg[2:])
				}
			}

			if !external.IsGitRepo(dirPath) {
				return fmt.Errorf("'%s' is not a git repository", arg)
			}

			remoteURL, err := external.GetRemoteURL(dirPath)
			if err != nil {
				return fmt.Errorf("'%s' has no remote origin configured: %w", arg, err)
			}

			gitURL = remoteURL
			fmt.Printf("Detected URL: %s\n", gitURL)

		} else if githubShorthand.MatchString(arg) {
			// GitHub shorthand mode: expand owner/repo
			gitURL = fmt.Sprintf("https://github.com/%s.git", arg)
			fmt.Printf("Expanded URL: %s\n", gitURL)

		} else {
			// Full URL mode (existing behavior)
			gitURL = arg
		}
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

	devDir := cfg.GetDevDir()
	if err := fileutil.EnsureDir(devDir); err != nil {
		return fmt.Errorf("failed to create dev directory: %w", err)
	}

	repos, err := api.GetLinkedRepos(projectDir, devDir)
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

	// Clear focus if archiving focused project
	if projectName == cfg.GetFocusedProject() {
		if err := cfg.SetFocusedProject(""); err != nil {
			return fmt.Errorf("failed to clear focus: %w", err)
		}
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	if err := api.ArchiveProject(brainPath, projectName); err != nil {
		return err
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
	_, _ = fmt.Scanln(&confirmation)

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

func runProjectDescribe(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Resolve target project (focused or interactive)
	projectName, projectDir, err := resolveTargetProject(cfg, "describe project")
	if err != nil {
		return err
	}

	descPath := filepath.Join(projectDir, "description.md")

	// Display mode
	if projectDescribeShowFlag {
		content, err := api.ReadProjectDescription(projectDir)
		if err != nil {
			return err
		}

		if projectJSONFlag {
			type DescriptionOutput struct {
				Project     string `json:"project"`
				Description string `json:"description"`
				HasContent  bool   `json:"has_content"`
			}

			output := DescriptionOutput{
				Project:     projectName,
				Description: content,
				HasContent:  content != "",
			}

			data, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		// Human-readable output
		if content == "" {
			fmt.Println("(No description)")
			return nil
		}

		fmt.Print(content)
		return nil
	}

	// Interactive editing mode
	// Create file with helpful comment if it doesn't exist
	if !fileutil.FileExists(descPath) {
		initialContent := `# Add a brief description of this project
# Lines starting with # are comments and can be removed

`
		if err := os.WriteFile(descPath, []byte(initialContent), 0644); err != nil {
			return fmt.Errorf("failed to create description.md: %w", err)
		}
	}

	// Open in editor
	if err := external.OpenFile(descPath); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	fmt.Printf("OK: Updated description for: %s\n", projectName)
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
