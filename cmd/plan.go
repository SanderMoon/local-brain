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

Select tasks with FZF (Tab to select multiple, Enter to confirm), then set:
  - Priority (1/2/3)
  - Due date (YYYY-MM-DD, tomorrow, +3d, next-friday)
  - Tags (comma separated, autocomplete from existing)
  - State (open/in-progress/blocked/done)

All fields are optional - press Enter to skip.

Multi-select workflow:
  - Use Tab to select/deselect tasks (moves cursor up for easy spam-selecting)
  - Select one task for individual planning
  - Select multiple tasks to batch-apply the same metadata
  - Perfect for similar tasks (e.g., multiple reading tasks, similar bugs)

Complements 'brain add' for the capture-curate workflow:
  - Capture fast: brain add "task"
  - Curate later: brain plan`,
	Example: `  brain plan  # Interactive planning with multi-select`,
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

	// Get todos
	todos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}

	// Filter for planning
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

	// Multi-select todos (now default behavior)
	selectedTodos, err := selectMultipleTodos(filtered, "Select tasks to plan (Tab to select, Enter to confirm, Esc to cancel)")
	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}

	if len(selectedTodos) == 0 {
		fmt.Println("No tasks selected")
		return nil
	}

	// Show selected tasks
	fmt.Println("\n" + strings.Repeat("=", 60))
	if len(selectedTodos) == 1 {
		fmt.Printf("Selected task: %s (%s)\n", selectedTodos[0].Content, selectedTodos[0].Project)
	} else {
		fmt.Printf("Selected %d tasks:\n", len(selectedTodos))
		for i, todo := range selectedTodos {
			fmt.Printf("  %d. %s (%s)\n", i+1, todo.Content, todo.Project)
		}
	}
	fmt.Println(strings.Repeat("-", 60))

	// Prompt for shared metadata
	if len(selectedTodos) == 1 {
		fmt.Println("Set metadata for this task:")
	} else {
		fmt.Println("Apply metadata to all selected tasks:")
	}
	fmt.Println("")

	// Prompt for priority
	priority := promptForPriority()
	if priority != nil {
		for _, todo := range selectedTodos {
			if err := api.SetTodoPriority(todo, priority); err != nil {
				fmt.Printf("Error setting priority for %s: %v\n", todo.ID, err)
			}
		}
		if priority != nil && *priority == 0 {
			if len(selectedTodos) == 1 {
				fmt.Println("✓ Cleared priority")
			} else {
				fmt.Printf("✓ Cleared priority for %d tasks\n", len(selectedTodos))
			}
		} else if priority != nil {
			if len(selectedTodos) == 1 {
				fmt.Printf("✓ Set priority to %d (%s)\n", *priority, getPriorityName(*priority))
			} else {
				fmt.Printf("✓ Set priority to %d (%s) for %d tasks\n", *priority, getPriorityName(*priority), len(selectedTodos))
			}
		}
	}

	// Prompt for due date
	dueDate := promptForDueDate()
	if dueDate != "" {
		for _, todo := range selectedTodos {
			if err := api.SetTodoDueDate(todo, dueDate); err != nil {
				fmt.Printf("Error setting due date for %s: %v\n", todo.ID, err)
			}
		}
		if dueDate == "clear" {
			if len(selectedTodos) == 1 {
				fmt.Println("✓ Cleared due date")
			} else {
				fmt.Printf("✓ Cleared due date for %d tasks\n", len(selectedTodos))
			}
		} else {
			if len(selectedTodos) == 1 {
				fmt.Printf("✓ Set due date to %s\n", dueDate)
			} else {
				fmt.Printf("✓ Set due date to %s for %d tasks\n", dueDate, len(selectedTodos))
			}
		}
	}

	// Prompt for tags
	if len(existingTags) > 0 {
		fmt.Printf("(Existing tags: %s)\n", strings.Join(existingTags, ", "))
	}
	tags := promptForTags()
	if len(tags) > 0 {
		for _, todo := range selectedTodos {
			if err := api.AddTodoTags(todo, tags); err != nil {
				fmt.Printf("Error adding tags for %s: %v\n", todo.ID, err)
			}
		}
		if len(selectedTodos) == 1 {
			fmt.Printf("✓ Added tags %s\n", formatTags(tags))
		} else {
			fmt.Printf("✓ Added tags %s to %d tasks\n", formatTags(tags), len(selectedTodos))
		}
	}

	// Prompt for state
	state := promptForState()
	if state != "" {
		for _, todo := range selectedTodos {
			if err := api.SetTodoStatus(todo, state); err != nil {
				fmt.Printf("Error setting state for %s: %v\n", todo.ID, err)
			}
		}
		if len(selectedTodos) == 1 {
			fmt.Printf("✓ Set state to %s\n", state)
		} else {
			fmt.Printf("✓ Set state to %s for %d tasks\n", state, len(selectedTodos))
		}
	}

	fmt.Println("\n✓ Planning complete!")
	return nil
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
	fmt.Print("State (open, in-progress, blocked, done, or Enter to skip): ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return "" // Skip
	}

	validStates := []string{"open", "in-progress", "blocked", "done"}
	for _, s := range validStates {
		if s == input {
			return input
		}
	}

	fmt.Println("Invalid state, skipping")
	return ""
}

// selectMultipleTodos allows selecting multiple todos with FZF
func selectMultipleTodos(todos []api.TodoItem, prompt string) ([]*api.TodoItem, error) {
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

	// Select with FZF (multi-select enabled, Tab selects and moves up)
	selections, err := external.Select(items, external.FZFOptions{
		Header:        prompt,
		Multi:         true,
		Preview:       "",
		PreviewWindow: "",
		ExtraArgs:     []string{"--bind", "tab:toggle+up"},
	})

	if err != nil {
		return nil, err
	}

	// Extract todos from selections
	var selectedTodos []*api.TodoItem
	for _, selected := range selections {
		// Extract ID from selection (first field)
		parts := strings.Fields(selected)
		if len(parts) == 0 {
			continue
		}

		todoID := parts[0]
		if todo, ok := todoMap[todoID]; ok {
			selectedTodos = append(selectedTodos, todo)
		}
	}

	return selectedTodos, nil
}
