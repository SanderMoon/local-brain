package cmd

import (
	"fmt"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config <key> [value]",
	Short: "Get or set configuration values",
	Long: `Get or set configuration values.

Available keys:
  editor    Preferred text editor (e.g., nvim, vim, emacs, nano)

Examples:
  brain config editor              Show current editor
  brain config editor nvim         Set editor to nvim
  brain config editor vim          Set editor to vim
  brain config editor emacs        Set editor to emacs`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	key := args[0]

	// Get operation (no value provided)
	if len(args) == 1 {
		switch key {
		case "editor":
			editor := cfg.GetEditor()
			if editor == "" {
				fmt.Println("(not set - using auto-detection: nvim > vim > $EDITOR)")
			} else {
				fmt.Println(editor)
			}
			return nil
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}
	}

	// Set operation (value provided)
	value := args[1]
	switch key {
	case "editor":
		cfg.SetEditor(value)
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
		fmt.Printf("Editor set to: %s\n", value)
		return nil
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
}
