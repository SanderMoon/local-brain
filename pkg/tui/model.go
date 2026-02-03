package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/tui/views"
)

// FocusArea represents which area of the UI has focus
type FocusArea int

const (
	FocusSidebar FocusArea = iota
	FocusContent
)

// SidebarSection represents which section of the sidebar is focused
type SidebarSection int

const (
	SidebarBrains SidebarSection = iota
	SidebarProjects
)

// ViewType represents the current active view
type ViewType int

const (
	ViewTodosList ViewType = iota
	ViewNotes
	ViewTodosKanban
)

// Model represents the root application state
type Model struct {
	config         *config.Config
	width, height  int
	currentBrain   string
	currentProject string
	focusedArea    FocusArea
	activeView     ViewType
	sidebarSection SidebarSection

	// Components
	brainList   list.Model
	projectList list.Model
	kanbanView  views.KanbanViewModel
	notesView   views.NotesViewModel

	// Data cache
	brains   []Brain
	projects []api.ProjectInfo
	todos    []api.TodoItem
	notes    []api.NoteFile

	// UI state
	selectedBrainIdx   int
	selectedProjectIdx int
	selectedTodoIdx    int
	loading            bool
	err                error
	showHelp           bool
	showInput          bool
	showCompleted      bool // Toggle to show/hide completed todos
	showAllProjects    bool // Toggle to show todos from all projects
	inputPrompt        string
	inputMode          InputMode
	textInput          textinput.Model
}

// InputMode represents what the input is for
type InputMode int

const (
	InputNone InputMode = iota
	InputPriority
	InputDueDate
	InputTags
)

// NewModel creates a new TUI model
func NewModel(cfg *config.Config) Model {
	// Create brain list
	brainList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	brainList.Title = ""
	brainList.SetShowStatusBar(false)
	brainList.SetFilteringEnabled(false)
	brainList.SetShowHelp(false)

	// Create project list
	projectList := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	projectList.Title = ""
	projectList.SetShowStatusBar(false)
	projectList.SetFilteringEnabled(false)
	projectList.SetShowHelp(false)

	// Create text input for prompts
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 100
	ti.Width = 50

	return Model{
		config:             cfg,
		currentBrain:       cfg.GetCurrentBrain(),
		currentProject:     cfg.GetFocusedProject(),
		focusedArea:        FocusSidebar,
		sidebarSection:     SidebarProjects,
		activeView:         ViewTodosList,
		brainList:          brainList,
		projectList:        projectList,
		kanbanView:         views.NewKanbanViewModel(),
		notesView:          views.NewNotesViewModel(),
		textInput:          ti,
		selectedBrainIdx:   0,
		selectedProjectIdx: 0,
		selectedTodoIdx:    0,
	}
}

// Brain represents a brain in the TUI
type Brain struct {
	Name string
}

// brainItem implements list.Item for brain names
type brainItem struct {
	name      string
	isCurrent bool
}

func (b brainItem) FilterValue() string { return b.name }
func (b brainItem) Title() string       { return b.name }
func (b brainItem) Description() string {
	if b.isCurrent {
		return "← active"
	}
	return ""
}

// projectItem implements list.Item for project names
type projectItem struct {
	name      string
	isCurrent bool
}

func (p projectItem) FilterValue() string { return p.name }
func (p projectItem) Title() string       { return p.name }
func (p projectItem) Description() string {
	if p.isCurrent {
		return "← current"
	}
	return ""
}
