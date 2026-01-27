package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var dumpJSONFlag bool

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Inspect dump contents",
	Long: `Display items in the dump file.

Subcommands:
  ls       List items in human-readable table format
  ls --json  List items in JSON format for programmatic access`,
	Example: `  brain dump ls
  brain dump ls --json | jq '.[] | select(.type=="todo")'
  brain dump ls --json | jq -r '.[].id'`,
}

var dumpLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List dump items",
	Long: `List all items in the dump file.

Output formats:
  Human-readable (default): Table with ID, type, and content
  JSON (--json): Machine-readable array of objects`,
	Example: `  brain dump ls
  brain dump ls --json`,
	RunE: runDumpLs,
}

func init() {
	rootCmd.AddCommand(dumpCmd)
	dumpCmd.AddCommand(dumpLsCmd)

	dumpLsCmd.Flags().BoolVar(&dumpJSONFlag, "json", false, "Output JSON format")
}

func runDumpLs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	dumpPath := filepath.Join(brainPath, "00_dump.md")

	if !fileutil.FileExists(dumpPath) {
		if dumpJSONFlag {
			fmt.Println("[]")
		} else {
			return fmt.Errorf("dump not found at %s", dumpPath)
		}
		return nil
	}

	if dumpJSONFlag {
		return outputJSON(dumpPath)
	}

	return outputTable(dumpPath)
}

func outputJSON(dumpPath string) error {
	items, err := api.ParseDumpToJSON(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to parse dump: %w", err)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func outputTable(dumpPath string) error {
	items, err := api.ParseDumpToJSON(dumpPath)
	if err != nil {
		return fmt.Errorf("failed to parse dump: %w", err)
	}

	// Print table header
	fmt.Printf("%-8s | %-6s | %s\n", "ID", "Type", "Content")
	fmt.Println("--------+--------+----------")

	if len(items) == 0 {
		fmt.Println("(No items in dump)")
		return nil
	}

	// Print items
	for _, item := range items {
		// Add suffix for multi-line notes
		suffix := ""
		if item.Type == "note" && item.EndLine > item.StartLine {
			lines := item.EndLine - item.StartLine + 1
			suffix = fmt.Sprintf(" (%d lines)", lines)
		}

		// Truncate long content
		content := item.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		fmt.Printf("%-8s | %-6s | %s%s\n", item.ID, item.Type, content, suffix)
	}

	return nil
}
