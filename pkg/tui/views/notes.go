package views

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// NotesViewModel represents the notes view state
type NotesViewModel struct {
	Notes        []api.NoteFile
	SelectedIdx  int
	ShowPreview  bool
	PreviewPane  viewport.Model
	PreviewReady bool
}

// NewNotesViewModel creates a new notes view model
func NewNotesViewModel() NotesViewModel {
	vp := viewport.New(0, 0)
	return NotesViewModel{
		Notes:        []api.NoteFile{},
		SelectedIdx:  0,
		ShowPreview:  true,
		PreviewPane:  vp,
		PreviewReady: false,
	}
}

// UpdateNotes updates the notes list
func (n *NotesViewModel) UpdateNotes(notes []api.NoteFile) {
	n.Notes = notes
	if n.SelectedIdx >= len(notes) {
		n.SelectedIdx = len(notes) - 1
	}
	if n.SelectedIdx < 0 {
		n.SelectedIdx = 0
	}
}

// MoveUp moves selection up
func (n *NotesViewModel) MoveUp() {
	if n.SelectedIdx > 0 {
		n.SelectedIdx--
	}
}

// MoveDown moves selection down
func (n *NotesViewModel) MoveDown() {
	if n.SelectedIdx < len(n.Notes)-1 {
		n.SelectedIdx++
	}
}

// TogglePreview toggles the preview pane
func (n *NotesViewModel) TogglePreview() {
	n.ShowPreview = !n.ShowPreview
}

// GetSelectedNote returns the currently selected note
func (n *NotesViewModel) GetSelectedNote() *api.NoteFile {
	if n.SelectedIdx >= 0 && n.SelectedIdx < len(n.Notes) {
		return &n.Notes[n.SelectedIdx]
	}
	return nil
}

// LoadPreview loads the preview for the current note
func (n *NotesViewModel) LoadPreview() {
	note := n.GetSelectedNote()
	if note == nil {
		n.PreviewPane.SetContent("No note selected")
		return
	}

	content, err := os.ReadFile(note.Path)
	if err != nil {
		n.PreviewPane.SetContent("Error reading file: " + err.Error())
		return
	}

	n.PreviewPane.SetContent(string(content))
}

// SetPreviewSize sets the preview pane size
func (n *NotesViewModel) SetPreviewSize(width, height int) {
	// Account for borders (2) and padding (2) = 4 total
	n.PreviewPane.Width = width - 4
	n.PreviewPane.Height = height - 4
	n.PreviewReady = true
}

// Render renders the notes view
func (n *NotesViewModel) Render(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	if !n.ShowPreview {
		// Full-width list
		return n.renderList(width, height, primaryColor, mutedColor)
	}

	// Split view: list + preview
	listWidth := width / 3
	previewWidth := width - listWidth - 2

	// Update preview size if needed
	// Account for borders and padding (4 total) when comparing
	expectedWidth := previewWidth - 4
	expectedHeight := height - 4
	if !n.PreviewReady || n.PreviewPane.Width != expectedWidth || n.PreviewPane.Height != expectedHeight {
		n.SetPreviewSize(previewWidth, height)
		n.LoadPreview()
	}

	listView := n.renderList(listWidth, height, primaryColor, mutedColor)
	previewView := n.renderPreview(previewWidth, height, primaryColor, mutedColor)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, previewView)
}

// renderList renders the notes list
func (n *NotesViewModel) renderList(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	listStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height). // Prevents terminal scroll push
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(mutedColor).
		Padding(1)

	if len(n.Notes) == 0 {
		content := lipgloss.NewStyle().Foreground(mutedColor).Render("No notes found")
		return listStyle.Render(content)
	}

	var lines []string
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(primaryColor)
	lines = append(lines, titleStyle.Render("Notes"))
	lines = append(lines, "")

	// Render notes
	viewportStart := 0
	viewportEnd := height - 4

	// Adjust viewport if selection is out of bounds
	if n.SelectedIdx >= viewportEnd {
		viewportStart = n.SelectedIdx - height + 5
		viewportEnd = n.SelectedIdx + 1
	} else if n.SelectedIdx < viewportStart {
		viewportStart = n.SelectedIdx
		viewportEnd = viewportStart + height - 4
	}

	for i := viewportStart; i < viewportEnd && i < len(n.Notes); i++ {
		note := n.Notes[i]
		isSelected := i == n.SelectedIdx

		indicator := "  "
		if isSelected {
			indicator = "> "
		}

		title := note.Title
		if title == "" {
			title = "(untitled)"
		}
		if len(title) > width-10 {
			title = title[:width-13] + "..."
		}

		line := indicator + title

		if isSelected {
			line = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(line)
		}

		lines = append(lines, line)

		// Add date on next line
		if note.Created != "" {
			dateStyle := lipgloss.NewStyle().Foreground(mutedColor).Italic(true)
			lines = append(lines, "   "+dateStyle.Render(note.Created))
		}
	}

	content := strings.Join(lines, "\n")
	return listStyle.Render(content)
}

// renderPreview renders the preview pane
func (n *NotesViewModel) renderPreview(width, height int, primaryColor, mutedColor lipgloss.Color) string {
	previewStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		MaxHeight(height). // Prevents terminal scroll push
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1)

	if !n.PreviewReady {
		return previewStyle.Render(lipgloss.NewStyle().Foreground(mutedColor).Render("Loading..."))
	}

	content := n.PreviewPane.View()
	return previewStyle.Render(content)
}
