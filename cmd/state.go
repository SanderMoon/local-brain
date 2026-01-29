package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [ID] [STATUS]",
	Short: "Set task status",
	Long: `Set the status of a task.

Valid statuses:
  open        - Open task (default)
  in-progress - Currently working on
  blocked     - Waiting on something
  done        - Completed

Interactive mode (no arguments):
  Fuzzy search through all tasks and select one to set status

Prompt mode (ID only):
  Shows current status and prompts for new status

Direct mode (ID and status):
  Sets the status directly without prompting`,
	Example: `  brain todo status                    # Interactive selection
  brain todo status abc123             # Prompt for status
  brain todo status abc123 in-progress # Set to in-progress
  brain todo status abc123 blocked     # Set to blocked
  brain todo status abc123 open        # Set to open`,
	Args: cobra.MaximumNArgs(2),
	RunE: runStatus,
}

var startCmd = &cobra.Command{
	Use:   "start [ID]",
	Short: "Mark task as in-progress",
	Long: `Mark a task as in-progress by changing checkbox to [>].

If no ID is provided, shows interactive selection.`,
	Example: `  brain todo start abc123  # Start by ID
  brain todo start         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStart,
}

var blockCmd = &cobra.Command{
	Use:   "block [ID]",
	Short: "Mark task as blocked",
	Long: `Mark a task as blocked by changing checkbox to [-].

If no ID is provided, shows interactive selection.`,
	Example: `  brain todo block abc123  # Block by ID
  brain todo block         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBlock,
}

var unblockCmd = &cobra.Command{
	Use:   "unblock [ID]",
	Short: "Unblock a task (alias for reopen)",
	Long: `Unblock a task by changing checkbox back to [ ].

This is an alias for 'brain todo reopen'.
If no ID is provided, shows interactive selection.`,
	Example: `  brain todo unblock abc123  # Unblock by ID
  brain todo unblock         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUnblock,
}

func init() {
	todoCmd.AddCommand(statusCmd)
	todoCmd.AddCommand(startCmd)
	todoCmd.AddCommand(blockCmd)
	todoCmd.AddCommand(unblockCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	// Route based on arguments
	if len(args) == 0 {
		// Interactive mode
		return runStatusInteractive(activeDir)
	} else if len(args) == 1 {
		// Prompt mode
		return runStatusWithPrompt(activeDir, args[0])
	} else {
		// Direct mode
		return runStatusDirect(activeDir, args[0], args[1])
	}
}

func runStatusInteractive(activeDir string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Select a todo (all non-completed tasks)
		todo, err := selectTodoByStatus(activeDir, []string{"open", "in-progress", "blocked"}, "Select task to set status (Esc to exit)")
		if err != nil {
			// Check if user cancelled (FZF returns error on Esc)
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Prompt for status
		err = promptAndSetStatus(todo)
		if err != nil {
			// If user cancels status prompt, go back to task selection
			fmt.Println("Status not set, returning to task selection...")
			continue
		}

		// After setting status, loop back to show updated list
		fmt.Println("") // Add blank line for readability
	}
}

func runStatusWithPrompt(activeDir, query string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Prompt for status
	return promptAndSetStatus(todo)
}

func runStatusDirect(activeDir, query, statusArg string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Validate status
	validStatuses := []string{"open", "in-progress", "blocked", "done"}
	statusArg = strings.ToLower(statusArg)
	found := false
	for _, s := range validStatuses {
		if s == statusArg {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid status: %s (must be: open, in-progress, blocked, done)", statusArg)
	}

	// Set status
	if err := api.SetTodoStatus(todo, statusArg); err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}

	// Show result
	fmt.Printf("OK: Set status to %s for: %s (%s)\n", statusArg, todo.Content, todo.Project)

	return nil
}

func runStart(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection
		todo, err = selectTodoByStatus(activeDir, []string{"open", "blocked"}, "Select task to start")
		if err != nil {
			return err
		}
	} else {
		// Find by ID
		todo, err = findTodo(activeDir, args[0], false)
		if err != nil {
			return err
		}
	}

	if todo.Status == "in-progress" {
		fmt.Printf("Task is already in-progress: %s\n", todo.Content)
		return nil
	}

	// Set status to in-progress
	if err := api.SetTodoStatus(todo, "in-progress"); err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	fmt.Printf("OK: Started task: %s (%s)\n", todo.Content, todo.Project)
	return nil
}

func runBlock(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection
		todo, err = selectTodoByStatus(activeDir, []string{"open", "in-progress"}, "Select task to block")
		if err != nil {
			return err
		}
	} else {
		// Find by ID
		todo, err = findTodo(activeDir, args[0], false)
		if err != nil {
			return err
		}
	}

	if todo.Status == "blocked" {
		fmt.Printf("Task is already blocked: %s\n", todo.Content)
		return nil
	}

	// Set status to blocked
	if err := api.SetTodoStatus(todo, "blocked"); err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	fmt.Printf("OK: Blocked task: %s (%s)\n", todo.Content, todo.Project)
	return nil
}

func runUnblock(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection from blocked tasks
		todo, err = selectTodoByStatus(activeDir, []string{"blocked"}, "Select task to unblock")
		if err != nil {
			return err
		}
	} else {
		// Find by ID
		todo, err = findTodo(activeDir, args[0], true)
		if err != nil {
			return err
		}
	}

	if todo.Status == "open" {
		fmt.Printf("Task is already open: %s\n", todo.Content)
		return nil
	}

	// Set status to open
	if err := api.SetTodoStatus(todo, "open"); err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	fmt.Printf("OK: Unblocked task: %s (%s)\n", todo.Content, todo.Project)
	return nil
}

func promptAndSetStatus(todo *api.TodoItem) error {
	// Show current status
	fmt.Printf("Task: %s (%s)\n", todo.Content, todo.Project)
	fmt.Printf("Current status: %s\n", todo.Status)
	fmt.Print("Enter new status (open/in-progress/blocked/done): ")

	// Read input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		fmt.Println("Cancelled")
		return fmt.Errorf("cancelled")
	}

	// Validate status
	validStatuses := []string{"open", "in-progress", "blocked", "done"}
	found := false
	for _, s := range validStatuses {
		if s == input {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("invalid status: %s (must be: open, in-progress, blocked, done)", input)
	}

	// Set status
	if err := api.SetTodoStatus(todo, input); err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}

	// Show result
	fmt.Printf("OK: Set status to %s for: %s\n", input, todo.Content)

	return nil
}

// selectTodoByStatus selects a todo filtered by specific statuses
func selectTodoByStatus(activeDir string, statuses []string, prompt string) (*api.TodoItem, error) {
	includeCompleted := false
	for _, s := range statuses {
		if s == "done" {
			includeCompleted = true
			break
		}
	}

	todos, err := api.ParseAllTodos(activeDir, includeCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todos: %w", err)
	}

	// Filter by status
	var filtered []api.TodoItem
	for _, todo := range todos {
		for _, s := range statuses {
			if todo.Status == s {
				filtered = append(filtered, todo)
				break
			}
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no matching tasks found")
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
