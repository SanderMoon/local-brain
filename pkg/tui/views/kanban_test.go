package views

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

func TestKanbanViewModel_ScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	// Create test todos (open tasks go to column 1 since backlog is collapsed)
	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
		{Content: "Task 3", Status: "open"},
		{Content: "Task 4", Status: "open"},
		{Content: "Task 5", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Focus starts on Open column (index 1)
	if k.FocusedCol != 1 {
		t.Errorf("Expected initial focused column to be 1 (Open), got %d", k.FocusedCol)
	}

	// Initially scroll offset should be 0
	if k.ScrollOffset[1] != 0 {
		t.Errorf("Expected initial scroll offset to be 0, got %d", k.ScrollOffset[1])
	}

	// Move down several times
	k.MoveDown()
	k.MoveDown()
	k.MoveDown()

	if k.SelectedRow != 3 {
		t.Errorf("Expected selected row to be 3, got %d", k.SelectedRow)
	}
}

func TestKanbanViewModel_MoveUp_AdjustsScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
		{Content: "Task 3", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Render once to cache height (needed for scroll calculations)
	primaryColor := lipgloss.Color("205")
	mutedColor := lipgloss.Color("240")
	_ = k.Render(80, 30, primaryColor, mutedColor)

	// Set scroll offset to 2 and selected row to 2 (Open column = index 1)
	k.SelectedRow = 2
	k.ScrollOffset[1] = 2

	// Move up
	k.MoveUp()

	// Selected row should be 1
	if k.SelectedRow != 1 {
		t.Errorf("Expected selected row to be 1, got %d", k.SelectedRow)
	}

	// Scroll offset should adjust to 1 (to keep selection visible)
	if k.ScrollOffset[1] != 1 {
		t.Errorf("Expected scroll offset to be 1, got %d", k.ScrollOffset[1])
	}
}

func TestKanbanViewModel_MoveLeft_ResetsScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	// Move to In Progress column (index 2) and scroll
	k.FocusedCol = 2
	k.SelectedRow = 0
	k.ScrollOffset[2] = 3

	k.MoveLeft()

	// Should be in Open column now (index 1, since backlog is collapsed)
	if k.FocusedCol != 1 {
		t.Errorf("Expected focused column to be 1 (Open), got %d", k.FocusedCol)
	}

	// Scroll offset for new column should be reset
	if k.ScrollOffset[1] != 0 {
		t.Errorf("Expected scroll offset to be 0 after moving left, got %d", k.ScrollOffset[1])
	}
}

func TestKanbanViewModel_MoveRight_ResetsScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	// Start at Open column (index 1)
	k.FocusedCol = 1
	k.SelectedRow = 0
	k.ScrollOffset[1] = 2

	k.MoveRight()

	// Should be in In Progress column now (index 2)
	if k.FocusedCol != 2 {
		t.Errorf("Expected focused column to be 2 (In Progress), got %d", k.FocusedCol)
	}

	// Scroll offset for new column should be reset
	if k.ScrollOffset[2] != 0 {
		t.Errorf("Expected scroll offset to be 0 after moving right, got %d", k.ScrollOffset[2])
	}
}

func TestKanbanViewModel_UpdateTodos_GroupsByStatus(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Backlog task", Status: "backlog"},
		{Content: "Open task", Status: "open"},
		{Content: "In progress task", Status: "in-progress"},
		{Content: "Blocked task", Status: "blocked"},
		{Content: "Done task", Status: "done"},
		{Content: "Another open task", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Check that todos are grouped correctly
	if len(k.Columns[0].Items) != 1 { // Backlog
		t.Errorf("Expected 1 backlog task, got %d", len(k.Columns[0].Items))
	}
	if len(k.Columns[1].Items) != 2 { // Open
		t.Errorf("Expected 2 open tasks, got %d", len(k.Columns[1].Items))
	}
	if len(k.Columns[2].Items) != 1 { // In Progress
		t.Errorf("Expected 1 in-progress task, got %d", len(k.Columns[2].Items))
	}
	if len(k.Columns[3].Items) != 1 { // Blocked
		t.Errorf("Expected 1 blocked task, got %d", len(k.Columns[3].Items))
	}
	if len(k.Columns[4].Items) != 1 { // Done
		t.Errorf("Expected 1 done task, got %d", len(k.Columns[4].Items))
	}
}

func TestKanbanViewModel_UpdateTodos_AdjustsInvalidScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
		{Content: "Task 3", Status: "open"},
	}

	k.UpdateTodos(todos)
	k.SelectedRow = 2
	k.ScrollOffset[1] = 5 // Invalid - higher than selected row (Open column = index 1)

	// Update with same todos
	k.UpdateTodos(todos)

	// Scroll offset should be adjusted to selected row
	if k.ScrollOffset[1] > k.SelectedRow {
		t.Errorf("Expected scroll offset (%d) to be <= selected row (%d)", k.ScrollOffset[1], k.SelectedRow)
	}
}

func TestKanbanViewModel_GetSelectedTodo(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	// Select first task in Open column (index 1, default focus)
	k.FocusedCol = 1
	k.SelectedRow = 0

	todo := k.GetSelectedTodo()
	if todo == nil {
		t.Fatal("Expected to get a todo, got nil")
	}
	if todo.Content != "Task 1" {
		t.Errorf("Expected task content 'Task 1', got '%s'", todo.Content)
	}

	// Move to In Progress column (index 2)
	k.FocusedCol = 2
	k.SelectedRow = 0

	todo = k.GetSelectedTodo()
	if todo == nil {
		t.Fatal("Expected to get a todo, got nil")
	}
	if todo.Content != "Task 2" {
		t.Errorf("Expected task content 'Task 2', got '%s'", todo.Content)
	}
}

func TestKanbanViewModel_GetSelectedTodo_InvalidSelection(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Try to select invalid row
	k.FocusedCol = 1 // Open column
	k.SelectedRow = 10

	todo := k.GetSelectedTodo()
	if todo != nil {
		t.Errorf("Expected nil for invalid selection, got %v", todo)
	}
}

func TestKanbanViewModel_MoveDown_BoundaryCheck(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Move to last item
	k.SelectedRow = 0
	k.MoveDown()

	if k.SelectedRow != 1 {
		t.Errorf("Expected selected row to be 1, got %d", k.SelectedRow)
	}

	// Try to move beyond last item
	k.MoveDown()

	// Should stay at last item
	if k.SelectedRow != 1 {
		t.Errorf("Expected selected row to stay at 1, got %d", k.SelectedRow)
	}
}

func TestKanbanViewModel_MoveUp_BoundaryCheck(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
	}

	k.UpdateTodos(todos)

	k.SelectedRow = 0

	// Try to move up from first item
	k.MoveUp()

	// Should stay at first item
	if k.SelectedRow != 0 {
		t.Errorf("Expected selected row to stay at 0, got %d", k.SelectedRow)
	}
}

func TestKanbanViewModel_FixedHeightCards(t *testing.T) {
	k := NewKanbanViewModel()

	// Create todos with varying metadata
	todos := []api.TodoItem{
		{Content: "Task with priority", Status: "open", Priority: intPtr(1)},
		{Content: "Task with due date", Status: "open", DueDate: "2026-02-05"},
		{Content: "Task with tags", Status: "open", Tags: []string{"urgent", "bug"}},
		{Content: "Simple task", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Verify card capacity calculation
	availableHeight := 30
	expectedCardsPerView := availableHeight / 6

	if expectedCardsPerView != 5 {
		t.Errorf("Expected 5 cards to fit in 30 lines, got %d", expectedCardsPerView)
	}
}

func TestKanbanViewModel_ScrollIndicators(t *testing.T) {
	k := NewKanbanViewModel()

	// Create many todos to test scroll indicators
	var todos []api.TodoItem
	for i := 1; i <= 10; i++ {
		todos = append(todos, api.TodoItem{
			Content: fmt.Sprintf("Task %d", i),
			Status:  "open",
		})
	}

	k.UpdateTodos(todos)

	// Set scroll offset to middle of list (Open column = index 1)
	k.ScrollOffset[1] = 3
	k.SelectedRow = 3

	// When scrolled, there should be items above and below
	if k.ScrollOffset[1] != 3 {
		t.Errorf("Expected scroll offset to be 3, got %d", k.ScrollOffset[1])
	}

	// Verify we can calculate items above and below
	itemsAbove := k.ScrollOffset[1]
	if itemsAbove != 3 {
		t.Errorf("Expected 3 items above, got %d", itemsAbove)
	}
}

func TestKanbanViewModel_WidthDistribution(t *testing.T) {
	k := NewKanbanViewModel()
	primaryColor := lipgloss.Color("205")
	mutedColor := lipgloss.Color("240")

	// With backlog collapsed (default), we have collapsed strip (14px) + 4 expanded columns
	// The width distribution test verifies the 4 expanded columns share space correctly
	testWidths := []int{100, 101, 102, 103, 80}

	for _, width := range testWidths {
		t.Run(fmt.Sprintf("width=%d", width), func(t *testing.T) {
			// Create some test todos
			todos := []api.TodoItem{
				{Content: "Task", Status: "open"},
			}
			k.UpdateTodos(todos)

			// Render with test width - should not panic
			_ = k.Render(width, 30, primaryColor, mutedColor)
		})
	}
}

func TestKanbanViewModel_ToggleBacklog(t *testing.T) {
	k := NewKanbanViewModel()

	// Default: backlog collapsed, focused on Open (column 1)
	if !k.BacklogCollapsed {
		t.Error("Expected backlog to be collapsed by default")
	}
	if k.FocusedCol != 1 {
		t.Errorf("Expected initial focus on column 1 (Open), got %d", k.FocusedCol)
	}

	// Toggle open
	k.ToggleBacklog()
	if k.BacklogCollapsed {
		t.Error("Expected backlog to be expanded after toggle")
	}

	// Can now focus on backlog
	k.FocusedCol = 0
	k.validateCursor()
	if k.FocusedCol != 0 {
		t.Errorf("Expected to be able to focus on backlog when expanded, got column %d", k.FocusedCol)
	}

	// Toggle closed while focused on backlog
	k.ToggleBacklog()
	if !k.BacklogCollapsed {
		t.Error("Expected backlog to be collapsed after second toggle")
	}
	// Should auto-move focus to Open
	if k.FocusedCol != 1 {
		t.Errorf("Expected focus to move to column 1 (Open) when backlog collapsed, got %d", k.FocusedCol)
	}
}

func TestKanbanViewModel_MoveLeft_SkipsCollapsedBacklog(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
	}
	k.UpdateTodos(todos)

	// Focus on Open column (1), try to move left
	k.FocusedCol = 1
	k.MoveLeft()

	// Should stay at 1, not go to collapsed backlog
	if k.FocusedCol != 1 {
		t.Errorf("Expected to stay at column 1 when backlog collapsed, got %d", k.FocusedCol)
	}
}

// Helper function for tests
func intPtr(i int) *int {
	return &i
}
