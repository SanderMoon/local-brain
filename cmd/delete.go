package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Permanently delete a brain",
	Long: `Permanently delete a brain and all its contents.

WARNING: This will permanently delete the brain directory and all its contents.
This action cannot be undone.`,
	Example: `  brain delete work
  brain delete default`,
	Args: cobra.ExactArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	brainName := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate existence
	if !cfg.BrainExists(brainName) {
		return fmt.Errorf("brain '%s' does not exist", brainName)
	}

	brainPath, err := cfg.GetBrainPath(brainName)
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	// Warning
	fmt.Printf("WARNING: You are about to PERMANENTLY DELETE the brain '%s'.\n", brainName)
	fmt.Printf("Location: %s\n", brainPath)
	fmt.Println("This includes all projects, notes, and tasks within it.")
	fmt.Println("")
	fmt.Print("Are you sure? This cannot be undone. [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Aborted.")
		return nil
	}

	fmt.Println("")
	fmt.Printf("Deleting '%s'...\n", brainName)

	// 1. Update config (remove entry and clear current if needed)
	if err := cfg.DeleteBrain(brainName); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("OK: Removed from configuration")

	// 2. Check and remove symlink if it points to this brain
	symlinkPath := config.GetSymlinkPath()
	if isLink, _ := fileutil.IsSymlink(symlinkPath); isLink {
		if target, err := os.Readlink(symlinkPath); err == nil && target == brainPath {
			if err := os.Remove(symlinkPath); err == nil {
				fmt.Println("OK: Removed active symlink")
			}
		}
	}

	// 3. Delete directory
	if fileutil.FileExists(brainPath) {
		if err := os.RemoveAll(brainPath); err != nil {
			return fmt.Errorf("failed to delete directory: %w", err)
		}
		fmt.Println("OK: Deleted directory")
	} else {
		fmt.Println("Warning: Directory was already missing.")
	}

	fmt.Println("")
	fmt.Printf("Brain '%s' deleted successfully.\n", brainName)

	return nil
}
