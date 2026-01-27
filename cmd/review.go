package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "AI-powered project analysis (stub)",
	Long: `AI Project Analysis

This feature is not yet implemented.

Future functionality:
  - Analyze project notes and todos
  - Generate project summaries
  - Suggest next actions
  - Create .context.md files for AI agents

Planned integration:
  - LLM CLI tools (e.g., llm, aichat, gpt)
  - Auto-generate context summaries
  - Support for Model Context Protocol (MCP)

Implementation ideas:
  1. Read notes.md and todo.md from current project
  2. Pipe content to LLM with structured prompt
  3. Display analysis and suggestions
  4. Optionally update .context.md file

To implement:
  - Install LLM CLI: pip install llm (https://github.com/simonw/llm)
  - Edit this command to pipe project data to LLM
  - Example: cat notes.md todo.md | llm "Analyze this project and suggest priorities"

Current workaround:
  - Manually copy notes/tasks to your AI assistant
  - Use MCP filesystem server to give AI access to ~/brain`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func runReview(cmd *cobra.Command, args []string) error {
	fmt.Println("brain-review - AI Project Analysis")
	fmt.Println("")
	fmt.Println("This feature is not yet implemented.")
	fmt.Println("")
	fmt.Println("Future functionality:")
	fmt.Println("  - Analyze project notes and todos")
	fmt.Println("  - Generate project summaries")
	fmt.Println("  - Suggest next actions")
	fmt.Println("  - Create .context.md files for AI agents")
	fmt.Println("")
	fmt.Println("Planned integration:")
	fmt.Println("  - LLM CLI tools (e.g., llm, aichat, gpt)")
	fmt.Println("  - Auto-generate context summaries")
	fmt.Println("  - Support for Model Context Protocol (MCP)")
	fmt.Println("")
	fmt.Println("To implement:")
	fmt.Println("  - Install LLM CLI: pip install llm")
	fmt.Println("  - Example: cat notes.md todo.md | llm \"Analyze this project\"")
	fmt.Println("")
	fmt.Println("Current workaround:")
	fmt.Println("  - Manually copy notes/tasks to your AI assistant")
	fmt.Println("  - Use MCP filesystem server to give AI access to ~/brain")

	return nil
}
