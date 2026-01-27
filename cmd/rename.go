package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var renameCmd = &cobra.Command{
	Use:   "rename <old-name> <new-name>",
	Short: "Rename an existing brain",
	Long: `Rename an existing brain.

What it does:
  1. Renames the directory in ~/brains/
  2. Updates ~/.config/brain/config.json
  3. Updates ~/brain symlink if active`,
	Example: `  brain rename work old-work
  brain rename default personal`,
	Args: cobra.ExactArgs(2),
	RunE: runRename,
}

func init() {
	rootCmd.AddCommand(renameCmd)
}

func runRename(cmd *cobra.Command, args []string) error {
	oldName := args[0]
	newName := args[1]

	// Validate new name
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(newName) {
		return fmt.Errorf("new name contains invalid characters. Use alphanumerics, hyphens, underscores")
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check old brain exists
	if !cfg.BrainExists(oldName) {
		return fmt.Errorf("brain '%s' does not exist", oldName)
	}

	// Check new name doesn't exist
	if cfg.BrainExists(newName) {
		return fmt.Errorf("brain '%s' already exists", newName)
	}

	oldPath, err := cfg.GetBrainPath(oldName)
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	// Determine new path (same parent directory)
	parentDir := filepath.Dir(oldPath)
	newPath := filepath.Join(parentDir, newName)

	if fileutil.FileExists(newPath) {
		return fmt.Errorf("target directory '%s' already exists", newPath)
	}

	fmt.Printf("Renaming '%s' to '%s'...\n", oldName, newName)

	// 1. Move directory
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory: %w", err)
	}
	fmt.Printf("OK: Moved directory to %s\n", newPath)

	// 2. Update config
	if err := cfg.RenameBrain(oldName, newName, newPath); err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Println("OK: Updated configuration")

	// 3. Update symlink if necessary
	currentBrain := cfg.GetCurrentBrain()
	if currentBrain == newName {
		if err := config.UpdateSymlink(newName, cfg); err != nil {
			return fmt.Errorf("failed to update symlink: %w", err)
		}
		fmt.Println("OK: Updated active symlink")
	}

	fmt.Println("")
	fmt.Println("Success! Brain renamed.")

	return nil
}
