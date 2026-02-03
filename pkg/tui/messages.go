package tui

import (
	"github.com/sandermoonemans/local-brain/pkg/api"
)

// Custom message types for Bubble Tea

// DataRefreshedMsg is sent when data has been reloaded
type DataRefreshedMsg struct {
	Brains   []Brain
	Projects []api.ProjectInfo
	Todos    []api.TodoItem
	Notes    []api.NoteFile
	Err      error
}

// EditorClosedMsg is sent when external editor closes
type EditorClosedMsg struct {
	Error error
}

// ProjectSelectedMsg is sent when a project is selected in sidebar
type ProjectSelectedMsg struct {
	ProjectName string
}

// DataRefreshNeededMsg signals that data needs to be refreshed
type DataRefreshNeededMsg struct{}
