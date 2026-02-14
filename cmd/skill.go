package cmd

import (
	"fmt"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/skillscatalog"
	"github.com/spf13/cobra"
)

var (
	skillAgentFlag string
	skillForceFlag bool
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Manage bundled AI agent skills",
	Long: `Install, remove, and inspect bundled AI agent skills.

Skills are SKILL.md files installed into AI coding agents (Claude Code, Codex,
Gemini CLI, OpenCode). They teach the agent context-aware workflows for your
Local Brain workspace.`,
}

var skillListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bundled skills",
	RunE:  runSkillList,
}

var skillInstallCmd = &cobra.Command{
	Use:   "install [name]",
	Short: "Install skill(s) to detected AI agents",
	Long: `Install bundled skills to AI coding agents.

Without a name argument, all bundled skills are installed.
Use --agent to target a specific agent; defaults to all detected agents.`,
	RunE: runSkillInstall,
}

var skillRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a skill from AI agents",
	Args:  cobra.ExactArgs(1),
	RunE:  runSkillRemove,
}

var skillStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installation status of bundled skills",
	RunE:  runSkillStatus,
}

func init() {
	rootCmd.AddCommand(skillCmd)
	skillCmd.AddCommand(skillListCmd)
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillRemoveCmd)
	skillCmd.AddCommand(skillStatusCmd)

	skillInstallCmd.Flags().StringVar(&skillAgentFlag, "agent", "all", "Target agent (claude, codex, gemini, opencode, all)")
	skillInstallCmd.Flags().BoolVar(&skillForceFlag, "force", false, "Overwrite existing skill files")
	skillRemoveCmd.Flags().StringVar(&skillAgentFlag, "agent", "all", "Target agent (claude, codex, gemini, opencode, all)")
}

func runSkillList(cmd *cobra.Command, args []string) error {
	skills, err := skillscatalog.ListSkills()
	if err != nil {
		return err
	}
	for _, s := range skills {
		fmt.Fprintf(cmd.OutOrStdout(), "%-20s %s\n", s.Name, s.Description)
	}
	return nil
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	skills, err := skillscatalog.ListSkills()
	if err != nil {
		return err
	}

	// Filter to requested skill if a name was provided
	if len(args) > 0 {
		name := args[0]
		content, err := skillscatalog.GetSkillContent(name)
		if err != nil {
			return err
		}
		_, desc := skillscatalog.ParseFrontmatter(content)
		skills = []skillscatalog.SkillInfo{{Name: name, Description: desc, Content: content}}
	}

	agents, err := resolveAgents(skillAgentFlag)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No AI agents detected. Install one of the following agents first:")
		for _, a := range skillscatalog.KnownAgents() {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-12s  config dir: %s\n", a.Name, a.ConfigDir)
		}
		return nil
	}

	for _, agent := range agents {
		for _, skill := range skills {
			installed, err := skillscatalog.InstallSkill(skill, agent, skillForceFlag)
			if err != nil {
				return fmt.Errorf("installing %s for %s: %w", skill.Name, agent.Name, err)
			}
			if installed {
				dest := agent.SkillsDir + "/" + skill.Name + "/"
				fmt.Fprintf(cmd.OutOrStdout(), "OK:   Installed %s → %s (%s)\n", skill.Name, agent.Name, dest)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "SKIP: %s already installed for %s (use --force to overwrite)\n", skill.Name, agent.Name)
			}
		}
	}
	return nil
}

func runSkillRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	agents, err := resolveAgents(skillAgentFlag)
	if err != nil {
		return err
	}
	if len(agents) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No AI agents detected.")
		return nil
	}

	for _, agent := range agents {
		removed, err := skillscatalog.RemoveSkill(name, agent)
		if err != nil {
			return fmt.Errorf("removing %s from %s: %w", name, agent.Name, err)
		}
		if removed {
			fmt.Fprintf(cmd.OutOrStdout(), "OK:   Removed %s from %s\n", name, agent.Name)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "SKIP: %s not installed for %s\n", name, agent.Name)
		}
	}
	return nil
}

func runSkillStatus(cmd *cobra.Command, args []string) error {
	skills, err := skillscatalog.ListSkills()
	if err != nil {
		return err
	}
	allAgents := skillscatalog.KnownAgents()
	detected := skillscatalog.DetectedAgents()
	detectedSet := make(map[string]bool)
	for _, a := range detected {
		detectedSet[a.ID] = true
	}

	// Print header
	header := fmt.Sprintf("%-20s", "Skill")
	for _, a := range allAgents {
		header += fmt.Sprintf("  %-14s", a.Name)
	}
	fmt.Fprintln(cmd.OutOrStdout(), header)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", len(header)))

	// Print rows
	for _, skill := range skills {
		row := fmt.Sprintf("%-20s", skill.Name)
		for _, agent := range allAgents {
			var cell string
			if !detectedSet[agent.ID] {
				cell = "-"
			} else if skillscatalog.IsInstalled(skill.Name, agent) {
				cell = "installed"
			} else {
				cell = "not installed"
			}
			row += fmt.Sprintf("  %-14s", cell)
		}
		fmt.Fprintln(cmd.OutOrStdout(), row)
	}
	return nil
}

// resolveAgents returns the list of agents to operate on based on the --agent flag.
// "all" returns all detected agents; a specific ID returns that agent (even if not detected).
func resolveAgents(agentFlag string) ([]skillscatalog.AgentInfo, error) {
	if agentFlag == "all" {
		return skillscatalog.DetectedAgents(), nil
	}
	a, err := skillscatalog.FindAgent(agentFlag)
	if err != nil {
		return nil, err
	}
	return []skillscatalog.AgentInfo{a}, nil
}
