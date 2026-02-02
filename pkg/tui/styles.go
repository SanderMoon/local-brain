package tui

import "github.com/charmbracelet/lipgloss"

// Color palette
var (
	primaryColor = lipgloss.Color("205") // Magenta
	warningColor = lipgloss.Color("208") // Orange
	errorColor   = lipgloss.Color("196") // Red
	mutedColor   = lipgloss.Color("240") // Gray
)

// Component styles
var (
	sidebarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1)

	sidebarFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1)

	contentBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(mutedColor).
			Padding(1)

	contentFocusedStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(1)

	statusBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("250")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 0, 1, 0)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	currentProjectStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Render
)

// Todo status styles
var (
	todoOpenStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255"))

	todoInProgressStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	todoBlockedStyle = lipgloss.NewStyle().
				Foreground(warningColor)

	todoDoneStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Strikethrough(true)
)

// Priority badge styles
var (
	priorityHighStyle = lipgloss.NewStyle().
				Background(errorColor).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1).
				Bold(true)

	priorityMediumStyle = lipgloss.NewStyle().
				Background(warningColor).
				Foreground(lipgloss.Color("0")).
				Padding(0, 1)

	priorityLowStyle = lipgloss.NewStyle().
				Background(mutedColor).
				Foreground(lipgloss.Color("255")).
				Padding(0, 1)
)

// Helper functions
func renderTodoStatus(status string) string {
	switch status {
	case "open":
		return todoOpenStyle.Render("[ ]")
	case "in-progress":
		return todoInProgressStyle.Render("[>]")
	case "blocked":
		return todoBlockedStyle.Render("[-]")
	case "done":
		return todoDoneStyle.Render("[x]")
	default:
		return "[ ]"
	}
}

func renderPriorityBadge(priority *int) string {
	if priority == nil {
		return ""
	}

	var style lipgloss.Style
	switch *priority {
	case 1:
		style = priorityHighStyle
	case 2:
		style = priorityMediumStyle
	case 3:
		style = priorityLowStyle
	default:
		return ""
	}

	return style.Render(lipgloss.JoinHorizontal(lipgloss.Left, "P", lipgloss.NewStyle().Render(string(rune('0'+*priority)))))
}
