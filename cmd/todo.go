package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var (
	todoJSONFlag bool
	todoAllFlag  bool
)

var todoCmd = &cobra.Command{
	Use:   "todo",
	Short: "Manage tasks across projects",
	Long: `Manage tasks across all active projects.

Interactive mode (no subcommand):
  Fuzzy search through all open tasks and open selected task in editor

Subcommands:
  ls          List tasks
  done        Mark task as complete
  delete      Delete a task
  reopen      Reopen a completed task`,
	Example: `  brain todo                  # Browse and select from all open tasks
  brain todo ls               # List all open tasks
  brain todo ls --json        # List as JSON with IDs
  brain todo ls --all         # Include completed tasks
  brain todo done abc123      # Mark complete by ID
  brain todo delete abc123    # Delete by ID with confirmation
  brain todo reopen abc123    # Reopen a completed task`,
	RunE: runTodoInteractive,
}

var todoLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List todos",
	Long:  "List all tasks across active projects",
	RunE:  runTodoLs,
}

var todoDoneCmd = &cobra.Command{
	Use:   "done [ID]",
	Short: "Mark task as complete",
	Long: `Mark a task as complete by toggling [ ] to [x].

If no ID is provided, shows interactive selection.`,
	Example: `  brain todo done abc123  # Mark complete by ID
  brain todo done         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTodoDone,
}

var todoDeleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "Delete a task",
	Long: `Delete a task permanently.

If no ID is provided, shows interactive selection.
Requires confirmation before deleting.`,
	Example: `  brain todo delete abc123  # Delete by ID
  brain todo delete         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTodoDelete,
}

var todoReopenCmd = &cobra.Command{
	Use:   "reopen [ID]",
	Short: "Reopen a completed task",
	Long: `Reopen a completed task by toggling [x] back to [ ].

If no ID is provided, shows interactive selection from completed tasks.`,
	Example: `  brain todo reopen abc123  # Reopen by ID
  brain todo reopen         # Interactive selection`,
	Args: cobra.MaximumNArgs(1),
	RunE: runTodoReopen,
}

func init() {
	rootCmd.AddCommand(todoCmd)
	todoCmd.AddCommand(todoLsCmd)
	todoCmd.AddCommand(todoDoneCmd)
	todoCmd.AddCommand(todoDeleteCmd)
	todoCmd.AddCommand(todoReopenCmd)

	todoLsCmd.Flags().BoolVar(&todoJSONFlag, "json", false, "Output JSON format")
	todoLsCmd.Flags().BoolVar(&todoAllFlag, "all", false, "Include completed tasks")
}

func runTodoLs(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")

	todos, err := api.ParseAllTodos(activeDir, todoAllFlag)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}

	if len(todos) == 0 {
		if todoJSONFlag {
			fmt.Println("[]")
		} else {
			fmt.Println("No tasks found")
		}
		return nil
	}

	if todoJSONFlag {
		data, err := json.MarshalIndent(todos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		// Human-readable list
		for _, todo := range todos {
			statusMark := "[ ]"
			if todo.Status == "done" {
				statusMark = "[x]"
			}
			fmt.Printf("%s  %s %s (%s)\n", todo.ID, statusMark, todo.Content, todo.Project)
		}
	}

	return nil
}

func runTodoDone(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection
		todo, err = selectTodo(activeDir, "open", "Select task to complete")
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

	if todo.Status == "done" {
		fmt.Printf("Task is already completed: %s\n", todo.Content)
		return nil
	}

	// Toggle status
	if err := api.ToggleTodoStatus(todo, "done"); err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	fmt.Printf("OK: Completed task: %s (%s)\n", todo.Content, todo.Project)
	return nil
}

func runTodoDelete(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection (include all tasks)
		todo, err = selectTodo(activeDir, "all", "Select task to DELETE")
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

	// Confirmation
	fmt.Printf("About to delete: %s (%s)\n", todo.Content, todo.Project)
	fmt.Print("Are you sure? [y/N] ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Cancelled")
		return nil
	}

	// Delete
	if err := api.DeleteTodoLine(todo); err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	fmt.Printf("OK: Deleted task: %s\n", todo.Content)
	return nil
}

func runTodoReopen(cmd *cobra.Command, args []string) error {
	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	var todo *api.TodoItem

	if len(args) == 0 {
		// Interactive selection from completed tasks
		todo, err = selectTodo(activeDir, "done", "Select completed task to reopen")
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

	// Toggle status
	if err := api.ToggleTodoStatus(todo, "open"); err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	fmt.Printf("OK: Reopened task: %s (%s)\n", todo.Content, todo.Project)
	return nil
}

func runTodoInteractive(cmd *cobra.Command, args []string) error {
	if !external.IsFZFAvailable() {
		return fmt.Errorf("fzf not found (required for interactive mode)")
	}

	activeDir, err := getActiveDir()
	if err != nil {
		return err
	}

	// Get all open tasks
	todos, err := api.ParseAllTodos(activeDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse todos: %w", err)
	}

	if len(todos) == 0 {
		fmt.Println("No open tasks found")
		fmt.Println("")
		fmt.Println("Add tasks to project todo.md files with:")
		fmt.Println("  - [ ] Task description")
		return nil
	}

	// Format for FZF: "FILE:LINE:CONTENT (PROJECT)"
	var items []string
	for _, todo := range todos {
		item := fmt.Sprintf("%s:%d:%s (%s)", todo.File, todo.Line, todo.Content, todo.Project)
		items = append(items, item)
	}

	// Select with FZF and preview
	selected, err := external.SelectOne(items, external.FZFOptions{
		Header:        "Select a task to open in editor (Esc to cancel)",
		Preview:       "bat --color=always --style=numbers --highlight-line {2} {1} 2>/dev/null || cat -n {1}",
		PreviewWindow: "right:60%:+{2}-5",
	})

	if err != nil {
		if err.Error() == "cancelled" {
			return nil
		}
		return err
	}

	// Parse selection: FILE:LINE:...
	parts := strings.SplitN(selected, ":", 3)
	if len(parts) < 2 {
		return fmt.Errorf("invalid selection format")
	}

	filePath := parts[0]
	lineNum := parts[1]

	// Open in editor
	return external.OpenFileAtLineFromString(filePath, lineNum)
}

// Helper functions

func getActiveDir() (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	brainPath, err := cfg.GetCurrentBrainPath()
	if err != nil {
		return "", fmt.Errorf("failed to get brain path: %w", err)
	}

	activeDir := filepath.Join(brainPath, "01_active")

	if _, err := os.Stat(activeDir); os.IsNotExist(err) {
		return "", fmt.Errorf("active projects directory not found: %s", activeDir)
	}

	return activeDir, nil
}

func findTodo(activeDir, query string, includeCompleted bool) (*api.TodoItem, error) {
	todos, err := api.ParseAllTodos(activeDir, includeCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todos: %w", err)
	}

	// Try exact ID match first
	if todo := api.FindTodoByID(todos, query); todo != nil {
		return todo, nil
	}

	// Try pattern match
	matches := api.FindTodoByPattern(todos, query)
	if len(matches) == 1 {
		return &matches[0], nil
	} else if len(matches) > 1 {
		fmt.Println("Error: Multiple matches found. Be more specific or use ID:")
		for _, todo := range matches {
			fmt.Printf("  %s: [%s] %s (%s)\n", todo.ID, todo.Status, todo.Content, todo.Project)
		}
		return nil, fmt.Errorf("ambiguous query")
	}

	return nil, fmt.Errorf("todo not found: %s", query)
}

func selectTodo(activeDir, filter, prompt string) (*api.TodoItem, error) {
	includeCompleted := filter == "all" || filter == "done"

	todos, err := api.ParseAllTodos(activeDir, includeCompleted)
	if err != nil {
		return nil, fmt.Errorf("failed to parse todos: %w", err)
	}

	// Filter by status
	var filtered []api.TodoItem
	for _, todo := range todos {
		if filter == "all" {
			filtered = append(filtered, todo)
		} else if filter == "open" && todo.Status == "open" {
			filtered = append(filtered, todo)
		} else if filter == "done" && todo.Status == "done" {
			filtered = append(filtered, todo)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no matching tasks found")
	}

	// Format for FZF
	var items []string
	for _, todo := range filtered {
		statusMark := "[ ]"
		if todo.Status == "done" {
			statusMark = "[x]"
		}
		item := fmt.Sprintf("%s|%s|%d|%s|%s|%s", todo.ID, todo.File, todo.Line, todo.Status, todo.Content, todo.Project)
		display := fmt.Sprintf("%s  %s %s (%s)", todo.ID, statusMark, todo.Content, todo.Project)
		items = append(items, item+"|||"+display)
	}

	// Select with FZF
	selected, err := external.SelectOne(items, external.FZFOptions{
		Header:        prompt + " (Esc to cancel)",
		Preview:       "bat --color=always --style=numbers --highlight-line $(echo {} | cut -d'|' -f3) $(echo {} | cut -d'|' -f2) 2>/dev/null || cat -n $(echo {} | cut -d'|' -f2)",
		PreviewWindow: "right:60%",
	})

	if err != nil {
		return nil, err
	}

	// Parse selection (take data before "|||")
	dataPart := strings.Split(selected, "|||")[0]
	parts := strings.Split(dataPart, "|")
	if len(parts) < 6 {
		return nil, fmt.Errorf("invalid selection format")
	}

	// Find the todo by ID
	return api.FindTodoByID(filtered, parts[0]), nil
}
