package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// Update handles incoming messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Set list sizes (half the sidebar height for each)
		listHeight := (m.height - 10) / 2
		m.brainList.SetSize(28, listHeight)
		m.projectList.SetSize(28, listHeight)
		return m, nil

	case tea.KeyMsg:
		// Dismiss error on any key
		if m.err != nil {
			m.err = nil
			return m, nil
		}

		// Global shortcuts
		switch msg.String() {
		case KeyQuit, KeyCtrlC:
			return m, tea.Quit

		case KeyHelp:
			m.showHelp = !m.showHelp
			return m, nil

		case KeyTab:
			// Toggle focus
			if m.focusedArea == FocusSidebar {
				m.focusedArea = FocusContent
			} else {
				m.focusedArea = FocusSidebar
			}
			return m, nil

		case KeyRefresh:
			return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)

		case "c": // Toggle showing completed todos
			m.showCompleted = !m.showCompleted
			return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)

		case "a": // Toggle showing all projects
			m.showAllProjects = !m.showAllProjects
			return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)

		case Key1:
			m.activeView = ViewNotes
			return m, nil
		case Key2:
			m.activeView = ViewTodosList
			return m, nil
		case Key3:
			m.activeView = ViewTodosKanban
			return m, nil
		case Key4:
			m.activeView = ViewDump
			return m, nil
		}

		// Handle focus-specific keys
		if m.focusedArea == FocusSidebar {
			return m.updateSidebar(msg)
		} else {
			return m.updateContent(msg)
		}

	case DataRefreshedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.brains = msg.Brains
		m.projects = msg.Projects
		m.todos = msg.Todos
		m.notes = msg.Notes
		m.dumpItems = msg.DumpItems
		m.updateBrainList()
		m.updateProjectList()
		m.kanbanView.UpdateTodos(msg.Todos)
		m.notesView.UpdateNotes(msg.Notes)
		m.dumpView.UpdateItems(msg.DumpItems)
		return m, nil

	case EditorClosedMsg:
		if msg.Error != nil {
			m.err = msg.Error
		}
		// Refresh data after editor closes
		return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)

	case DataRefreshNeededMsg:
		return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)
	}

	return m, nil
}

// updateSidebar handles updates when sidebar has focus
func (m Model) updateSidebar(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.String() {
	case KeyDown, KeyArrowDown:
		if m.sidebarSection == SidebarBrains {
			m.brainList, cmd = m.brainList.Update(msg)
			// Check if we should switch to projects section
			if m.brainList.Index() >= len(m.brains)-1 {
				// At the end of brains, stay here
			}
		} else {
			m.projectList, cmd = m.projectList.Update(msg)
		}
		return m, cmd

	case KeyUp, KeyArrowUp:
		if m.sidebarSection == SidebarBrains {
			m.brainList, cmd = m.brainList.Update(msg)
		} else {
			m.projectList, cmd = m.projectList.Update(msg)
			// Check if we should switch to brains section
			if m.projectList.Index() == 0 {
				// At the top of projects, could switch to brains
				// But let's keep it simple for now
			}
		}
		return m, cmd

	case KeyLeft, KeyArrowLeft:
		// Switch to brains section
		if m.sidebarSection == SidebarProjects {
			m.sidebarSection = SidebarBrains
		}
		return m, nil

	case KeyRight, KeyArrowRight:
		// Switch to projects section
		if m.sidebarSection == SidebarBrains {
			m.sidebarSection = SidebarProjects
		}
		return m, nil

	case KeyEnter:
		if m.sidebarSection == SidebarBrains {
			// Select brain
			if selected, ok := m.brainList.SelectedItem().(brainItem); ok {
				// Update config and switch brain
				if err := m.config.SetCurrentBrain(selected.name); err != nil {
					m.err = fmt.Errorf("failed to set current brain: %w", err)
					return m, nil
				}

				// Update the symlink
				if err := config.UpdateSymlink(selected.name, m.config); err != nil {
					m.err = fmt.Errorf("failed to update symlink: %w", err)
					return m, nil
				}

				// Save config
				if err := m.config.Save(); err != nil {
					m.err = fmt.Errorf("failed to save config: %w", err)
					return m, nil
				}

				m.currentBrain = selected.name
				return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)
			}
		} else {
			// Select project
			if selected, ok := m.projectList.SelectedItem().(projectItem); ok {
				m.currentProject = selected.name
				// Update config
				if err := m.config.SetFocusedProject(selected.name); err == nil {
					_ = m.config.Save()
				}
				return m, refreshDataCmd(m.config, m.showCompleted, m.showAllProjects)
			}
		}
	}

	return m, nil
}

// updateContent handles updates when content area has focus
func (m Model) updateContent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input mode
	if m.showInput {
		return m.handleInputMode(msg)
	}

	// Handle dump navigation
	if m.activeView == ViewDump {
		switch msg.String() {
		case KeyDown, KeyArrowDown:
			m.dumpView.MoveDown()
			return m, nil
		case KeyUp, KeyArrowUp:
			m.dumpView.MoveUp()
			return m, nil
		case "r": // Refile
			item := m.dumpView.GetSelectedItem()
			if item != nil {
				// Show project selector prompt
				m.showInput = true
				m.inputMode = InputNone // We'll use a special mode
				m.inputPrompt = "Refile to project (or cancel with Esc): "
				m.textInput.SetValue("")
				m.textInput.Placeholder = "project-name"
				m.textInput.Focus()
				return m, nil
			}
			return m, nil
		}
		return m, nil
	}

	// Handle notes navigation
	if m.activeView == ViewNotes {
		switch msg.String() {
		case KeyDown, KeyArrowDown:
			m.notesView.MoveDown()
			if m.notesView.ShowPreview {
				m.notesView.LoadPreview()
			}
			return m, nil
		case KeyUp, KeyArrowUp:
			m.notesView.MoveUp()
			if m.notesView.ShowPreview {
				m.notesView.LoadPreview()
			}
			return m, nil
		case "p": // Toggle preview
			m.notesView.TogglePreview()
			if m.notesView.ShowPreview {
				m.notesView.LoadPreview()
			}
			return m, nil
		case KeyEdit:
			note := m.notesView.GetSelectedNote()
			if note != nil {
				return m, launchNoteEditorCmd(note.Path)
			}
			return m, nil
		}
		return m, nil
	}

	// Handle kanban navigation
	if m.activeView == ViewTodosKanban {
		switch msg.String() {
		case KeyLeft, KeyArrowLeft:
			m.kanbanView.MoveLeft()
			return m, nil
		case KeyRight, KeyArrowRight:
			m.kanbanView.MoveRight()
			return m, nil
		case KeyDown, KeyArrowDown:
			m.kanbanView.MoveDown()
			return m, nil
		case KeyUp, KeyArrowUp:
			m.kanbanView.MoveUp()
			return m, nil
		case "m": // Move task to next column
			todo := m.kanbanView.GetSelectedTodo()
			if todo != nil {
				return m, cycleStatusCmd(*todo)
			}
			return m, nil
		case KeyEdit:
			todo := m.kanbanView.GetSelectedTodo()
			if todo != nil {
				return m, launchEditorCmd(*todo)
			}
			return m, nil
		}
		return m, nil
	}

	// Handle todo list navigation
	switch msg.String() {
	case KeyDown, KeyArrowDown:
		if m.selectedTodoIdx < len(m.todos)-1 {
			m.selectedTodoIdx++
		}
		return m, nil

	case KeyUp, KeyArrowUp:
		if m.selectedTodoIdx > 0 {
			m.selectedTodoIdx--
		}
		return m, nil

	case KeyEdit:
		if m.activeView == ViewTodosList && len(m.todos) > 0 && m.selectedTodoIdx < len(m.todos) {
			todo := m.todos[m.selectedTodoIdx]
			return m, launchEditorCmd(todo)
		}

	case "p": // Set priority
		if m.activeView == ViewTodosList && len(m.todos) > 0 && m.selectedTodoIdx < len(m.todos) {
			m.showInput = true
			m.inputMode = InputPriority
			m.inputPrompt = "Priority (1=high, 2=med, 3=low): "
			m.textInput.SetValue("")
			m.textInput.Placeholder = "1-3"
			m.textInput.Focus()
			return m, nil
		}

	case "s": // Cycle status
		if m.activeView == ViewTodosList && len(m.todos) > 0 && m.selectedTodoIdx < len(m.todos) {
			return m, cycleStatusCmd(m.todos[m.selectedTodoIdx])
		}

	case "d": // Set due date
		if m.activeView == ViewTodosList && len(m.todos) > 0 && m.selectedTodoIdx < len(m.todos) {
			m.showInput = true
			m.inputMode = InputDueDate
			m.inputPrompt = "Due date (YYYY-MM-DD or 'tomorrow', '+3d'): "
			m.textInput.SetValue("")
			m.textInput.Placeholder = "tomorrow, +3d, 2026-02-01"
			m.textInput.Focus()
			return m, nil
		}

	case "t": // Add tags
		if m.activeView == ViewTodosList && len(m.todos) > 0 && m.selectedTodoIdx < len(m.todos) {
			m.showInput = true
			m.inputMode = InputTags
			m.inputPrompt = "Tags (space-separated): "
			m.textInput.SetValue("")
			m.textInput.Placeholder = "tag1 tag2 tag3"
			m.textInput.Focus()
			return m, nil
		}
	}

	return m, nil
}

// updateBrainList rebuilds the brain list from current data
func (m *Model) updateBrainList() {
	items := make([]list.Item, len(m.brains))
	for i, brain := range m.brains {
		items[i] = brainItem{
			name:      brain.Name,
			isCurrent: brain.Name == m.currentBrain,
		}
	}
	m.brainList.SetItems(items)
}

// updateProjectList rebuilds the project list from current data
func (m *Model) updateProjectList() {
	items := make([]list.Item, len(m.projects))
	for i, proj := range m.projects {
		items[i] = projectItem{
			name:      proj.Name,
			isCurrent: proj.Name == m.currentProject,
		}
	}
	m.projectList.SetItems(items)
}

// refreshDataCmd loads brain, project, and todo data
func refreshDataCmd(cfg *config.Config, showCompleted bool, showAllProjects bool) tea.Cmd {
	return func() tea.Msg {
		// Load all brains
		brainNames := cfg.ListBrains()
		brains := make([]Brain, len(brainNames))
		for i, name := range brainNames {
			brains[i] = Brain{Name: name}
		}

		brainPath, err := cfg.GetCurrentBrainPath()
		if err != nil {
			return DataRefreshedMsg{Err: err}
		}

		activeDir := filepath.Join(brainPath, "01_active")

		// Load projects
		projects, err := api.ListProjects(activeDir, cfg.GetFocusedProject())
		if err != nil {
			return DataRefreshedMsg{Err: err}
		}

		// Load todos for current project
		todos, err := api.ParseAllTodos(activeDir, showCompleted)
		if err != nil {
			return DataRefreshedMsg{Err: err}
		}

		// Filter todos for focused project if set (unless showing all projects)
		focusedProject := cfg.GetFocusedProject()
		if !showAllProjects && focusedProject != "" {
			filteredTodos := make([]api.TodoItem, 0)
			for _, todo := range todos {
				if todo.Project == focusedProject {
					filteredTodos = append(filteredTodos, todo)
				}
			}
			todos = filteredTodos
		}

		// Load notes
		notes := []api.NoteFile{}
		if focusedProject != "" {
			projectDir := filepath.Join(activeDir, focusedProject)
			notes, _ = api.ListNotes(projectDir) // Ignore errors, just return empty
		}

		// Load dump items
		dumpPath := filepath.Join(brainPath, "00_dump.md")
		dumpItems, _ := api.ParseDumpToJSON(dumpPath) // Ignore errors, just return empty

		return DataRefreshedMsg{
			Brains:    brains,
			Projects:  projects,
			Todos:     todos,
			Notes:     notes,
			DumpItems: dumpItems,
		}
	}
}

// handleInputMode handles keyboard input when in input mode
func (m Model) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEscape:
		// Cancel input
		m.showInput = false
		m.inputMode = InputNone
		m.textInput.Blur()
		return m, nil

	case tea.KeyEnter:
		// Submit input
		value := m.textInput.Value()
		m.showInput = false
		m.textInput.Blur()

		if len(m.todos) == 0 || m.selectedTodoIdx >= len(m.todos) {
			m.inputMode = InputNone
			return m, nil
		}

		todo := m.todos[m.selectedTodoIdx]

		switch m.inputMode {
		case InputPriority:
			m.inputMode = InputNone
			return m, setPriorityCmd(todo, value)
		case InputDueDate:
			m.inputMode = InputNone
			return m, setDueDateCmd(todo, value)
		case InputTags:
			m.inputMode = InputNone
			return m, addTagsCmd(todo, value)
		}

		m.inputMode = InputNone
		return m, nil
	}

	// Update text input
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// launchEditorCmd opens a todo in the external editor
func launchEditorCmd(todo api.TodoItem) tea.Cmd {
	// Get the editor from environment or use vim as default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Build command to open file at specific line
	cmd := exec.Command(editor, "+"+strconv.Itoa(todo.Line), todo.File)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return EditorClosedMsg{Error: err}
	})
}

// launchNoteEditorCmd opens a note file in the external editor
func launchNoteEditorCmd(notePath string) tea.Cmd {
	// Get the editor from environment or use vim as default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	// Build command to open file
	cmd := exec.Command(editor, notePath)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return EditorClosedMsg{Error: err}
	})
}

// setPriorityCmd sets the priority of a todo
func setPriorityCmd(todo api.TodoItem, priorityStr string) tea.Cmd {
	return func() tea.Msg {
		priority, err := strconv.Atoi(priorityStr)
		if err != nil || priority < 1 || priority > 3 {
			return DataRefreshNeededMsg{} // Invalid input, just refresh
		}

		err = api.SetTodoPriority(&todo, &priority)
		if err != nil {
			return DataRefreshNeededMsg{}
		}

		return DataRefreshNeededMsg{}
	}
}

// cycleStatusCmd cycles through todo statuses
func cycleStatusCmd(todo api.TodoItem) tea.Cmd {
	return func() tea.Msg {
		statuses := []string{"open", "in-progress", "blocked", "done"}
		nextStatus := "open"

		for i, status := range statuses {
			if status == todo.Status {
				nextStatus = statuses[(i+1)%len(statuses)]
				break
			}
		}

		err := api.SetTodoStatus(&todo, nextStatus)
		if err != nil {
			return DataRefreshNeededMsg{}
		}

		return DataRefreshNeededMsg{}
	}
}

// setDueDateCmd sets the due date of a todo
func setDueDateCmd(todo api.TodoItem, dateStr string) tea.Cmd {
	return func() tea.Msg {
		err := api.SetTodoDueDate(&todo, dateStr)
		if err != nil {
			return DataRefreshNeededMsg{}
		}

		return DataRefreshNeededMsg{}
	}
}

// addTagsCmd adds tags to a todo
func addTagsCmd(todo api.TodoItem, tagsStr string) tea.Cmd {
	return func() tea.Msg {
		tags := strings.Fields(tagsStr)
		if len(tags) == 0 {
			return DataRefreshNeededMsg{}
		}

		err := api.AddTodoTags(&todo, tags)
		if err != nil {
			return DataRefreshNeededMsg{}
		}

		return DataRefreshNeededMsg{}
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		refreshDataCmd(m.config, m.showCompleted, m.showAllProjects),
		tea.EnterAltScreen,
	)
}

// getViewName returns the human-readable name of the current view
func (m Model) getViewName() string {
	switch m.activeView {
	case ViewNotes:
		return "Notes"
	case ViewTodosList:
		return "Todos (List)"
	case ViewTodosKanban:
		return "Todos (Kanban)"
	case ViewDump:
		return "Dump"
	default:
		return ""
	}
}

// getKeyHints returns context-sensitive keyboard hints
func (m Model) getKeyHints() string {
	if m.showHelp {
		return "Press ? to close help"
	}

	if m.focusedArea == FocusSidebar {
		return "↑↓: navigate | ←→: switch section | enter: select | tab: content | ?: help | q: quit"
	}

	switch m.activeView {
	case ViewTodosList:
		return fmt.Sprintf("↑↓: navigate | e: edit | p: priority | s: status | c: completed | a: all projects | ?: help")
	case ViewNotes:
		return "↑↓: navigate | e: edit | p: toggle preview | tab: sidebar | 1-4: views | ?: help | q: quit"
	case ViewTodosKanban:
		return "↑↓←→: navigate | e: edit | m: move | c: completed | a: all projects | ?: help"
	case ViewDump:
		return "↑↓: navigate | r: refile | tab: sidebar | 1-4: views | ?: help | q: quit"
	default:
		return "tab: sidebar | 1-4: views | c: completed | a: all projects | ?: help | q: quit"
	}
}
