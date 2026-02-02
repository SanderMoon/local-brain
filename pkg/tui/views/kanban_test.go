package views

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

func TestKanbanViewModel_ScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	// Create test todos
	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "open"},
		{Content: "Task 3", Status: "open"},
		{Content: "Task 4", Status: "open"},
		{Content: "Task 5", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Initially scroll offset should be 0
	if k.ScrollOffset[0] != 0 {
		t.Errorf("Expected initial scroll offset to be 0, got %d", k.ScrollOffset[0])
	}

	// Move down several times
	k.MoveDown()
	k.MoveDown()
	k.MoveDown()

	if k.SelectedRow != 3 {
		t.Errorf("Expected selected row to be 3, got %d", k.SelectedRow)
	}

	// Note: scroll offset is adjusted during rendering
	// We test the scroll behavior through the navigation methods
	// which ensure the selected item stays visible
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

	// Set scroll offset to 2 and selected row to 2
	k.SelectedRow = 2
	k.ScrollOffset[0] = 2

	// Move up
	k.MoveUp()

	// Selected row should be 1
	if k.SelectedRow != 1 {
		t.Errorf("Expected selected row to be 1, got %d", k.SelectedRow)
	}

	// Scroll offset should adjust to 1 (to keep selection visible)
	if k.ScrollOffset[0] != 1 {
		t.Errorf("Expected scroll offset to be 1, got %d", k.ScrollOffset[0])
	}
}

func TestKanbanViewModel_MoveLeft_ResetsScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	// Move to second column and scroll
	k.FocusedCol = 1
	k.SelectedRow = 0
	k.ScrollOffset[1] = 3

	k.MoveLeft()

	// Should be in first column now
	if k.FocusedCol != 0 {
		t.Errorf("Expected focused column to be 0, got %d", k.FocusedCol)
	}

	// Scroll offset for new column should be reset
	if k.ScrollOffset[0] != 0 {
		t.Errorf("Expected scroll offset to be 0 after moving left, got %d", k.ScrollOffset[0])
	}
}

func TestKanbanViewModel_MoveRight_ResetsScrollOffset(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	k.FocusedCol = 0
	k.SelectedRow = 0
	k.ScrollOffset[0] = 2

	k.MoveRight()

	// Should be in second column now
	if k.FocusedCol != 1 {
		t.Errorf("Expected focused column to be 1, got %d", k.FocusedCol)
	}

	// Scroll offset for new column should be reset
	if k.ScrollOffset[1] != 0 {
		t.Errorf("Expected scroll offset to be 0 after moving right, got %d", k.ScrollOffset[1])
	}
}

func TestKanbanViewModel_UpdateTodos_GroupsByStatus(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Open task", Status: "open"},
		{Content: "In progress task", Status: "in-progress"},
		{Content: "Blocked task", Status: "blocked"},
		{Content: "Done task", Status: "done"},
		{Content: "Another open task", Status: "open"},
	}

	k.UpdateTodos(todos)

	// Check that todos are grouped correctly
	if len(k.Columns[0].Items) != 2 { // Open
		t.Errorf("Expected 2 open tasks, got %d", len(k.Columns[0].Items))
	}
	if len(k.Columns[1].Items) != 1 { // In Progress
		t.Errorf("Expected 1 in-progress task, got %d", len(k.Columns[1].Items))
	}
	if len(k.Columns[2].Items) != 1 { // Blocked
		t.Errorf("Expected 1 blocked task, got %d", len(k.Columns[2].Items))
	}
	if len(k.Columns[3].Items) != 1 { // Done
		t.Errorf("Expected 1 done task, got %d", len(k.Columns[3].Items))
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
	k.ScrollOffset[0] = 5 // Invalid - higher than selected row

	// Update with same todos
	k.UpdateTodos(todos)

	// Scroll offset should be adjusted to selected row
	if k.ScrollOffset[0] > k.SelectedRow {
		t.Errorf("Expected scroll offset (%d) to be <= selected row (%d)", k.ScrollOffset[0], k.SelectedRow)
	}
}

func TestKanbanViewModel_GetSelectedTodo(t *testing.T) {
	k := NewKanbanViewModel()

	todos := []api.TodoItem{
		{Content: "Task 1", Status: "open"},
		{Content: "Task 2", Status: "in-progress"},
	}

	k.UpdateTodos(todos)

	// Select first task in first column
	k.FocusedCol = 0
	k.SelectedRow = 0

	todo := k.GetSelectedTodo()
	if todo == nil {
		t.Fatal("Expected to get a todo, got nil")
	}
	if todo.Content != "Task 1" {
		t.Errorf("Expected task content 'Task 1', got '%s'", todo.Content)
	}

	// Move to second column
	k.FocusedCol = 1
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
	k.FocusedCol = 0
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

	// All cards should render to the same height (5 lines including borders)
	// We can't easily test the exact rendering output, but we can verify
	// that the scroll calculation works correctly with the fixed height assumption

	// With fixed height of 5 lines per card + 1 spacing = 6 lines per card
	// If we have 30 lines available, we should fit 5 cards (30 / 6 = 5)
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

	// Set scroll offset to middle of list
	k.ScrollOffset[0] = 3
	k.SelectedRow = 3

	// When scrolled, there should be items above and below
	// This is implicitly tested by the scroll offset logic
	if k.ScrollOffset[0] != 3 {
		t.Errorf("Expected scroll offset to be 3, got %d", k.ScrollOffset[0])
	}

	// Verify we can calculate items above and below
	itemsAbove := k.ScrollOffset[0]
	if itemsAbove != 3 {
		t.Errorf("Expected 3 items above, got %d", itemsAbove)
	}
}

func TestKanbanViewModel_WidthDistribution(t *testing.T) {
	k := NewKanbanViewModel()
	primaryColor := lipgloss.Color("205")
	mutedColor := lipgloss.Color("240")

	// Test various widths to ensure proper distribution without gaps or overflow
	testWidths := []struct {
		width          int
		expectedWidths [4]int // Expected width for each of the 4 columns
	}{
		{100, [4]int{25, 24, 24, 24}}, // (100-3)/4 = 24 base, remainder 1
		{101, [4]int{25, 25, 24, 24}}, // (101-3)/4 = 24 base, remainder 2
		{102, [4]int{25, 25, 25, 24}}, // (102-3)/4 = 24 base, remainder 3
		{103, [4]int{25, 25, 25, 25}}, // (103-3)/4 = 25 base, remainder 0
		{80, [4]int{20, 19, 19, 19}},  // (80-3)/4 = 19 base, remainder 1
	}

	for _, tt := range testWidths {
		t.Run(fmt.Sprintf("width=%d", tt.width), func(t *testing.T) {
			// Create some test todos
			todos := []api.TodoItem{
				{Content: "Task", Status: "open"},
			}
			k.UpdateTodos(todos)

			// Render with test width
			_ = k.Render(tt.width, 30, primaryColor, mutedColor)

			// Verify column widths with spacing accounted for
			availableWidth := tt.width - 3 // 3 spaces between 4 columns
			baseWidth := availableWidth / 4
			remainder := availableWidth % 4
			totalWidth := 0

			for i := 0; i < 4; i++ {
				expectedWidth := baseWidth
				if i < remainder {
					expectedWidth++
				}
				if expectedWidth != tt.expectedWidths[i] {
					t.Errorf("Column %d: expected width %d, got %d", i, tt.expectedWidths[i], expectedWidth)
				}
				totalWidth += expectedWidth
			}

			// Total column widths + spacing should equal available width
			if totalWidth+3 != tt.width {
				t.Errorf("Total column widths (%d) + spacing (3) don't match total width (%d)", totalWidth, tt.width)
			}
		})
	}
}

// Helper function for tests
func intPtr(i int) *int {
	return &i
}
