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

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new brain",
	Long: `Create a new brain with standardized directory structure.

If no name is provided, defaults to 'default'.
Brains are created in ~/brains/<name> by default (can be overridden with BRAIN_ROOT).

Structure created:
  ~/brains/<name>/
  ├── 00_dump.md
  ├── 01_active/
  ├── 02_areas/
  ├── 03_resources/
  └── 99_archive/`,
	Example: `  brain new              # Creates ~/brains/default
  brain new work         # Creates ~/brains/work`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Get brain name
	brainName := "default"
	if len(args) > 0 {
		brainName = args[0]
	}

	// Validate brain name
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(brainName) {
		return fmt.Errorf("brain name can only contain letters, numbers, hyphens, and underscores")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Check if brain already exists in config
	if cfg.BrainExists(brainName) {
		existingPath, _ := cfg.GetBrainPath(brainName)
		if fileutil.FileExists(existingPath) {
			fmt.Printf("Error: Brain '%s' already configured and exists.\n", brainName)
			fmt.Printf("Location: %s\n", existingPath)
			fmt.Printf("Use 'brain switch %s' to activate it.\n", brainName)
			return nil
		}

		fmt.Printf("Warning: Brain '%s' is registered in config but missing on disk.\n", brainName)
		fmt.Println("Repairing/Re-initializing at standardized location...")
	}

	// Define standardized location
	brainRoot := os.Getenv("BRAIN_ROOT")
	if brainRoot == "" {
		brainRoot = filepath.Join(os.Getenv("HOME"), "brains")
	}

	location := filepath.Join(brainRoot, brainName)

	// Check directory collision
	if fileutil.FileExists(location) {
		// Check if it looks like a brain (adoption/recovery)
		activeDir := filepath.Join(location, "01_active")
		dumpFile := filepath.Join(location, "00_dump.md")

		if fileutil.FileExists(activeDir) && fileutil.FileExists(dumpFile) {
			fmt.Printf("Found existing brain data at %s.\n", location)
			fmt.Println("Adopting it into configuration...")
		} else {
			// Check if directory is empty
			entries, err := os.ReadDir(location)
			if err == nil && len(entries) > 0 {
				return fmt.Errorf("directory %s exists and is not empty", location)
			}
		}
	} else {
		// Create new brain structure
		fmt.Printf("Initializing brain '%s' at %s...\n", brainName, location)

		// Create directories
		dirs := []string{
			filepath.Join(location, "01_active"),
			filepath.Join(location, "02_areas"),
			filepath.Join(location, "03_resources"),
			filepath.Join(location, "99_archive"),
		}

		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}

		// Create dump file
		dumpContent := `# Dump

Quick capture landing zone. Process with ` + "`brain refile`" + `.

`
		dumpPath := filepath.Join(location, "00_dump.md")
		if err := os.WriteFile(dumpPath, []byte(dumpContent), 0644); err != nil {
			return fmt.Errorf("failed to create dump file: %w", err)
		}

		fmt.Println("OK: Created directory structure")
	}

	// Add to config
	if err := cfg.AddBrain(brainName, location); err != nil {
		return fmt.Errorf("failed to add brain to config: %w", err)
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("OK: Registered '%s'\n", brainName)

	// Auto-switch if it's the first brain or if it's named "default"
	current := cfg.GetCurrentBrain()
	if current == "" || brainName == "default" {
		if err := cfg.SetCurrentBrain(brainName); err != nil {
			return fmt.Errorf("failed to set current brain: %w", err)
		}

		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("OK: Set as current brain")
		fmt.Printf("OK: Updated symlink %s -> %s\n", config.GetSymlinkPath(), location)
	}

	fmt.Println("")
	fmt.Printf("Success! Brain '%s' is ready.\n", brainName)

	return nil
}
