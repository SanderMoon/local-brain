package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/external"
	"github.com/spf13/cobra"
)

var (
	todoJSONFlag        bool
	todoAllFlag         bool
	todoPriorityFlag    int
	todoNoPriorityFlag  bool
	todoStatusFlag      string
	todoTagFlag         []string
	todoTagModeFlag     string
	todoDueTodayFlag    bool
	todoDueThisWeekFlag bool
	todoOverdueFlag     bool
	todoSortFlag        string
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
	todoLsCmd.Flags().IntVar(&todoPriorityFlag, "priority", 0, "Filter by priority (1-3)")
	todoLsCmd.Flags().BoolVar(&todoNoPriorityFlag, "no-priority", false, "Show only unprioritized tasks")
	todoLsCmd.Flags().StringVar(&todoStatusFlag, "status", "", "Filter by status (open, in-progress, blocked, done)")
	todoLsCmd.Flags().StringSliceVar(&todoTagFlag, "tag", []string{}, "Filter by tag (can specify multiple)")
	todoLsCmd.Flags().StringVar(&todoTagModeFlag, "tag-mode", "or", "Tag filter mode: 'and' or 'or'")
	todoLsCmd.Flags().BoolVar(&todoDueTodayFlag, "due-today", false, "Show tasks due today")
	todoLsCmd.Flags().BoolVar(&todoDueThisWeekFlag, "due-this-week", false, "Show tasks due this week")
	todoLsCmd.Flags().BoolVar(&todoOverdueFlag, "overdue", false, "Show overdue tasks")
	todoLsCmd.Flags().StringVar(&todoSortFlag, "sort", "", "Sort by: priority, deadline, project, status")
}

// sortTodosByPriority sorts todos with prioritized items first (P1, P2, P3), then unprioritized
func sortTodosByPriority(todos []api.TodoItem) {
	sort.SliceStable(todos, func(i, j int) bool {
		// Both unprioritized - maintain order
		if todos[i].Priority == nil && todos[j].Priority == nil {
			return false
		}
		// i unprioritized, j prioritized - j comes first
		if todos[i].Priority == nil {
			return false
		}
		// i prioritized, j unprioritized - i comes first
		if todos[j].Priority == nil {
			return true
		}
		// Both prioritized - lower number (higher priority) comes first
		return *todos[i].Priority < *todos[j].Priority
	})
}

// sortTodosByPriorityReverse sorts todos with unprioritized items first, then P3, P2, P1
// This is useful for FZF where cursor starts at first item - we want it on unprioritized tasks
// but visually show prioritized items at the top of the display
func sortTodosByPriorityReverse(todos []api.TodoItem) {
	sort.SliceStable(todos, func(i, j int) bool {
		// Both unprioritized - maintain order
		if todos[i].Priority == nil && todos[j].Priority == nil {
			return false
		}
		// i unprioritized, j prioritized - i comes first (reverse)
		if todos[i].Priority == nil {
			return true
		}
		// i prioritized, j unprioritized - j comes first (reverse)
		if todos[j].Priority == nil {
			return false
		}
		// Both prioritized - higher number (lower priority) comes first (reverse)
		return *todos[i].Priority > *todos[j].Priority
	})
}

// formatPriorityBadge returns a colored priority badge for display
func formatPriorityBadge(priority *int) string {
	if priority == nil {
		return "    " // 4 spaces for alignment
	}
	switch *priority {
	case 1:
		return "[P1]"
	case 2:
		return "[P2]"
	case 3:
		return "[P3]"
	default:
		return "    "
	}
}

// formatStatusMark returns the checkbox mark for a task status
func formatStatusMark(status string) string {
	switch status {
	case "open":
		return "[ ]"
	case "in-progress":
		return "[>]"
	case "blocked":
		return "[-]"
	case "done":
		return "[x]"
	default:
		return "[ ]"
	}
}

// filterTodos applies all active filters to the todo list
func filterTodos(todos []api.TodoItem) []api.TodoItem {
	var filtered []api.TodoItem

	for _, todo := range todos {
		// Priority filter
		if todoPriorityFlag > 0 {
			if todo.Priority == nil || *todo.Priority != todoPriorityFlag {
				continue
			}
		}
		if todoNoPriorityFlag {
			if todo.Priority != nil {
				continue
			}
		}

		// Status filter
		if todoStatusFlag != "" {
			if todo.Status != todoStatusFlag {
				continue
			}
		}

		// Tag filter
		if len(todoTagFlag) > 0 {
			if !matchesTags(todo, todoTagFlag, todoTagModeFlag) {
				continue
			}
		}

		// Due date filters
		if todoDueTodayFlag || todoDueThisWeekFlag || todoOverdueFlag {
			if !matchesDueDateFilter(todo) {
				continue
			}
		}

		filtered = append(filtered, todo)
	}

	return filtered
}

// matchesTags checks if a todo matches the tag filter
func matchesTags(todo api.TodoItem, requiredTags []string, mode string) bool {
	if len(requiredTags) == 0 {
		return true
	}

	todoTagsMap := make(map[string]bool)
	for _, tag := range todo.Tags {
		todoTagsMap[strings.ToLower(tag)] = true
	}

	if mode == "and" {
		// All required tags must be present
		for _, reqTag := range requiredTags {
			if !todoTagsMap[strings.ToLower(reqTag)] {
				return false
			}
		}
		return true
	} else {
		// At least one required tag must be present (or mode)
		for _, reqTag := range requiredTags {
			if todoTagsMap[strings.ToLower(reqTag)] {
				return true
			}
		}
		return false
	}
}

// matchesDueDateFilter checks if a todo matches the due date filter
func matchesDueDateFilter(todo api.TodoItem) bool {
	if todo.DueDate == "" {
		return false
	}

	dueDate, err := time.Parse("2006-01-02", todo.DueDate)
	if err != nil {
		return false
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	if todoOverdueFlag {
		return dueDate.Before(today)
	}

	if todoDueTodayFlag {
		return dueDate.Equal(today)
	}

	if todoDueThisWeekFlag {
		endOfWeek := today.AddDate(0, 0, 7)
		return (dueDate.Equal(today) || dueDate.After(today)) && dueDate.Before(endOfWeek)
	}

	return false
}

// sortTodos sorts todos by the specified criterion
func sortTodos(todos []api.TodoItem, sortBy string) {
	switch sortBy {
	case "priority":
		sortTodosByPriority(todos)
	case "deadline":
		sortTodosByDeadline(todos)
	case "project":
		sortTodosByProject(todos)
	case "status":
		sortTodosByStatus(todos)
	default:
		sortTodosByDeadlineAndPriority(todos)
	}
}

// sortTodosByDeadline sorts todos by due date (earliest first)
func sortTodosByDeadline(todos []api.TodoItem) {
	sort.SliceStable(todos, func(i, j int) bool {
		// Tasks with due dates come first
		iHasDue := todos[i].DueDate != ""
		jHasDue := todos[j].DueDate != ""

		if !iHasDue && !jHasDue {
			return false
		}
		if !iHasDue {
			return false
		}
		if !jHasDue {
			return true
		}

		return todos[i].DueDate < todos[j].DueDate
	})
}

// sortTodosByProject sorts todos alphabetically by project name
func sortTodosByProject(todos []api.TodoItem) {
	sort.SliceStable(todos, func(i, j int) bool {
		return todos[i].Project < todos[j].Project
	})
}

// sortTodosByStatus sorts todos by status (in-progress, open, blocked, done)
func sortTodosByStatus(todos []api.TodoItem) {
	statusOrder := map[string]int{
		"in-progress": 1,
		"open":        2,
		"blocked":     3,
		"done":        4,
	}

	sort.SliceStable(todos, func(i, j int) bool {
		return statusOrder[todos[i].Status] < statusOrder[todos[j].Status]
	})
}

// sortTodosByDeadlineAndPriority is the default sort
func sortTodosByDeadlineAndPriority(todos []api.TodoItem) {
	sort.SliceStable(todos, func(i, j int) bool {
		iHasDue := todos[i].DueDate != ""
		jHasDue := todos[j].DueDate != ""

		// Both have no due date - sort by priority
		if !iHasDue && !jHasDue {
			return comparePriority(todos[i].Priority, todos[j].Priority)
		}

		// One has due date, one doesn't - due date comes first
		if !iHasDue {
			return false
		}
		if !jHasDue {
			return true
		}

		// Both have due dates - sort by date, then priority
		iDate, _ := time.Parse("2006-01-02", todos[i].DueDate)
		jDate, _ := time.Parse("2006-01-02", todos[j].DueDate)

		if !iDate.Equal(jDate) {
			return iDate.Before(jDate)
		}

		// Same date - sort by priority
		return comparePriority(todos[i].Priority, todos[j].Priority)
	})
}

// comparePriority compares two priority values (lower number = higher priority comes first)
func comparePriority(a, b *int) bool {
	if a == nil && b == nil {
		return false
	}
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return *a < *b
}

// displayTodos shows todos with enhanced formatting
func displayTodos(todos []api.TodoItem) {
	for _, todo := range todos {
		statusMark := formatStatusMark(todo.Status)
		prioBadge := formatPriorityBadge(todo.Priority)

		// Build display line
		line := fmt.Sprintf("%s %s %s %s", todo.ID, prioBadge, statusMark, todo.Content)

		// Add tags
		if len(todo.Tags) > 0 {
			line += " " + formatTags(todo.Tags)
		}

		// Add project
		line += fmt.Sprintf(" (%s)", todo.Project)

		// Add due date with overdue highlighting
		if todo.DueDate != "" {
			dueDate, err := time.Parse("2006-01-02", todo.DueDate)
			if err == nil {
				now := time.Now()
				today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
				if dueDate.Before(today) {
					line += fmt.Sprintf(" [OVERDUE: %s]", todo.DueDate)
				} else {
					line += fmt.Sprintf(" [Due: %s]", todo.DueDate)
				}
			}
		}

		fmt.Println(line)
	}
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

	// Apply filters
	todos = filterTodos(todos)

	// Apply sorting
	if todoSortFlag != "" {
		sortTodos(todos, todoSortFlag)
	} else {
		// Default sort: deadline first (overdue/upcoming), then priority
		sortTodosByDeadlineAndPriority(todos)
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
		// Enhanced human-readable display
		displayTodos(todos)
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

	// Set status to done
	if err := api.SetTodoStatus(todo, "done"); err != nil {
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

	// Set status to open
	if err := api.SetTodoStatus(todo, "open"); err != nil {
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

	// Sort by priority in reverse (unprioritized first for FZF cursor)
	// This puts unprioritized items at the cursor position (bottom of display)
	// and prioritized items visible at the top
	sortTodosByPriorityReverse(filtered)

	// Format for FZF - show clean display, store metadata for lookup
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
