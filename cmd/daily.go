package cmd

import (
	"fmt"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var dailyOpenFlag bool

var dailyCmd = &cobra.Command{
	Use:   "daily",
	Short: "Create or open today's daily note",
	Long: `Create today's daily note in {brain}/00_daily/YYYY-MM-DD.md.
If the note already exists, prints the path.
Includes a briefing section with overdue todos.`,
	RunE: runDaily,
}

func init() {
	rootCmd.AddCommand(dailyCmd)
	dailyCmd.Flags().BoolVar(&dailyOpenFlag, "open", false, "Open note in editor after creating")
}

func runDaily(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir, err := config.GetProjectsPath(cfg)
	if err != nil {
		return fmt.Errorf("failed to get projects path: %w", err)
	}

	today := time.Now().Format("2006-01-02")

	allTodos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}

	var overdueTodos []api.TodoItem
	for _, todo := range allTodos {
		if todo.DueDate != "" && todo.DueDate < today {
			overdueTodos = append(overdueTodos, todo)
		}
	}

	result, err := api.CreateOrOpenDailyNote(brainPath, today, overdueTodos)
	if err != nil {
		return fmt.Errorf("failed to create daily note: %w", err)
	}

	fmt.Printf("Daily note: %s\n", result.Path)
	if result.IsNew {
		fmt.Println("Created new daily note.")
	}

	if dailyOpenFlag {
		if err := external.OpenFile(result.Path); err != nil {
			return fmt.Errorf("failed to open editor: %w", err)
		}
	}

	return nil
}
