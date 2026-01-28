package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var goCmd = &cobra.Command{
	Use:   "go",
	Short: "Jump to a project directory",
	Long: `Select and jump to a project with intelligent environment setup.

Features:
  - Fuzzy search through all active projects
  - Auto-detects linked repositories
  - Intelligent Environment Setup:
    - If Tmux is available and repos are linked:
      - Creates/Attaches to session 'brain-<project>'
      - Window 1 (Code): Repos dir, venv activated, editor opened
      - Window 2 (Notes): Project notes, editor opened
    - Fallback: Opens new shell in project directory`,
	RunE: runGo,
}

func init() {
	rootCmd.AddCommand(goCmd)
}

func runGo(cmd *cobra.Command, args []string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for project selection)")
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

	// Get list of projects
	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return fmt.Errorf("failed to read active directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, filepath.Join(activeDir, entry.Name()))
		}
	}

	if len(projects) == 0 {
		fmt.Printf("No projects found in %s\n", activeDir)
		return nil
	}

	// Select project with FZF
	previewCmd := `
		NOTES={}/notes.md
		TODO={}/todo.md
		echo " Notes "
		if [ -f "$NOTES" ]; then
			if command -v bat &>/dev/null; then
				bat --color=always --style=plain "$NOTES" 2>/dev/null
			else
				cat "$NOTES"
			fi
		else
			echo "No notes yet"
		fi
		echo ""
		echo " Tasks "
		if [ -f "$TODO" ]; then
			if command -v bat &>/dev/null; then
				bat --color=always --style=plain "$TODO" 2>/dev/null
			else
				cat "$TODO"
			fi
		else
			echo "No tasks yet"
		fi
	`

	selected, err := external.SelectOne(projects, external.FZFOptions{
		Header:        "Select a project to jump to (Esc to cancel)",
		Prompt:        "Project> ",
		Preview:       previewCmd,
		PreviewWindow: "right:60%",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			fmt.Println("No project selected")
			return nil
		}
		return err
	}

	projectDir := selected
	projectName := filepath.Base(projectDir)

	// Get linked repos
	repos, err := api.GetLinkedRepos(projectDir)
	if err != nil {
		return fmt.Errorf("failed to get linked repos: %w", err)
	}

	// Decide: Tmux or simple shell
	useTmux := external.IsTmuxAvailable() && len(repos) > 0

	if !useTmux {
		// Simple shell mode
		return runSimpleShell(projectDir, projectName, repos)
	}

	// Tmux dev environment mode
	return runTmuxEnvironment(projectDir, projectName, repos)
}

func runSimpleShell(projectDir, projectName string, repos []string) error {
	fmt.Printf("Entering project: %s\n", projectName)
	fmt.Printf("Location: %s\n", projectDir)

	if len(repos) > 0 {
		fmt.Println("Linked Repositories:")
		for _, repo := range repos {
			fmt.Printf("  - %s\n", repo)
		}
		fmt.Println("Tip: Install tmux to enable auto-environment setup")
	}

	fmt.Println("")
	fmt.Println("Type 'exit' or press Ctrl+D to return")

	// Change to project directory and exec shell
	if err := os.Chdir(projectDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}

	// TODO: Actually exec the shell (requires syscall.Exec)
	// shell := os.Getenv("SHELL")
	// if shell == "" {
	//     shell = "/bin/bash"
	// }

	return external.OpenFile(projectDir) // This will fail, but we want to exec shell
}

func runTmuxEnvironment(projectDir, projectName string, repos []string) error {
	var primaryRepo string

	if len(repos) == 1 {
		primaryRepo = repos[0]
	} else {
		// Multiple repos: ask user
		selected, err := external.SelectOne(repos, external.FZFOptions{
			Header: "Select primary repository for this session",
			Prompt: "Repo> ",
		})

		if err != nil {
			if err.Error() == "cancelled" {
				// Fall back to notes only
				if err := os.Chdir(projectDir); err != nil {
					return fmt.Errorf("failed to change directory: %w", err)
				}
				// TODO: exec shell
				// shell := os.Getenv("SHELL")
				// if shell == "" {
				//     shell = "/bin/bash"
				// }
				return nil
			}
			return err
		}

		primaryRepo = selected
	}

	// Check if repo exists
	if !fileutil.FileExists(primaryRepo) {
		fmt.Printf("Warning: Repository directory not found at %s\n", primaryRepo)
		fmt.Println("Run 'brain project pull' to clone it first.")
		fmt.Print("Press Enter to continue to notes only...")
		_, _ = fmt.Scanln()

		if err := os.Chdir(projectDir); err != nil {
			return fmt.Errorf("failed to change directory: %w", err)
		}
		// TODO: exec shell
		return nil
	}

	// Session name (replace dots with underscores)
	sessionName := "brain-" + strings.ReplaceAll(projectName, ".", "_")

	// Check if session already exists
	if external.HasSession(sessionName) {
		fmt.Printf("Attaching to existing session: %s\n", sessionName)
		return external.AttachSession(sessionName)
	}

	fmt.Printf("Setting up dev environment for %s...\n", projectName)

	// Create tmux session
	if err := external.CreateSession(sessionName, primaryRepo); err != nil {
		return fmt.Errorf("failed to create tmux session: %w", err)
	}

	// Rename first window to "code"
	_ = external.SendKeys(sessionName+":1", "tmux rename-window code") // Non-critical, ignore errors

	// Setup code window
	editorCmd := "vim"
	if _, err := external.DetectEditor(); err == nil {
		editorCmd = "nvim"
	}

	// Send commands to code window
	setupScript := fmt.Sprintf(`
if [ ! -d .venv ] && [ ! -d venv ]; then
  echo 'Creating virtual environment...'
  python3 -m venv .venv 2>/dev/null || true
fi

if [ -f .venv/bin/activate ]; then
  source .venv/bin/activate
elif [ -f venv/bin/activate ]; then
  source venv/bin/activate
fi

clear
%s .
`, editorCmd)

	_ = external.SendKeys(sessionName+":code", setupScript) // Non-critical, ignore errors

	// Create notes window
	_ = external.NewWindow(sessionName, 2, "notes", projectDir) // Non-critical, ignore errors

	// Setup notes window
	notesCmd := fmt.Sprintf("%s notes.md todo.md", editorCmd)
	_ = external.SendKeys(sessionName+":notes", notesCmd) // Non-critical, ignore errors

	// Select code window
	_ = external.SelectWindow(sessionName, 1) // Non-critical, ignore errors

	// Attach to session
	return external.AttachSession(sessionName)
}
