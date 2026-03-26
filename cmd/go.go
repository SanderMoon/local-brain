package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var (
	goPrint bool
	goRepo  bool
)

var goCmd = &cobra.Command{
	Use:   "go [project]",
	Short: "Jump into a project or its linked repository",
	Long: `Open a new shell in a project directory or one of its linked repos.

Without arguments, shows a fuzzy picker for project selection.
With a project name, jumps directly to that project.

By default, opens the project directory. With --repo, selects from
linked repositories instead (auto-selects if only one repo is linked).

Use --print to output the path instead of launching a shell.
This is useful for composing with other tools:

  cd $(brain go myproject --print)
  code $(brain go myproject --repo --print)`,
	Example: `  brain go                          # pick a project interactively
  brain go website-redesign         # jump to project directory
  brain go website-redesign --repo  # jump to linked repo
  brain go --repo                   # pick project, then pick repo
  brain go myproject --print        # print path (for scripting)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGo,
}

func init() {
	rootCmd.AddCommand(goCmd)
	goCmd.Flags().BoolVarP(&goPrint, "print", "p", false, "Print path instead of launching shell")
	goCmd.Flags().BoolVarP(&goRepo, "repo", "r", false, "Jump to a linked repository instead of the project directory")
}

func runGo(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")

	var projectDir string

	if len(args) == 1 {
		// Direct project name given
		projectDir = filepath.Join(activeDir, args[0])
		if !fileutil.FileExists(projectDir) {
			return fmt.Errorf("project %q not found", args[0])
		}
	} else {
		// Interactive selection
		projectDir, err = selectProject(activeDir)
		if err != nil {
			return err
		}
	}

	// Determine target directory
	targetDir := projectDir

	if goRepo {
		targetDir, err = selectRepo(projectDir, cfg.GetDevDir())
		if err != nil {
			return err
		}
	}

	// --print mode: just output the path
	if goPrint {
		fmt.Println(targetDir)
		return nil
	}

	// Launch a subshell in the target directory
	return launchShell(targetDir)
}

func selectProject(activeDir string) (string, error) {
	if !external.IsFZFAvailable() {
		return "", fmt.Errorf("fzf is required for interactive selection (or pass a project name as argument)")
	}

	entries, err := os.ReadDir(activeDir)
	if err != nil {
		return "", fmt.Errorf("failed to read active directory: %w", err)
	}

	var projects []string
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			projects = append(projects, entry.Name())
		}
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("no projects found in %s", activeDir)
	}

	previewCmd := `
		DIR="` + activeDir + `/{}"
		REPOS="$DIR/.repos"
		TODO="$DIR/todo.md"
		if [ -f "$REPOS" ] && [ -s "$REPOS" ]; then
			echo "REPOS"
			grep -v '^#' "$REPOS" | grep -v '^$' | while read -r url; do echo "  $url"; done
			echo ""
		fi
		if [ -f "$TODO" ]; then
			echo "TODO"
			if command -v bat &>/dev/null; then
				bat --color=always --style=plain "$TODO" 2>/dev/null
			else
				cat "$TODO"
			fi
		fi
	`

	selected, err := external.SelectOne(projects, external.FZFOptions{
		Header:        "Select a project",
		Prompt:        "Project> ",
		Preview:       previewCmd,
		PreviewWindow: "right:50%",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return "", fmt.Errorf("cancelled")
		}
		return "", err
	}

	return filepath.Join(activeDir, selected), nil
}

func selectRepo(projectDir string, devDir string) (string, error) {
	repos, err := api.GetLinkedRepos(projectDir, devDir)
	if err != nil {
		return "", fmt.Errorf("failed to get linked repos: %w", err)
	}

	if len(repos) == 0 {
		return "", fmt.Errorf("no repos linked to %s (add URLs to .repos file)", filepath.Base(projectDir))
	}

	// Auto-select if only one repo
	if len(repos) == 1 {
		if !fileutil.FileExists(repos[0]) {
			return "", fmt.Errorf("repo directory not found: %s\nRun 'brain project pull' to clone it", repos[0])
		}
		return repos[0], nil
	}

	// Multiple repos: let user pick
	if !external.IsFZFAvailable() {
		return "", fmt.Errorf("fzf is required to select from multiple repos (project has %d linked repos)", len(repos))
	}

	// Show just repo names for cleaner display
	var displayNames []string
	repoMap := make(map[string]string)
	for _, repo := range repos {
		name := filepath.Base(repo)
		displayNames = append(displayNames, name)
		repoMap[name] = repo
	}

	selected, err := external.SelectOne(displayNames, external.FZFOptions{
		Header: fmt.Sprintf("Select a repo (%s)", filepath.Base(projectDir)),
		Prompt: "Repo> ",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return "", fmt.Errorf("cancelled")
		}
		return "", err
	}

	repoPath := repoMap[selected]
	if !fileutil.FileExists(repoPath) {
		return "", fmt.Errorf("repo directory not found: %s\nRun 'brain project pull' to clone it", repoPath)
	}

	return repoPath, nil
}

func launchShell(dir string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}

	shellBin, err := exec.LookPath(shell)
	if err != nil {
		return fmt.Errorf("shell not found: %w", err)
	}

	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", dir, err)
	}

	return syscall.Exec(shellBin, []string{shell}, os.Environ())
}
