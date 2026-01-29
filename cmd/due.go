package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/dateutil"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	Use:   "due [ID] [DATE]",
	Short: "Set or clear task due date",
	Long: `Set or clear the due date of a task using #due:YYYY-MM-DD tags.

Date formats supported:
  ISO date: 2026-02-15
  Keywords: today, tomorrow, yesterday
  Relative: +3d, -2w, +1m, +1y
  Day names: monday, next-friday, this-saturday
  Clear: clear (removes due date)

Interactive mode (no arguments):
  Fuzzy search through all tasks and select one to set due date

Prompt mode (ID only):
  Shows current due date and prompts for new due date

Direct mode (ID and date):
  Sets the due date directly without prompting`,
	Example: `  brain todo due                    # Interactive selection
  brain todo due abc123             # Prompt for due date
  brain todo due abc123 2026-02-15  # Set to specific date
  brain todo due abc123 tomorrow    # Set to tomorrow
  brain todo due abc123 +3d         # Set to 3 days from now
  brain todo due abc123 next-friday # Set to next Friday
  brain todo due abc123 clear       # Remove due date`,
	Args: cobra.MaximumNArgs(2),
	RunE: runDue,
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Batch schedule tasks interactively",
	Long: `Interactive batch mode for setting due dates on multiple tasks.

Loops through tasks with FZF selection, prompting for due dates.
Supports all natural language date formats.`,
	Example: `  brain todo schedule  # Interactive batch scheduling`,
	RunE:    runSchedule,
}

func init() {
	todoCmd.AddCommand(dueCmd)
	todoCmd.AddCommand(scheduleCmd)
}

func runDue(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	// Route based on arguments
	if len(args) == 0 {
		// Interactive mode
		return runDueInteractive(activeDir)
	} else if len(args) == 1 {
		// Prompt mode
		return runDueWithPrompt(activeDir, args[0])
	} else {
		// Direct mode
		return runDueDirect(activeDir, args[0], args[1])
	}
}

func runDueInteractive(activeDir string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Select a todo (only open tasks)
		todo, err := selectTodo(activeDir, "open", "Select task to set due date (Esc to exit)")
		if err != nil {
			// Check if user cancelled (FZF returns error on Esc)
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Prompt for due date
		err = promptAndSetDueDate(todo)
		if err != nil {
			// If user cancels due date prompt, go back to task selection
			fmt.Println("Due date not set, returning to task selection...")
			continue
		}

		// After setting due date, loop back to show updated list
		fmt.Println("") // Add blank line for readability
	}
}

func runDueWithPrompt(activeDir, query string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Prompt for due date
	return promptAndSetDueDate(todo)
}

func runDueDirect(activeDir, query, dateArg string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Parse date (supports natural language)
	var dueDate string
	if strings.ToLower(dateArg) == "clear" {
		dueDate = "clear"
	} else {
		parsed, err := dateutil.ParseNaturalDate(dateArg)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (%v)", dateArg, err)
		}
		dueDate = parsed
	}

	// Set due date
	if err := api.SetTodoDueDate(todo, dueDate); err != nil {
		return fmt.Errorf("failed to set due date: %w", err)
	}

	// Show result
	if dueDate == "" || dueDate == "clear" {
		fmt.Printf("OK: Cleared due date for: %s (%s)\n", todo.Content, todo.Project)
	} else {
		fmt.Printf("OK: Set due date to %s for: %s (%s)\n", dueDate, todo.Content, todo.Project)
	}

	return nil
}

func runSchedule(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Select a todo (only open tasks without due dates)
		todo, err := selectTodoNoDueDate(activeDir, "Select task to schedule (Esc to exit)")
		if err != nil {
			// Check if user cancelled (FZF returns error on Esc)
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Prompt for due date
		err = promptAndSetDueDate(todo)
		if err != nil {
			// If user cancels due date prompt, go back to task selection
			fmt.Println("Due date not set, returning to task selection...")
			continue
		}

		// After setting due date, loop back to show updated list
		fmt.Println("") // Add blank line for readability
	}
}

func promptAndSetDueDate(todo *api.TodoItem) error {
	// Show current due date
	currentDue := "none"
	if todo.DueDate != "" {
		currentDue = todo.DueDate
	}

	fmt.Printf("Task: %s (%s)\n", todo.Content, todo.Project)
	fmt.Printf("Current due date: %s\n", currentDue)
	fmt.Print("Enter due date (YYYY-MM-DD, tomorrow, +3d, next-friday, or clear): ")

	// Read input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("Cancelled")
		return fmt.Errorf("cancelled")
	}

	// Parse date (supports natural language)
	var dueDate string
	if strings.ToLower(input) == "clear" || strings.ToLower(input) == "none" {
		dueDate = "clear"
	} else {
		parsed, err := dateutil.ParseNaturalDate(input)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (%v)", input, err)
		}
		dueDate = parsed
	}

	// Set due date
	if err := api.SetTodoDueDate(todo, dueDate); err != nil {
		return fmt.Errorf("failed to set due date: %w", err)
	}

	// Show result
	if dueDate == "" || dueDate == "clear" {
		fmt.Printf("OK: Cleared due date for: %s\n", todo.Content)
	} else {
		fmt.Printf("OK: Set due date to %s for: %s\n", dueDate, todo.Content)
	}

	return nil
}

// selectTodoNoDueDate selects a todo without a due date
func selectTodoNoDueDate(activeDir string, prompt string) (*api.TodoItem, error) {
	todos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todos: %w", err)
	}

	// Filter for open tasks without due dates
	var filtered []api.TodoItem
	for _, todo := range todos {
		if todo.Status == "open" && todo.DueDate == "" {
			filtered = append(filtered, todo)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no tasks without due dates found")
	}

	// Sort by priority in reverse (unprioritized first for FZF cursor)
	sortTodosByPriorityReverse(filtered)

	// Format for FZF
	var items []string
	todoMap := make(map[string]*api.TodoItem)
	for i := range filtered {
		todo := &filtered[i]
		statusMark := formatStatusMark(todo.Status)
		prioBadge := formatPriorityBadge(todo.Priority)
		display := fmt.Sprintf("%s %s %s %s (%s)", todo.ID, prioBadge, statusMark, todo.Content, todo.Project)
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
