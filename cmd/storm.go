package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var stormAgent string

var stormCmd = &cobra.Command{
	Use:   "storm",
	Short: "Open an AI agent in your brains directory",
	Long: `Launch a CLI AI agent in the brains root directory.

Defaults to Claude Code (claude). Use --agent to specify a different tool.

Tip: Create a CLAUDE.md file in ~/brains to give Claude Code standing
instructions as your personal assistant. Other agents may use a different
mechanism for context (e.g., system prompts or project config files).`,
	Example: `  brain storm                    # Launch claude in ~/brains
  brain storm --agent aider       # Launch aider instead
  brain storm -a llm              # Launch llm instead`,
	RunE: runStorm,
}

func init() {
	rootCmd.AddCommand(stormCmd)
	stormCmd.Flags().StringVarP(&stormAgent, "agent", "a", "claude", "CLI agent to launch (e.g., claude, aider, llm)")
}

func runStorm(cmd *cobra.Command, args []string) error {
	brainRoot := config.GetBrainRoot()

	agentBin, err := exec.LookPath(stormAgent)
	if err != nil {
		return fmt.Errorf("agent %q not found in PATH: %w", stormAgent, err)
	}

	if err := os.Chdir(brainRoot); err != nil {
		return fmt.Errorf("failed to change to brains directory %s: %w", brainRoot, err)
	}

	return syscall.Exec(agentBin, []string{stormAgent}, os.Environ())
}
