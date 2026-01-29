package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var (
	tagRemoveFlag bool
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags with counts",
	Long: `List all tags used across tasks with their occurrence counts.

Helps discover existing tags for autocompletion and consistency.`,
	Example: `  brain todo tags  # List all tags with counts`,
	RunE:    runTags,
}

var tagCmd = &cobra.Command{
	Use:   "tag [ID] [TAG...]",
	Short: "Add or remove tags",
	Long: `Add or remove tags from a task.

Tags are freeform hashtags (e.g., #bug, #feature, #urgent).
Use --rm flag to remove tags instead of adding them.

Interactive mode (no arguments):
  Fuzzy search through all tasks and add tags interactively

Direct mode (ID and tags):
  Add specified tags to the task`,
	Example: `  brain todo tag                      # Interactive selection
  brain todo tag abc123 bug feature   # Add tags
  brain todo tag abc123 --rm bug      # Remove tags
  brain todo tag abc123 security ui   # Add multiple tags`,
	Args: cobra.MinimumNArgs(0),
	RunE: runTag,
}

func init() {
	todoCmd.AddCommand(tagsCmd)
	todoCmd.AddCommand(tagCmd)

	tagCmd.Flags().BoolVar(&tagRemoveFlag, "rm", false, "Remove tags instead of adding")
}

func runTags(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	todos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}

	tagCounts := api.ListAllTags(todos)

	if len(tagCounts) == 0 {
		fmt.Println("No tags found")
		return nil
	}

	// Sort tags by count (descending), then alphabetically
	type tagCount struct {
		tag   string
		count int
	}
	var tags []tagCount
	for tag, count := range tagCounts {
		tags = append(tags, tagCount{tag, count})
	}

	sort.Slice(tags, func(i, j int) bool {
		if tags[i].count != tags[j].count {
			return tags[i].count > tags[j].count
		}
		return tags[i].tag < tags[j].tag
	})

	// Display
	fmt.Printf("%-20s %s\n", "TAG", "COUNT")
	fmt.Println(strings.Repeat("-", 30))
	for _, tc := range tags {
		fmt.Printf("%-20s %d\n", "#"+tc.tag, tc.count)
	}

	return nil
}

func runTag(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	// Route based on arguments
	if len(args) == 0 {
		// Interactive mode
		return runTagInteractive(activeDir)
	} else if len(args) == 1 {
		return fmt.Errorf("usage: brain todo tag <id> <tag...> or brain todo tag (interactive)")
	} else {
		// Direct mode: brain todo tag <id> <tag1> <tag2> ...
		return runTagDirect(activeDir, args[0], args[1:])
	}
}

func runTagInteractive(activeDir string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Select a todo (only open tasks)
		todo, err := selectTodo(activeDir, "open", "Select task to tag (Esc to exit)")
		if err != nil {
			// Check if user cancelled (FZF returns error on Esc)
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Show current tags
		fmt.Printf("Task: %s (%s)\n", todo.Content, todo.Project)
		if len(todo.Tags) > 0 {
			fmt.Printf("Current tags: %s\n", formatTags(todo.Tags))
		} else {
			fmt.Println("Current tags: none")
		}

		// Get all existing tags for suggestions
		todos, _ := api.ParseAllTodos(activeDir, false)
		allTags := api.ListAllTags(todos)
		var suggestions []string
		for tag := range allTags {
			suggestions = append(suggestions, tag)
		}
		sort.Strings(suggestions)

		if len(suggestions) > 0 {
			fmt.Printf("Existing tags: %s\n", strings.Join(suggestions, ", "))
		}

		fmt.Print("Enter tags to add (space separated, or 'rm <tags>' to remove): ")

		// Read input
		var input string
		_, _ = fmt.Scanln(&input)
		if input == "" {
			fmt.Println("Cancelled")
			continue
		}

		// Parse input
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		// Check if removing tags
		if parts[0] == "rm" || parts[0] == "remove" {
			if len(parts) < 2 {
				fmt.Println("Error: Specify tags to remove")
				continue
			}
			tagsToRemove := parts[1:]
			if err := api.RemoveTodoTags(todo, tagsToRemove); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("OK: Removed tags %s from: %s\n", formatTags(tagsToRemove), todo.Content)
		} else {
			// Add tags
			tagsToAdd := parts
			if err := api.AddTodoTags(todo, tagsToAdd); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("OK: Added tags %s to: %s\n", formatTags(tagsToAdd), todo.Content)
		}

		fmt.Println("") // Add blank line for readability
	}
}

func runTagDirect(activeDir, query string, tags []string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	if tagRemoveFlag {
		// Remove tags
		if err := api.RemoveTodoTags(todo, tags); err != nil {
			return fmt.Errorf("failed to remove tags: %w", err)
		}
		fmt.Printf("OK: Removed tags %s from: %s (%s)\n", formatTags(tags), todo.Content, todo.Project)
	} else {
		// Add tags
		if err := api.AddTodoTags(todo, tags); err != nil {
			return fmt.Errorf("failed to add tags: %w", err)
		}
		fmt.Printf("OK: Added tags %s to: %s (%s)\n", formatTags(tags), todo.Content, todo.Project)
	}

	return nil
}

// formatTags formats a slice of tags for display
func formatTags(tags []string) string {
	var formatted []string
	for _, tag := range tags {
		formatted = append(formatted, "#"+tag)
	}
	return strings.Join(formatted, " ")
}
