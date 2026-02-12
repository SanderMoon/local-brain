package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate brain files to portable markdown standards (dry-run by default)",
	Long: `Harmonizes existing brain files to portable markdown standards.

Dry-run by default - shows what would change without modifying files.
Use --apply to actually write the changes.

Operations:
  1. Add YAML frontmatter to notes that only have 'Created:' body text
  2. Move system task metadata (#id:, #captured:, #done:) into HTML comments
  3. Append missing relative links to notes.md index files

Safe to run multiple times - idempotent.`,
	RunE: runMigrate,
}

var (
	migrateApplyFlag   bool
	migrateProjectFlag string
	migrateSectionFlag string
)

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&migrateApplyFlag, "apply", false, "Write changes (default is dry-run)")
	migrateCmd.Flags().StringVar(&migrateProjectFlag, "project", "", "Migrate only this project")
	migrateCmd.Flags().StringVar(&migrateSectionFlag, "section", "01_active", "PARA section to migrate")
}

func runMigrate(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	sectionDir, err := config.GetSectionPath(cfg, migrateSectionFlag)
	if err != nil {
		return err
	}

	dryRun := !migrateApplyFlag

	if dryRun {
		fmt.Println("[DRY RUN] No changes written. Use --apply to write changes.")
	} else {
		fmt.Println("[APPLY] Writing changes...")
	}
	fmt.Println()

	var projectResults []api.ProjectMigrateResult

	if migrateProjectFlag != "" {
		projectDir := filepath.Join(sectionDir, migrateProjectFlag)
		single, err := api.MigrateProject(projectDir, dryRun)
		if err != nil {
			return fmt.Errorf("failed to migrate project %s: %w", migrateProjectFlag, err)
		}
		projectResults = []api.ProjectMigrateResult{single}
	} else {
		projectResults, err = api.MigrateAllProjects(sectionDir, dryRun)
		if err != nil {
			return fmt.Errorf("failed to migrate projects: %w", err)
		}
	}

	totalNotes := 0
	totalTodos := 0
	totalLinks := 0

	for _, pr := range projectResults {
		notesChanged := 0
		todosChanged := 0
		linksChanged := 0

		for _, nr := range pr.Notes {
			if nr.Changed {
				notesChanged++
			}
		}
		for _, tr := range pr.Todos {
			if tr.Changed {
				todosChanged++
			}
		}
		for _, lr := range pr.Links {
			if lr.Changed {
				linksChanged++
			}
		}

		if notesChanged == 0 && todosChanged == 0 && linksChanged == 0 {
			continue
		}

		fmt.Printf("Project: %s\n", pr.ProjectName)
		if notesChanged > 0 {
			changes := collectNoteChanges(pr.Notes)
			fmt.Printf("  Notes: %d change(s) (%s)\n", notesChanged, changes)
		}
		if todosChanged > 0 {
			changes := collectTodoChanges(pr.Todos)
			fmt.Printf("  Todos: %d change(s) (%s)\n", todosChanged, changes)
		}
		if linksChanged > 0 {
			changes := collectLinkChanges(pr.Links)
			fmt.Printf("  Links: %d change(s) (%s)\n", linksChanged, changes)
		}
		fmt.Println()

		totalNotes += notesChanged
		totalTodos += todosChanged
		totalLinks += linksChanged
	}

	fmt.Printf("Total: %d notes, %d todos, %d links changed\n", totalNotes, totalTodos, totalLinks)
	return nil
}

func collectNoteChanges(results []api.MigrateNoteResult) string {
	var changes []string
	seen := make(map[string]int)
	for _, r := range results {
		if r.Changed {
			seen[r.Change]++
		}
	}
	for change, count := range seen {
		if count > 1 {
			changes = append(changes, fmt.Sprintf("%s x%d", change, count))
		} else {
			changes = append(changes, change)
		}
	}
	if len(changes) == 0 {
		return "no changes"
	}
	return joinStrings(changes)
}

func collectTodoChanges(results []api.MigrateTodoResult) string {
	var changes []string
	for _, r := range results {
		if r.Changed {
			changes = append(changes, r.Change)
		}
	}
	if len(changes) == 0 {
		return "no changes"
	}
	return joinStrings(changes)
}

func collectLinkChanges(results []api.MigrateLinkResult) string {
	var changes []string
	for _, r := range results {
		if r.Changed {
			changes = append(changes, r.Change)
		}
	}
	if len(changes) == 0 {
		return "no changes"
	}
	return joinStrings(changes)
}

func joinStrings(ss []string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += ", "
		}
		result += s
	}
	return result
}
