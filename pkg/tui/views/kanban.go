package views

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// Geometry Constants
const (
	// CardHeight: 3 lines of content + 2 lines of border = 5 lines total
	CardHeight = 5
	// CardSpacing: 1 line of space between cards
	CardSpacing = 1
	// CardStride: Total vertical space a single card occupies (used for calculation)
	CardStride = CardHeight + CardSpacing
)

// KanbanColumn represents a column in the kanban board
type KanbanColumn struct {
	Title  string
	Status string
	Items  []api.TodoItem
}

// KanbanViewModel represents the kanban view state
type KanbanViewModel struct {
	Columns      [4]KanbanColumn
	FocusedCol   int
	SelectedRow  int
	ScrollOffset [4]int // Scroll offset for each column

	// Cache dimensions to prevent scrolling glitches
	lastWidth  int
	lastHeight int
}

// NewKanbanViewModel creates a new kanban view model
func NewKanbanViewModel() KanbanViewModel {
	return KanbanViewModel{
		Columns: [4]KanbanColumn{
			{Title: "Open", Status: "open"},
			{Title: "In Progress", Status: "in-progress"},
			{Title: "Blocked", Status: "blocked"},
			{Title: "Done", Status: "done"},
		},
		FocusedCol:  0,
		SelectedRow: 0,
	}
}

// UpdateTodos updates the kanban board with new todos
func (k *KanbanViewModel) UpdateTodos(todos []api.TodoItem) {
	// 1. Reset all columns
	for i := range k.Columns {
		k.Columns[i].Items = []api.TodoItem{}
	}

	// 2. Group todos by status
	for _, todo := range todos {
		for i := range k.Columns {
			if k.Columns[i].Status == todo.Status {
				k.Columns[i].Items = append(k.Columns[i].Items, todo)
				break
			}
		}
	}

	// 3. Validate cursor and scroll positions against new data
	k.validateCursor()
}

// MoveLeft moves focus to the left column
func (k *KanbanViewModel) MoveLeft() {
	if k.FocusedCol > 0 {
		k.FocusedCol--
		k.SelectedRow = 0 // Reset selection to top when switching columns
		k.validateCursor()
	}
}

// MoveRight moves focus to the right column
func (k *KanbanViewModel) MoveRight() {
	if k.FocusedCol < len(k.Columns)-1 {
		k.FocusedCol++
		k.SelectedRow = 0
		k.validateCursor()
	}
}

// MoveUp moves selection up within the current column
func (k *KanbanViewModel) MoveUp() {
	if k.SelectedRow > 0 {
		k.SelectedRow--
		k.validateCursor()
	}
}

// MoveDown moves selection down within the current column
func (k *KanbanViewModel) MoveDown() {
	colLen := len(k.Columns[k.FocusedCol].Items)
	if k.SelectedRow < colLen-1 {
		k.SelectedRow++
		k.validateCursor()
	}
}

// GetSelectedTodo returns the currently selected todo, if any
func (k *KanbanViewModel) GetSelectedTodo() *api.TodoItem {
	if k.FocusedCol >= 0 && k.FocusedCol < len(k.Columns) {
		col := &k.Columns[k.FocusedCol]
		if len(col.Items) > 0 && k.SelectedRow >= 0 && k.SelectedRow < len(col.Items) {
			return &col.Items[k.SelectedRow]
		}
	}
	return nil
}

// cardCapacity calculates how many cards fit in the given screen height safely.
func (k *KanbanViewModel) cardCapacity(screenHeight int) int {
	// Safety Margin: -2 for Border, -2 for Header, -1 buffer to prevent overflow
	availableHeight := screenHeight - 5
	
	if availableHeight < CardHeight {
		return 0
	}
	
	count := availableHeight / CardStride
	if count < 1 {
		return 1 // Always try to show at least one
	}
	return count
}

// validateCursor ensures the selected item and scroll offset are valid
// This acts as the single source of truth for scrolling logic.
func (k *KanbanViewModel) validateCursor() {
	// 1. Bounds check column
	if k.FocusedCol < 0 { k.FocusedCol = 0 }
	if k.FocusedCol >= len(k.Columns) { k.FocusedCol = len(k.Columns) - 1 }

	col := &k.Columns[k.FocusedCol]
	nItems := len(col.Items)

	// 2. Handle empty columns
	if nItems == 0 {
		k.SelectedRow = 0
		k.ScrollOffset[k.FocusedCol] = 0
		return
	}

	// 3. Clamp SelectedRow
	if k.SelectedRow >= nItems {
		k.SelectedRow = nItems - 1
	}
	if k.SelectedRow < 0 {
		k.SelectedRow = 0
	}

	// 4. Update Scroll Offset
	// We use the cached lastHeight to ensure consistent behavior with the last render
	if k.lastHeight > 0 {
		maxVisible := k.cardCapacity(k.lastHeight)

		// Scroll Up if selected is above viewport
		if k.SelectedRow < k.ScrollOffset[k.FocusedCol] {
			k.ScrollOffset[k.FocusedCol] = k.SelectedRow
		}
		
		// Scroll Down if selected is below viewport
		// If index 5 is selected, and we can see 3 items...
		// 5 >= 0 + 3? Yes. Offset = 5 - 3 + 1 = 3. Visible: 3, 4, 5.
		if k.SelectedRow >= k.ScrollOffset[k.FocusedCol] + maxVisible {
			k.ScrollOffset[k.FocusedCol] = k.SelectedRow - maxVisible + 1
		}
	}

	// 5. Final safety clamp on scroll offset
	if k.ScrollOffset[k.FocusedCol] > nItems-1 {
		k.ScrollOffset[k.FocusedCol] = nItems - 1
	}
	if k.ScrollOffset[k.FocusedCol] < 0 {
		k.ScrollOffset[k.FocusedCol] = 0
	}
}

// Render renders the kanban board
func (k *KanbanViewModel) Render(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	// Update dimensions cache for the next logic update
	k.lastWidth = width
	k.lastHeight = height

	// Ensure cursor is valid before drawing
	k.validateCursor()

	// Calculate column widths accounting for spacing between columns
	// Reserve space for gaps (1 char between each of 4 columns = 3 chars)
	availableWidth := width - 3
	baseWidth := availableWidth / 4
	remainder := availableWidth % 4

	var renderedCols []string
	for i := range k.Columns {
		// Distribute remainder pixels to first N columns
		colWidth := baseWidth
		if i < remainder {
			colWidth++
		}

		renderedCols = append(renderedCols, k.renderColumn(i, colWidth, height, primaryColor, mutedColor))

		// Add spacing between columns (except after last column)
		if i < 3 {
			renderedCols = append(renderedCols, " ")
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}

// renderColumn renders a single kanban column
func (k *KanbanViewModel) renderColumn(colIdx int, width, height int, primaryColor, mutedColor lipgloss.Color) string {
	col := k.Columns[colIdx]
	isFocused := k.FocusedCol == colIdx

	borderColor := mutedColor
	if isFocused {
		borderColor = primaryColor
	}

	// 1. Calculate Capacity
	maxCards := k.cardCapacity(height)

	// 2. Determine visible range
	startIdx := k.ScrollOffset[colIdx]
	endIdx := startIdx + maxCards
	if endIdx > len(col.Items) {
		endIdx = len(col.Items)
	}

	// 3. Render Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(primaryColor).
		Padding(0, 1).
		Width(width - 2) // Account for borders

	header := headerStyle.Render(fmt.Sprintf("%s (%d)", col.Title, len(col.Items)))
	
	// 4. Render Body content
	var bodyParts []string
	
	// "More Above" marker
	if startIdx > 0 {
		bodyParts = append(bodyParts, lipgloss.NewStyle().Foreground(mutedColor).Align(lipgloss.Center).Width(width-4).Render("↑"))
	} else {
		bodyParts = append(bodyParts, "") // Empty string to maintain alignment
	}

	// Cards
	if len(col.Items) == 0 {
		bodyParts = append(bodyParts, lipgloss.NewStyle().Foreground(mutedColor).Padding(1).Render("(empty)"))
	} else {
		for i := startIdx; i < endIdx; i++ {
			isSelected := isFocused && i == k.SelectedRow
			bodyParts = append(bodyParts, k.renderCard(col.Items[i], width-4, isSelected, primaryColor, mutedColor))
		}
	}
	
	// "More Below" marker
	if endIdx < len(col.Items) {
		bodyParts = append(bodyParts, lipgloss.NewStyle().Foreground(mutedColor).Align(lipgloss.Center).Width(width-4).Render("↓"))
	}

	content := lipgloss.JoinVertical(lipgloss.Left, bodyParts...)
	
	// 5. Combine everything
	fullContent := lipgloss.JoinVertical(lipgloss.Left, header, "\n", content)

	// 6. Strict container styling to prevent overflow
	baseStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		MaxHeight(height - 2). // Prevents terminal scroll push
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor)

	return baseStyle.Render(fullContent)
}

// renderCard renders a single todo card with fixed height
func (k *KanbanViewModel) renderCard(todo api.TodoItem, width int, selected bool, primaryColor, mutedColor lipgloss.Color) string {
	borderColor := mutedColor
	if selected {
		borderColor = primaryColor
	}

	cardStyle := lipgloss.NewStyle().
		Width(width).
		Height(CardHeight - 2). // -2 for borders
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	if selected {
		cardStyle = cardStyle.Bold(true)
	}

	// Line 1: Priority + Content
	var titleBuilder strings.Builder
	
	if todo.Priority != nil {
		pColor := mutedColor
		switch *todo.Priority {
		case 1: pColor = lipgloss.Color("196") // Red
		case 2: pColor = lipgloss.Color("208") // Orange
		}
		titleBuilder.WriteString(lipgloss.NewStyle().Foreground(pColor).Bold(true).Render(fmt.Sprintf("P%d ", *todo.Priority)))
	}

	// Truncate content safely
	availWidth := width - 7 // rough estimate of used space
	content := todo.Content
	if len(content) > availWidth {
		safeLen := int(math.Max(0, float64(availWidth-3)))
		content = content[:safeLen] + "..."
	}
	titleBuilder.WriteString(content)

	// Line 2: Project
	project := lipgloss.NewStyle().Foreground(mutedColor).Italic(true).Render(todo.Project)

	// Line 3: Meta
	var meta string
	if todo.DueDate != "" {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("📅 " + todo.DueDate)
	} else if len(todo.Tags) > 0 {
		meta = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render(strings.Join(todo.Tags, " "))
	}

	// Force exactly 3 lines using JoinVertical
	innerContent := lipgloss.JoinVertical(lipgloss.Left, 
		titleBuilder.String(),
		project,
		meta,
	)

	return cardStyle.Render(innerContent)
}
