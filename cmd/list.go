package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var (
	listShowPaths bool
	listJSONFlag  bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured brains",
	Long: `Display all configured brains with the current brain marked.

Lists brain names and optionally their file system paths.
The currently active brain is marked with an asterisk (*).`,
	Example: `  brain list           # Simple list
  brain list --paths   # Include full paths
  brain list --json    # JSON output`,
	RunE: runList,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVar(&listShowPaths, "paths", false, "Show full paths")
	listCmd.Flags().BoolVar(&listJSONFlag, "json", false, "Output JSON format")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brains := cfg.ListBrains()
	current := cfg.GetCurrentBrain()

	// JSON output
	if listJSONFlag {
		type BrainInfo struct {
			Name    string `json:"name"`
			Path    string `json:"path"`
			Current bool   `json:"current"`
		}

		var brainsJSON []BrainInfo
		for _, brain := range brains {
			path, err := cfg.GetBrainPath(brain)
			if err != nil {
				continue
			}
			brainsJSON = append(brainsJSON, BrainInfo{
				Name:    brain,
				Path:    path,
				Current: brain == current,
			})
		}

		data, err := json.MarshalIndent(brainsJSON, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	if len(brains) == 0 {
		symlinkPath := config.GetSymlinkPath()

		// Check if default symlink exists
		if _, err := config.GetCurrentSymlinkTarget(); err == nil {
			fmt.Println("No configured brains.")
			fmt.Printf("Using default shadow brain at %s\n", symlinkPath)
			fmt.Println("(Run 'brain new <name> <path>' to formalize)")
		} else {
			fmt.Println("No brains configured")
			fmt.Println("")
			fmt.Println("Initialize your first brain:")
			fmt.Println("  brain new <name> <path>")
		}

		return nil
	}

	// Sort brains alphabetically
	sort.Strings(brains)

	fmt.Println("Configured brains:")
	fmt.Println("")

	for _, brain := range brains {
		path, err := cfg.GetBrainPath(brain)
		if err != nil {
			continue
		}

		marker := " "
		if brain == current {
			marker = "*"
		}

		if listShowPaths {
			fmt.Printf(" %s %-20s %s\n", marker, brain, path)
		} else {
			if brain == current {
				fmt.Printf(" %s %s (current)\n", marker, brain)
			} else {
				fmt.Printf(" %s %s\n", marker, brain)
			}
		}
	}

	if current != "" {
		fmt.Println("")
		fmt.Println("* = current active brain")
	}

	return nil
}
