package cmd

import (
	"fmt"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/agents"
	"github.com/spf13/cobra"
)

var mcpAgentFlag string

var mcpManageCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Manage MCP server registration for AI agents",
	Long: `Register, remove, and check the brain-mcp server with AI coding agents.

This manages the MCP server configuration so agents can access your
Local Brain workspace through the MCP protocol.`,
}

var mcpInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Register brain-mcp with detected AI agents",
	Long: `Register the brain-mcp server with AI coding agents.

Detects which agents are installed and adds brain-mcp to their MCP config.
The server is registered as "local-brain" using stdio transport.`,
	RunE: runMCPInstall,
}

var mcpRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove brain-mcp from AI agents",
	RunE:  runMCPRemove,
}

var mcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show MCP registration status across agents",
	RunE:  runMCPStatus,
}

func init() {
	rootCmd.AddCommand(mcpManageCmd)
	mcpManageCmd.AddCommand(mcpInstallCmd)
	mcpManageCmd.AddCommand(mcpRemoveCmd)
	mcpManageCmd.AddCommand(mcpStatusCmd)

	mcpInstallCmd.Flags().StringVar(&mcpAgentFlag, "agent", "all", "Target agent (claude, codex, gemini, opencode, all)")
	mcpRemoveCmd.Flags().StringVar(&mcpAgentFlag, "agent", "all", "Target agent (claude, codex, gemini, opencode, all)")
}

const mcpServerName = "local-brain"

func runMCPInstall(cmd *cobra.Command, _ []string) error {
	binaryPath, err := agents.FindMCPBinary()
	if err != nil {
		return err
	}

	agentList, err := resolveMCPAgents(mcpAgentFlag)
	if err != nil {
		return err
	}
	if len(agentList) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No AI agents detected. Install one first:")
		for _, a := range agents.All() {
			fmt.Fprintf(cmd.OutOrStdout(), "  %-14s config dir: %s\n", a.Name, a.ConfigDir)
		}
		return nil
	}

	for _, a := range agentList {
		installed, err := a.MCP.IsInstalled(mcpServerName)
		if err != nil {
			return fmt.Errorf("checking %s: %w", a.Name, err)
		}
		if installed {
			fmt.Fprintf(cmd.OutOrStdout(), "SKIP: %s already registered for %s\n", mcpServerName, a.Name)
			continue
		}
		if err := a.MCP.Install(mcpServerName, binaryPath); err != nil {
			return fmt.Errorf("registering with %s: %w", a.Name, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "OK:   Registered %s → %s\n", mcpServerName, a.Name)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "\nUsing binary: %s\n", binaryPath)
	return nil
}

func runMCPRemove(cmd *cobra.Command, _ []string) error {
	agentList, err := resolveMCPAgents(mcpAgentFlag)
	if err != nil {
		return err
	}
	if len(agentList) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No AI agents detected.")
		return nil
	}

	for _, a := range agentList {
		installed, err := a.MCP.IsInstalled(mcpServerName)
		if err != nil {
			return fmt.Errorf("checking %s: %w", a.Name, err)
		}
		if !installed {
			fmt.Fprintf(cmd.OutOrStdout(), "SKIP: %s not registered for %s\n", mcpServerName, a.Name)
			continue
		}
		if err := a.MCP.Remove(mcpServerName); err != nil {
			return fmt.Errorf("removing from %s: %w", a.Name, err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "OK:   Removed %s from %s\n", mcpServerName, a.Name)
	}
	return nil
}

func runMCPStatus(cmd *cobra.Command, _ []string) error {
	allAgents := agents.All()
	detected := agents.Detected()
	detectedSet := make(map[string]bool)
	for _, a := range detected {
		detectedSet[a.ID] = true
	}

	header := fmt.Sprintf("%-14s  %-14s", "Agent", "MCP Server")
	fmt.Fprintln(cmd.OutOrStdout(), header)
	fmt.Fprintln(cmd.OutOrStdout(), strings.Repeat("-", len(header)))

	for _, a := range allAgents {
		var cell string
		if !detectedSet[a.ID] {
			cell = "-"
		} else {
			installed, err := a.MCP.IsInstalled(mcpServerName)
			if err != nil {
				cell = "error"
			} else if installed {
				cell = "registered"
			} else {
				cell = "not registered"
			}
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%-14s  %-14s\n", a.Name, cell)
	}
	return nil
}

func resolveMCPAgents(flag string) ([]agents.Agent, error) {
	if flag == "all" {
		return agents.Detected(), nil
	}
	a, err := agents.Find(flag)
	if err != nil {
		return nil, err
	}
	return []agents.Agent{a}, nil
}
