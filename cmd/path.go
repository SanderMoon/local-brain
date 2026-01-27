package cmd

import (
	"fmt"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path [brain-name]",
	Short: "Show path to current brain or specified brain",
	Long: `Display the filesystem path to a brain.

If no brain name is provided, shows the path to the current active brain.
Otherwise, shows the path to the specified brain.`,
	Example: `  brain path           # Show current brain path
  brain path work      # Show path to 'work' brain`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPath,
}

func init() {
	rootCmd.AddCommand(pathCmd)
}

func runPath(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var brainName string
	if len(args) == 0 {
		// Use current brain
		brainName = cfg.GetCurrentBrain()
		if brainName == "" {
			return fmt.Errorf("no current brain set")
		}
	} else {
		brainName = args[0]
	}

	path, err := cfg.GetBrainPath(brainName)
	if err != nil {
		return fmt.Errorf("brain '%s' not found", brainName)
	}

	fmt.Println(path)

	return nil
}
