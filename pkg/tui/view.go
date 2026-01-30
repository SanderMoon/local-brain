package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// View renders the entire TUI
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Show error overlay if there's an error
	if m.err != nil {
		return m.renderErrorOverlay()
	}

	// Show help overlay if requested
	if m.showHelp {
		return m.renderHelpOverlay()
	}

	// Show input prompt if requested
	if m.showInput {
		return m.renderInputPrompt()
	}

	// Calculate dimensions
	sidebarWidth := 30
	contentWidth := m.width - sidebarWidth - 2
	contentHeight := m.height - 3 // Reserve space for status bar

	// Render components
	sidebarView := m.renderSidebar(sidebarWidth, contentHeight)
	contentView := m.renderContent(contentWidth, contentHeight)
	statusView := m.renderStatusBar(m.width)

	// Assemble layout
	mainLayout := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, contentView)
	fullLayout := lipgloss.JoinVertical(lipgloss.Left, mainLayout, statusView)

	return fullLayout
}

// renderSidebar renders the sidebar with brain and project list
func (m Model) renderSidebar(width, height int) string {
	style := sidebarStyle
	if m.focusedArea == FocusSidebar {
		style = sidebarFocusedStyle
	}

	var sections []string

	// Brains section
	brainsTitleStyle := titleStyle
	if m.focusedArea == FocusSidebar && m.sidebarSection == SidebarBrains {
		brainsTitleStyle = titleStyle.Copy().Foreground(primaryColor).Underline(true)
	}
	brainsTitle := brainsTitleStyle.Render("Brains")
	sections = append(sections, brainsTitle)

	// Render brains
	if len(m.brains) == 0 {
		sections = append(sections, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no brains)"))
	} else {
		for i, brain := range m.brains {
			indicator := "  "
			if i == m.brainList.Index() && m.focusedArea == FocusSidebar && m.sidebarSection == SidebarBrains {
				indicator = "> "
			}

			name := brain.Name
			if brain.Name == m.currentBrain {
				name = currentProjectStyle(name)
			}

			sections = append(sections, indicator+name)
		}
	}

	sections = append(sections, "") // Empty line

	// Projects section
	projectsTitleStyle := titleStyle
	if m.focusedArea == FocusSidebar && m.sidebarSection == SidebarProjects {
		projectsTitleStyle = titleStyle.Copy().Foreground(primaryColor).Underline(true)
	}
	projectsTitle := projectsTitleStyle.Render("Projects")
	sections = append(sections, projectsTitle)

	// Render projects
	if len(m.projects) == 0 {
		sections = append(sections, lipgloss.NewStyle().Foreground(mutedColor).Render("  (no projects)"))
	} else {
		for i, proj := range m.projects {
			indicator := "  "
			if i == m.projectList.Index() && m.focusedArea == FocusSidebar && m.sidebarSection == SidebarProjects {
				indicator = "> "
			}

			name := proj.Name
			if proj.Name == m.currentProject {
				name = currentProjectStyle(name)
			}

			sections = append(sections, indicator+name)
		}
	}

	content := strings.Join(sections, "\n")

	// Add some padding to fill the height
	lines := strings.Split(content, "\n")
	remainingHeight := height - len(lines) - 2
	if remainingHeight > 0 {
		content += strings.Repeat("\n", remainingHeight)
	}

	return style.Width(width).Height(height).Render(content)
}

// renderContent renders the main content area
func (m Model) renderContent(width, height int) string {
	style := contentBoxStyle
	if m.focusedArea == FocusContent {
		style = contentFocusedStyle
	}

	var content string

	switch m.activeView {
	case ViewTodosList:
		content = m.renderTodoList(width-4, height-4)
	case ViewNotes:
		content = m.notesView.Render(width-4, height-4, primaryColor, mutedColor)
	case ViewTodosKanban:
		content = m.kanbanView.Render(width-4, height-4, primaryColor, mutedColor)
	case ViewDump:
		content = m.dumpView.Render(width-4, height-4, primaryColor, mutedColor)
	}

	return style.Width(width).Height(height).Render(content)
}

// renderTodoList renders the todo list view
func (m Model) renderTodoList(width, height int) string {
	if m.loading {
		return "Loading todos..."
	}

	if m.err != nil {
		return lipgloss.NewStyle().Foreground(errorColor).Render("Error: " + m.err.Error())
	}

	if len(m.todos) == 0 {
		return lipgloss.NewStyle().Foreground(mutedColor).Render("No todos found")
	}

	var lines []string
	viewportStart := 0
	viewportEnd := height

	// Adjust viewport if selection is out of bounds
	if m.selectedTodoIdx >= viewportEnd {
		viewportStart = m.selectedTodoIdx - height + 1
		viewportEnd = m.selectedTodoIdx + 1
	} else if m.selectedTodoIdx < viewportStart {
		viewportStart = m.selectedTodoIdx
		viewportEnd = viewportStart + height
	}

	for i := viewportStart; i < viewportEnd && i < len(m.todos); i++ {
		todo := m.todos[i]
		isSelected := i == m.selectedTodoIdx && m.focusedArea == FocusContent

		// Build todo line
		var parts []string

		// Status icon
		parts = append(parts, renderTodoStatus(todo.Status))

		// Priority badge
		if todo.Priority != nil {
			parts = append(parts, renderPriorityBadge(todo.Priority))
		}

		// Content
		content := todo.Content
		if len(content) > width-20 {
			content = content[:width-23] + "..."
		}
		parts = append(parts, content)

		line := strings.Join(parts, " ")

		// Apply selection styling
		if isSelected {
			line = selectedItemStyle.Render("> " + line)
		} else {
			line = normalItemStyle.Render("  " + line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderNotesPlaceholder is a placeholder for notes view
func (m Model) renderNotesPlaceholder(width, height int) string {
	return lipgloss.NewStyle().Foreground(mutedColor).Render("Notes view - coming soon\nPress 2 for todos")
}

// renderKanbanPlaceholder is a placeholder for kanban view
func (m Model) renderKanbanPlaceholder(width, height int) string {
	return lipgloss.NewStyle().Foreground(mutedColor).Render("Kanban view - coming soon\nPress 2 for todos list")
}

// renderDumpPlaceholder is a placeholder for dump view
func (m Model) renderDumpPlaceholder(width, height int) string {
	return lipgloss.NewStyle().Foreground(mutedColor).Render("Dump view - coming soon\nPress 2 for todos")
}

// renderStatusBar renders the bottom status bar
func (m Model) renderStatusBar(width int) string {
	var indicators []string
	if m.showCompleted {
		indicators = append(indicators, "✓ Completed")
	}
	if m.showAllProjects {
		indicators = append(indicators, "✓ All projects")
	}

	indicatorText := ""
	if len(indicators) > 0 {
		indicatorText = " | " + strings.Join(indicators, " | ")
	}

	projectDisplay := m.currentProject
	if projectDisplay == "" {
		projectDisplay = "(none)"
	}

	leftText := fmt.Sprintf("Brain: %s | Project: %s | View: %s%s",
		m.currentBrain, projectDisplay, m.getViewName(), indicatorText)

	rightText := m.getKeyHints()

	// Calculate padding
	leftWidth := lipgloss.Width(leftText)
	rightWidth := lipgloss.Width(rightText)
	padding := width - leftWidth - rightWidth - 4
	if padding < 0 {
		padding = 0
	}

	return statusBarStyle.Width(width).Render(
		leftText + strings.Repeat(" ", padding) + rightText,
	)
}

// renderErrorOverlay renders an error message overlay
func (m Model) renderErrorOverlay() string {
	errorText := lipgloss.NewStyle().Bold(true).Foreground(errorColor).Render("Error")
	messageText := m.err.Error()

	content := lipgloss.JoinVertical(lipgloss.Left,
		errorText,
		"",
		messageText,
		"",
		lipgloss.NewStyle().Foreground(mutedColor).Render("Press any key to dismiss"),
	)

	errorPanel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(errorColor).
		Padding(1, 2).
		Width(60).
		Render(content)

	// Center the error panel
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		errorPanel,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
	)
}

// renderInputPrompt renders the input prompt overlay
func (m Model) renderInputPrompt() string {
	promptText := lipgloss.NewStyle().Bold(true).Foreground(primaryColor).Render(m.inputPrompt)
	inputView := m.textInput.View()

	content := lipgloss.JoinVertical(lipgloss.Left,
		promptText,
		"",
		inputView,
		"",
		lipgloss.NewStyle().Foreground(mutedColor).Render("Enter to submit, Esc to cancel"),
	)

	inputPanel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(1, 2).
		Width(60).
		Render(content)

	// Center the input panel
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		inputPanel,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
	)
}

// renderHelpOverlay renders the help panel
func (m Model) renderHelpOverlay() string {
	sections := []string{
		titleStyle.Render("Local Brain TUI - Keyboard Shortcuts"),
		"",
		lipgloss.NewStyle().Bold(true).Render("Navigation"),
		"  Tab         Toggle sidebar / content focus",
		"  ↑↓ or j/k   Navigate lists",
		"  ←→ or h/l   Switch sidebar section (brains/projects)",
		"  Enter       Select item (brain or project)",
		"",
		lipgloss.NewStyle().Bold(true).Render("Views"),
		"  1           Notes view",
		"  2           Todos list view",
		"  3           Todos kanban view",
		"  4           Dump inbox view",
		"",
		lipgloss.NewStyle().Bold(true).Render("Actions"),
		"  e           Edit in external editor",
		"  p           Set priority (todos) / toggle preview (notes)",
		"  s           Cycle status (todos)",
		"  d           Set due date (todos)",
		"  t           Add tags (todos)",
		"  c           Toggle showing completed todos",
		"  a           Toggle showing all projects (vs current only)",
		"  r           Refresh data",
		"",
		lipgloss.NewStyle().Bold(true).Render("Other"),
		"  ?           Toggle this help",
		"  q / ctrl+c  Quit",
	}

	content := strings.Join(sections, "\n")

	helpPanel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(primaryColor).
		Padding(2).
		Width(60).
		Render(content)

	// Center the help panel
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpPanel,
		lipgloss.WithWhitespaceChars("░"),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("236")),
	)
}
