package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var migrateIDsCmd = &cobra.Command{
	Use:   "migrate-ids",
	Short: "Add stable IDs to all todos that don't have them",
	Long: `Scans all project todos and dump file, adding #id: tags where missing.
This is a one-time migration for existing brain data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return err
		}

		// Migrate dump
		dumpPath := filepath.Join(brainPath, "00_dump.md")
		dumpCount, err := api.MigrateDumpFileIDs(dumpPath)
		if err != nil {
			fmt.Printf("Warning: dump migration failed: %v\n", err)
		} else if dumpCount > 0 {
			fmt.Printf("Migrated %d items in dump\n", dumpCount)
		}

		// Migrate all projects
		activeDir := filepath.Join(brainPath, "01_active")
		projectResults, err := api.MigrateAllProjectTodos(activeDir)
		if err != nil {
			return fmt.Errorf("failed to migrate projects: %w", err)
		}

		if len(projectResults) > 0 {
			fmt.Println("\nProject todos migrated:")
			totalTodos := 0
			for project, count := range projectResults {
				fmt.Printf("  %s: %d todos\n", project, count)
				totalTodos += count
			}
			fmt.Printf("\nTotal: %d todos migrated\n", totalTodos+dumpCount)
		} else {
			if dumpCount == 0 {
				fmt.Println("No items needed migration - all IDs are already present")
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateIDsCmd)
}
