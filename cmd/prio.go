package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var prioCmd = &cobra.Command{
	Use:   "prio [ID] [PRIORITY]",
	Short: "Set or clear task priority",
	Long: `Set or clear the priority of a task using #p:[1-3] tags.

Priority levels:
  1 - High priority
  2 - Medium priority
  3 - Low priority

Interactive mode (no arguments):
  Fuzzy search through all tasks and select one to set priority

Prompt mode (ID only):
  Shows current priority and prompts for new priority

Direct mode (ID and priority):
  Sets the priority directly without prompting`,
	Example: `  brain todo prio              # Interactive selection
  brain todo prio abc123       # Prompt for priority
  brain todo prio abc123 1     # Set to high priority
  brain todo prio abc123 2     # Set to medium priority
  brain todo prio abc123 3     # Set to low priority
  brain todo prio abc123 clear # Remove priority`,
	Args: cobra.MaximumNArgs(2),
	RunE: runPrio,
}

func init() {
	todoCmd.AddCommand(prioCmd)
}

func runPrio(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	// Route based on arguments
	if len(args) == 0 {
		// Interactive mode
		return runPrioInteractive(activeDir)
	} else if len(args) == 1 {
		// Prompt mode
		return runPrioWithPrompt(activeDir, args[0])
	} else {
		// Direct mode
		return runPrioDirect(activeDir, args[0], args[1])
	}
}

func runPrioInteractive(activeDir string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	// Loop until user cancels (Esc in FZF)
	for {
		// Select a todo (only open tasks)
		todo, err := selectTodo(activeDir, "open", "Select task to set priority (Esc to exit)")
		if err != nil {
			// Check if user cancelled (FZF returns error on Esc)
			if err.Error() == "cancelled" || strings.Contains(err.Error(), "no matching tasks") {
				return nil // Exit gracefully
			}
			return err
		}

		// Prompt for priority
		err = promptAndSetPriority(todo)
		if err != nil {
			// If user cancels priority prompt, go back to task selection
			fmt.Println("Priority not set, returning to task selection...")
			continue
		}

		// After setting priority, loop back to show updated list
		fmt.Println("") // Add blank line for readability
	}
}

func runPrioWithPrompt(activeDir, query string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Prompt for priority
	return promptAndSetPriority(todo)
}

func runPrioDirect(activeDir, query, priorityArg string) error {
	// Find the todo
	todo, err := findTodo(activeDir, query, true)
	if err != nil {
		return err
	}

	// Parse priority argument
	var priority *int
	if strings.ToLower(priorityArg) == "clear" || priorityArg == "0" {
		priority = nil
	} else {
		p, err := strconv.Atoi(priorityArg)
		if err != nil || p < 1 || p > 3 {
			return fmt.Errorf("invalid priority: %s (must be 1-3 or 'clear')", priorityArg)
		}
		priority = &p
	}

	// Set priority
	if err := api.SetTodoPriority(todo, priority); err != nil {
		return fmt.Errorf("failed to set priority: %w", err)
	}

	// Show result
	if priority == nil {
		fmt.Printf("OK: Cleared priority for: %s (%s)\n", todo.Content, todo.Project)
	} else {
		priorityName := getPriorityName(*priority)
		fmt.Printf("OK: Set priority to %s for: %s (%s)\n", priorityName, todo.Content, todo.Project)
	}

	return nil
}

func promptAndSetPriority(todo *api.TodoItem) error {
	// Show current priority
	currentPrio := "none"
	if todo.Priority != nil {
		currentPrio = fmt.Sprintf("%d (%s)", *todo.Priority, getPriorityName(*todo.Priority))
	}

	fmt.Printf("Task: %s (%s)\n", todo.Content, todo.Project)
	fmt.Printf("Current priority: %s\n", currentPrio)
	fmt.Print("Enter new priority (1=high, 2=medium, 3=low, 0/clear=none): ")

	// Read input
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		fmt.Println("Cancelled")
		return nil
	}

	// Parse priority
	var priority *int
	if input == "clear" || input == "0" || input == "none" {
		priority = nil
	} else {
		p, err := strconv.Atoi(input)
		if err != nil || p < 1 || p > 3 {
			return fmt.Errorf("invalid priority: %s (must be 1-3 or 0/clear)", input)
		}
		priority = &p
	}

	// Set priority
	if err := api.SetTodoPriority(todo, priority); err != nil {
		return fmt.Errorf("failed to set priority: %w", err)
	}

	// Show result
	if priority == nil {
		fmt.Printf("OK: Cleared priority for: %s\n", todo.Content)
	} else {
		priorityName := getPriorityName(*priority)
		fmt.Printf("OK: Set priority to %s for: %s\n", priorityName, todo.Content)
	}

	return nil
}

func getPriorityName(priority int) string {
	switch priority {
	case 1:
		return "high"
	case 2:
		return "medium"
	case 3:
		return "low"
	default:
		return "unknown"
	}
}
