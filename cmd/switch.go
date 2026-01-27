package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [brain-name]",
	Short: "Switch to a different brain",
	Long: `Switch to a different brain configuration.

If no brain name is provided, shows an interactive list of available brains.
After switching:
  - All brain commands operate on the new brain
  - ~/brain symlink points to new location
  - Shell prompt updates (if configured)`,
	Example: `  brain switch work      # Switch to 'work' brain
  brain switch           # Interactive selection
  brain switch default   # Switch back to default`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brains := cfg.ListBrains()
	if len(brains) == 0 {
		fmt.Println("No brains configured")
		fmt.Println("")
		fmt.Println("Initialize a brain first:")
		fmt.Println("  brain new <name>")
		return nil
	}

	// Sort brains alphabetically
	sort.Strings(brains)

	var targetBrain string
	if len(args) == 0 {
		// Interactive mode
		targetBrain, err = selectBrainInteractive(cfg, brains)
		if err != nil {
			return err
		}
	} else {
		targetBrain = args[0]
	}

	// Validate brain exists
	if !cfg.BrainExists(targetBrain) {
		fmt.Printf("Error: Brain '%s' not found\n", targetBrain)
		fmt.Println("")
		fmt.Println("Available brains:")
		for _, brain := range brains {
			fmt.Printf("  - %s\n", brain)
		}
		fmt.Println("")
		fmt.Println("To create new brain: brain new <name>")
		return nil
	}

	// Check if already current
	current := cfg.GetCurrentBrain()
	if targetBrain == current {
		path, _ := cfg.GetBrainPath(targetBrain)
		fmt.Printf("Already using '%s'\n", targetBrain)
		fmt.Printf("Location: %s\n", path)
		return nil
	}

	// Switch to new brain
	if err := cfg.SetCurrentBrain(targetBrain); err != nil {
		return fmt.Errorf("failed to switch brain: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	path, _ := cfg.GetBrainPath(targetBrain)
	fmt.Printf("OK: Switched to '%s'\n", targetBrain)
	fmt.Printf("  Location: %s\n", path)
	fmt.Printf("  Symlink: ~/brain -> %s\n", path)

	return nil
}

func selectBrainInteractive(cfg *config.Config, brains []string) (string, error) {
	fmt.Println("Available brains:")
	fmt.Println("")

	current := cfg.GetCurrentBrain()

	for i, brain := range brains {
		path, _ := cfg.GetBrainPath(brain)
		suffix := ""
		if brain == current {
			suffix = " (current)"
		}
		fmt.Printf("  %d) %s -> %s%s\n", i+1, brain, path, suffix)
	}

	fmt.Println("")
	fmt.Printf("Select brain [1-%d] or name: ", len(brains))

	reader := bufio.NewReader(os.Stdin)
	selection, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	selection = strings.TrimSpace(selection)

	// Check if number or name
	if num, err := strconv.Atoi(selection); err == nil {
		// It's a number
		idx := num - 1
		if idx >= 0 && idx < len(brains) {
			return brains[idx], nil
		}
		return "", fmt.Errorf("invalid selection: %d", num)
	}

	// It's a name
	return selection, nil
}
