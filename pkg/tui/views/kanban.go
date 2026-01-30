package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// KanbanColumn represents a column in the kanban board
type KanbanColumn struct {
	Title  string
	Status string
	Items  []api.TodoItem
}

// KanbanViewModel represents the kanban view state
type KanbanViewModel struct {
	Columns     [4]KanbanColumn
	FocusedCol  int
	SelectedRow int
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
	// Reset all columns
	for i := range k.Columns {
		k.Columns[i].Items = []api.TodoItem{}
	}

	// Group todos by status
	for _, todo := range todos {
		for i := range k.Columns {
			if k.Columns[i].Status == todo.Status {
				k.Columns[i].Items = append(k.Columns[i].Items, todo)
				break
			}
		}
	}

	// Ensure selected row is valid
	if k.FocusedCol >= 0 && k.FocusedCol < len(k.Columns) {
		if k.SelectedRow >= len(k.Columns[k.FocusedCol].Items) {
			k.SelectedRow = len(k.Columns[k.FocusedCol].Items) - 1
		}
		if k.SelectedRow < 0 {
			k.SelectedRow = 0
		}
	}
}

// MoveLeft moves focus to the left column
func (k *KanbanViewModel) MoveLeft() {
	if k.FocusedCol > 0 {
		k.FocusedCol--
		k.SelectedRow = 0
	}
}

// MoveRight moves focus to the right column
func (k *KanbanViewModel) MoveRight() {
	if k.FocusedCol < 3 {
		k.FocusedCol++
		k.SelectedRow = 0
	}
}

// MoveUp moves selection up within the current column
func (k *KanbanViewModel) MoveUp() {
	if k.SelectedRow > 0 {
		k.SelectedRow--
	}
}

// MoveDown moves selection down within the current column
func (k *KanbanViewModel) MoveDown() {
	if k.FocusedCol >= 0 && k.FocusedCol < len(k.Columns) {
		if k.SelectedRow < len(k.Columns[k.FocusedCol].Items)-1 {
			k.SelectedRow++
		}
	}
}

// GetSelectedTodo returns the currently selected todo, if any
func (k *KanbanViewModel) GetSelectedTodo() *api.TodoItem {
	if k.FocusedCol >= 0 && k.FocusedCol < len(k.Columns) {
		col := &k.Columns[k.FocusedCol]
		if k.SelectedRow >= 0 && k.SelectedRow < len(col.Items) {
			return &col.Items[k.SelectedRow]
		}
	}
	return nil
}

// Render renders the kanban board
func (k *KanbanViewModel) Render(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	colWidth := (width - 5) / 4 // 4 columns with spacing

	var renderedCols []string
	for i, col := range k.Columns {
		renderedCols = append(renderedCols, k.renderColumn(i, col, colWidth, height, primaryColor, mutedColor))
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedCols...)
}

// renderColumn renders a single kanban column
func (k *KanbanViewModel) renderColumn(idx int, col KanbanColumn, width, height int, primaryColor, mutedColor lipgloss.Color) string {
	focused := idx == k.FocusedCol

	// Column style
	colStyle := lipgloss.NewStyle().
		Width(width).
		MaxHeight(height).  // Use MaxHeight instead of Height
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1)

	if focused {
		colStyle = colStyle.BorderForeground(primaryColor)
	}

	// Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	header := headerStyle.Render(fmt.Sprintf("%s (%d)", col.Title, len(col.Items)))

	// Calculate available height for cards (account for header, padding, borders)
	availableHeight := height - 6 // Header (1) + empty line (1) + padding (2) + borders (2)

	// Render cards up to available height
	var cards []string
	var totalLines int
	for i, todo := range col.Items {
		if totalLines >= availableHeight {
			break // Don't add more cards if we're out of space
		}
		isSelected := focused && i == k.SelectedRow
		card := k.renderCard(todo, width-4, isSelected, primaryColor, mutedColor)
		cardLines := len(strings.Split(card, "\n"))
		if totalLines+cardLines+1 > availableHeight { // +1 for spacing between cards
			break
		}
		cards = append(cards, card)
		totalLines += cardLines + 1 // +1 for spacing
	}

	// Assemble column content
	content := header
	if len(cards) > 0 {
		content += "\n\n" + strings.Join(cards, "\n\n")
	} else {
		content += "\n\n" + lipgloss.NewStyle().Foreground(mutedColor).Render("(empty)")
	}

	return colStyle.Render(content)
}

// renderCard renders a single todo card
func (k *KanbanViewModel) renderCard(todo api.TodoItem, width int, selected bool, primaryColor, mutedColor lipgloss.Color) string {
	cardStyle := lipgloss.NewStyle().
		Width(width).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(0, 1)

	if selected {
		cardStyle = cardStyle.BorderForeground(primaryColor).Bold(true)
	}

	// Build card content
	var parts []string

	// Priority badge
	if todo.Priority != nil {
		priorityText := fmt.Sprintf("P%d", *todo.Priority)
		var priorityStyle lipgloss.Style
		switch *todo.Priority {
		case 1:
			priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		case 2:
			priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		case 3:
			priorityStyle = lipgloss.NewStyle().Foreground(mutedColor)
		}
		parts = append(parts, priorityStyle.Render(priorityText))
	}

	// Content
	content := todo.Content
	if len(content) > width-4 {
		content = content[:width-7] + "..."
	}
	parts = append(parts, content)

	cardContent := strings.Join(parts, " ")

	// Add project name
	projectStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
	cardContent += "\n" + projectStyle.Render(todo.Project)

	// Add due date if present
	if todo.DueDate != "" {
		dueDateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
		cardContent += "\n" + dueDateStyle.Render("📅 " + todo.DueDate)
	}

	// Add tags if present
	if len(todo.Tags) > 0 {
		tagsStr := strings.Join(todo.Tags, " ")
		if len(tagsStr) > width-4 {
			tagsStr = tagsStr[:width-7] + "..."
		}
		tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
		cardContent += "\n" + tagStyle.Render(tagsStr)
	}

	return cardStyle.Render(cardContent)
}
