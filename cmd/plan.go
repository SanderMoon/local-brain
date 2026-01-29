package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/dateutil"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Interactive batch task planning",
	Long: `Interactive workflow for enriching tasks with metadata.

Loops through tasks with FZF selection, prompting for:
  - Priority (1/2/3)
  - Due date (YYYY-MM-DD, tomorrow, +3d, next-friday)
  - Tags (comma separated, autocomplete from existing)
  - State (open/in-progress/blocked)

All fields are optional - press Enter to skip.
Ideal for weekly planning sessions.

Complements 'brain add' for the capture-curate workflow:
  - Capture fast: brain add "task"
  - Curate later: brain plan`,
	Example: `  brain plan  # Interactive batch planning`,
	RunE:    runPlan,
}

func init() {
	rootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Get all existing tags for suggestions
	allTodos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}
	allTagsMap := api.ListAllTags(allTodos)
	var existingTags []string
	for tag := range allTagsMap {
		existingTags = append(existingTags, tag)
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Refresh todos to get latest state
		todos, err := api.ParseAllTodos(activeDir, false)
		if err != nil {
			return fmt.Errorf("failed to parse todos: %w", err)
		}

		// Prioritize unprioritized and unscheduled tasks
		var filtered []api.TodoItem
		for _, todo := range todos {
			if todo.Status == "open" || todo.Status == "in-progress" {
				// Prioritize tasks without metadata
				if todo.Priority == nil || todo.DueDate == "" || len(todo.Tags) == 0 {
					filtered = append(filtered, todo)
				}
			}
		}

		// If no unprioritized tasks, include all open tasks
		if len(filtered) == 0 {
			for _, todo := range todos {
				if todo.Status == "open" || todo.Status == "in-progress" {
					filtered = append(filtered, todo)
				}
			}
		}

		if len(filtered) == 0 {
			fmt.Println("No tasks to plan")
			return nil
		}

		// Select a todo with FZF
		todo, err := selectTodoFromList(filtered, "Select task to plan (Esc to exit)")
		if err != nil {
			// Check if user cancelled
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Show current task state
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Printf("Task: %s (%s)\n", todo.Content, todo.Project)
		fmt.Println(strings.Repeat("-", 60))

		// Show current metadata
		currentPrio := "none"
		if todo.Priority != nil {
			currentPrio = fmt.Sprintf("%d (%s)", *todo.Priority, getPriorityName(*todo.Priority))
		}
		fmt.Printf("Current priority: %s\n", currentPrio)

		currentDue := "none"
		if todo.DueDate != "" {
			currentDue = todo.DueDate
		}
		fmt.Printf("Current due date: %s\n", currentDue)

		currentTags := "none"
		if len(todo.Tags) > 0 {
			currentTags = formatTags(todo.Tags)
		}
		fmt.Printf("Current tags: %s\n", currentTags)
		fmt.Printf("Current state: %s\n", todo.Status)
		fmt.Println(strings.Repeat("-", 60))

		// Prompt for priority
		priority := promptForPriority()
		if priority != nil {
			if err := api.SetTodoPriority(todo, priority); err != nil {
				fmt.Printf("Error setting priority: %v\n", err)
			} else {
				if priority == nil {
					fmt.Println("✓ Cleared priority")
				} else {
					fmt.Printf("✓ Set priority to %d (%s)\n", *priority, getPriorityName(*priority))
				}
			}
		}

		// Prompt for due date
		dueDate := promptForDueDate()
		if dueDate != "" {
			if err := api.SetTodoDueDate(todo, dueDate); err != nil {
				fmt.Printf("Error setting due date: %v\n", err)
			} else {
				if dueDate == "clear" {
					fmt.Println("✓ Cleared due date")
				} else {
					fmt.Printf("✓ Set due date to %s\n", dueDate)
				}
			}
		}

		// Prompt for tags
		if len(existingTags) > 0 {
			fmt.Printf("(Existing tags: %s)\n", strings.Join(existingTags, ", "))
		}
		tags := promptForTags()
		if len(tags) > 0 {
			if err := api.AddTodoTags(todo, tags); err != nil {
				fmt.Printf("Error adding tags: %v\n", err)
			} else {
				fmt.Printf("✓ Added tags: %s\n", formatTags(tags))
			}
		}

		// Prompt for state
		state := promptForState()
		if state != "" {
			if err := api.SetTodoStatus(todo, state); err != nil {
				fmt.Printf("Error setting state: %v\n", err)
			} else {
				fmt.Printf("✓ Set state to %s\n", state)
			}
		}

		fmt.Println("") // Blank line for readability
	}
}

func promptForPriority() *int {
	fmt.Print("Priority (1=high, 2=medium, 3=low, clear, or Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return nil // Skip
	}

	if input == "clear" || input == "0" || input == "none" {
		cleared := 0
		return &cleared // Signal to clear
	}

	p, err := strconv.Atoi(input)
	if err != nil || p < 1 || p > 3 {
		fmt.Println("Invalid priority, skipping")
		return nil
	}

	return &p
}

func promptForDueDate() string {
	fmt.Print("Due date (YYYY-MM-DD, tomorrow, +3d, next-friday, clear, or Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return "" // Skip
	}

	if input == "clear" || input == "none" {
		return "clear"
	}

	// Parse natural language date
	parsed, err := dateutil.ParseNaturalDate(input)
	if err != nil {
		fmt.Printf("Invalid date format (%v), skipping\n", err)
		return ""
	}

	return parsed
}

func promptForTags() []string {
	fmt.Print("Tags (comma or space separated, or Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return nil // Skip
	}

	// Split by comma or space
	var tags []string
	input = strings.ReplaceAll(input, ",", " ")
	parts := strings.Fields(input)

	for _, part := range parts {
		// Remove # if user included it
		part = strings.TrimPrefix(part, "#")
		if part != "" {
			tags = append(tags, part)
		}
	}

	return tags
}

func promptForState() string {
	fmt.Print("State (open, in-progress, blocked, or Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return "" // Skip
	}

	validStates := []string{"open", "in-progress", "blocked"}
	for _, s := range validStates {
		if s == input {
			return input
		}
	}

	fmt.Println("Invalid state, skipping")
	return ""
}

// selectTodoFromList selects a todo from a pre-filtered list
func selectTodoFromList(todos []api.TodoItem, prompt string) (*api.TodoItem, error) {
	if len(todos) == 0 {
		return nil, fmt.Errorf("no matching tasks found")
	}

	// Sort by priority in reverse (unprioritized first for FZF cursor)
	sortTodosByPriorityReverse(todos)

	// Format for FZF
	var items []string
	todoMap := make(map[string]*api.TodoItem)
	for i := range todos {
		todo := &todos[i]
		statusMark := formatStatusMark(todo.Status)
		prioBadge := formatPriorityBadge(todo.Priority)

		// Build display with metadata indicators
		display := fmt.Sprintf("%s %s %s %s", todo.ID, prioBadge, statusMark, todo.Content)

		// Add metadata indicators
		var indicators []string
		if todo.DueDate != "" {
			indicators = append(indicators, "📅"+todo.DueDate)
		}
		if len(todo.Tags) > 0 {
			indicators = append(indicators, formatTags(todo.Tags))
		}
		if len(indicators) > 0 {
			display += " [" + strings.Join(indicators, " ") + "]"
		}

		display += fmt.Sprintf(" (%s)", todo.Project)

		items = append(items, display)
		todoMap[todo.ID] = todo
	}

	// Select with FZF
	selected, err := external.SelectOne(items, external.FZFOptions{
		Header:        prompt + " (Esc to cancel)",
		Preview:       "",
		PreviewWindow: "",
	})

	if err != nil {
		return nil, err
	}

	// Extract ID from selection (first field)
	parts := strings.Fields(selected)
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid selection format")
	}

	todoID := parts[0]
	selectedTodo, ok := todoMap[todoID]
	if !ok {
		return nil, fmt.Errorf("todo not found: %s", todoID)
	}

	return selectedTodo, nil
}
