package views

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// DumpViewModel represents the dump inbox view state
type DumpViewModel struct {
	Items       []api.DumpItemJSON
	SelectedIdx int
}

// NewDumpViewModel creates a new dump view model
func NewDumpViewModel() DumpViewModel {
	return DumpViewModel{
		Items:       []api.DumpItemJSON{},
		SelectedIdx: 0,
	}
}

// UpdateItems updates the dump items list
func (d *DumpViewModel) UpdateItems(items []api.DumpItemJSON) {
	d.Items = items
	if d.SelectedIdx >= len(items) {
		d.SelectedIdx = len(items) - 1
	}
	if d.SelectedIdx < 0 {
		d.SelectedIdx = 0
	}
}

// MoveUp moves selection up
func (d *DumpViewModel) MoveUp() {
	if d.SelectedIdx > 0 {
		d.SelectedIdx--
	}
}

// MoveDown moves selection down
func (d *DumpViewModel) MoveDown() {
	if d.SelectedIdx < len(d.Items)-1 {
		d.SelectedIdx++
	}
}

// GetSelectedItem returns the currently selected item
func (d *DumpViewModel) GetSelectedItem() *api.DumpItemJSON {
	if d.SelectedIdx >= 0 && d.SelectedIdx < len(d.Items) {
		return &d.Items[d.SelectedIdx]
	}
	return nil
}

// Render renders the dump inbox view
func (d *DumpViewModel) Render(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1)

	if len(d.Items) == 0 {
		content := lipgloss.NewStyle().Foreground(mutedColor).Render("Dump inbox is empty\n\nUse 'brain add <task>' to add items")
		return style.Render(content)
	}

	var lines []string
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	lines = append(lines, titleStyle.Render("Dump Inbox"))
	lines = append(lines, lipgloss.NewStyle().Foreground(mutedColor).Render("r: refile to project | x: delete"))
	lines = append(lines, "")

	// Render items
	viewportStart := 0
	viewportEnd := height - 5

	// Adjust viewport if selection is out of bounds
	if d.SelectedIdx >= viewportEnd {
		viewportStart = d.SelectedIdx - height + 6
		viewportEnd = d.SelectedIdx + 1
	} else if d.SelectedIdx < viewportStart {
		viewportStart = d.SelectedIdx
		viewportEnd = viewportStart + height - 5
	}

	for i := viewportStart; i < viewportEnd && i < len(d.Items); i++ {
		item := d.Items[i]
		isSelected := i == d.SelectedIdx

		indicator := "  "
		if isSelected {
			indicator = "> "
		}

		// Build item display
		typeIndicator := ""
		if item.Type == "task" {
			typeIndicator = "[ ] "
		} else if item.Type == "note" {
			typeIndicator = "• "
		}

		// Content
		content := item.Content
		maxContentLen := width - 20
		if len(content) > maxContentLen {
			content = content[:maxContentLen-3] + "..."
		}

		line := indicator + typeIndicator + content

		if isSelected {
			line = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(line)
		}

		lines = append(lines, line)

		// Add metadata on next line
		var metaParts []string
		metaParts = append(metaParts, "ID: "+item.ID)
		if item.Timestamp != "" {
			metaParts = append(metaParts, "captured: "+item.Timestamp)
		}
		if len(metaParts) > 0 {
			metaStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
			lines = append(lines, "   "+metaStyle.Render(strings.Join(metaParts, " | ")))
		}
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}
