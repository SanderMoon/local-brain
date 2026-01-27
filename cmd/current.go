package cmd

import (
	"fmt"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var (
	currentNameOnly bool
	currentPathOnly bool
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current active brain",
	Long: `Display the currently active brain name and location.

The current brain is determined by the configuration at
~/.config/brain/config.json and the ~/brain symlink.`,
	Example: `  brain current              # Show name and path
  brain current --name-only  # Just the name
  brain current --path-only  # Just the path`,
	RunE: runCurrent,
}

func init() {
	rootCmd.AddCommand(currentCmd)

	currentCmd.Flags().BoolVar(&currentNameOnly, "name-only", false, "Output only the brain name")
	currentCmd.Flags().BoolVar(&currentPathOnly, "path-only", false, "Output only the path")
}

func runCurrent(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	current := cfg.GetCurrentBrain()
	if current == "" {
		fmt.Println("No brain currently active")
		fmt.Println("")
		fmt.Println("Initialize a brain:")
		fmt.Println("  brain new <name> <path>")
		return nil
	}

	path, err := cfg.GetBrainPath(current)
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	// Output based on flags
	if currentNameOnly {
		fmt.Println(current)
	} else if currentPathOnly {
		fmt.Println(path)
	} else {
		fmt.Printf("Current brain: %s\n", current)
		fmt.Printf("Location: %s\n", path)
	}

	return nil
}
