package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/fileutil"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import [root-path]",
	Short: "Import existing brains",
	Long: `Scan a directory for brain structures and register them in configuration.

Scans the directory (default: ~/brains) for brain structures
and registers them in the local configuration.

A valid brain structure must contain:
  - 01_active/ directory
  - 00_dump.md file`,
	Example: `  brain import             # Scan ~/brains
  brain import ~/Dropbox/Brains`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func runImport(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Determine root to scan
	scanRoot := filepath.Join(os.Getenv("HOME"), "brains")
	if brainRoot := os.Getenv("BRAIN_ROOT"); brainRoot != "" {
		scanRoot = brainRoot
	}
	if len(args) > 0 {
		scanRoot = args[0]
	}

	// Expand ~ in path
	if expandedPath, err := fileutil.ExpandPath(scanRoot); err == nil {
		scanRoot = expandedPath
	}

	if !fileutil.FileExists(scanRoot) {
		return fmt.Errorf("directory '%s' not found", scanRoot)
	}

	fmt.Printf("Scanning '%s' for brains...\n", scanRoot)
	fmt.Println("")

	importedCount := 0
	var lastImported string

	// Iterate directories
	entries, err := os.ReadDir(scanRoot)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		brainName := entry.Name()
		brainDir := filepath.Join(scanRoot, brainName)

		// Validate: Look for signature files
		activeDir := filepath.Join(brainDir, "01_active")
		dumpFile := filepath.Join(brainDir, "00_dump.md")

		if !fileutil.FileExists(activeDir) || !fileutil.FileExists(dumpFile) {
			continue
		}

		// Check if already registered
		if cfg.BrainExists(brainName) {
			currentPath, _ := cfg.GetBrainPath(brainName)
			if currentPath == brainDir {
				fmt.Printf(" [Skip] '%s' already registered.\n", brainName)
			} else {
				fmt.Printf(" [Skip] '%s' exists but points to different path (%s).\n", brainName, currentPath)
			}
			continue
		}

		// Register it
		if err := cfg.AddBrain(brainName, brainDir); err != nil {
			fmt.Printf(" [Error] Failed to import '%s': %v\n", brainName, err)
			continue
		}

		fmt.Printf(" [OK]   Imported '%s'\n", brainName)
		importedCount++
		lastImported = brainName
	}

	// Save config
	if importedCount > 0 {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}

	fmt.Println("")
	if importedCount == 0 {
		fmt.Println("No new brains found.")
	} else {
		fmt.Printf("Successfully imported %d brain(s).\n", importedCount)
		fmt.Println("Use 'brain list' to view them.")

		// Auto-switch if no current brain set
		if cfg.GetCurrentBrain() == "" && lastImported != "" {
			fmt.Printf("Setting '%s' as active brain...\n", lastImported)
			if err := cfg.SetCurrentBrain(lastImported); err != nil {
				fmt.Printf("Warning: Failed to set current brain: %v\n", err)
			} else {
				if err := cfg.Save(); err != nil {
					fmt.Printf("Warning: Failed to save config: %v\n", err)
				}
			}
		}
	}

	return nil
}
