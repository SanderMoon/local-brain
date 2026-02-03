package tui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandermoonemans/local-brain/pkg/api"
	"github.com/sandermoonemans/local-brain/pkg/config"
	"github.com/sandermoonemans/local-brain/pkg/testutil"
)

func TestCycleStatusCmd_Forward(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	tb.AddProject("test-project")

	todoFile := filepath.Join(tb.ActiveDirPath, "test-project", "todo.md")
	content := `# Test Project

## Active

- [ ] Test task

## Done
`
	if err := os.WriteFile(todoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	// Parse the todo
	todos, err := api.ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("Failed to parse todos: %v", err)
	}

	if len(todos) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(todos))
	}

	tests := []struct {
		initialStatus string
		expectedNext  string
	}{
		{"open", "in-progress"},
		{"in-progress", "blocked"},
		{"blocked", "done"},
		{"done", "open"},
	}

	for _, tt := range tests {
		t.Run(tt.initialStatus+"->"+tt.expectedNext, func(t *testing.T) {
			todo := todos[0]
			todo.Status = tt.initialStatus

			// Execute the command
			cmd := cycleStatusCmd(todo)
			msg := cmd()

			// Should return DataRefreshNeededMsg
			if _, ok := msg.(DataRefreshNeededMsg); !ok {
				t.Errorf("Expected DataRefreshNeededMsg, got %T", msg)
			}

			// Re-read the file to verify the status changed
			updatedTodos, err := api.ParseAllTodos(tb.ActiveDirPath, true)
			if err != nil {
				t.Fatalf("Failed to parse updated todos: %v", err)
			}

			if len(updatedTodos) != 1 {
				t.Fatalf("Expected 1 todo after update, got %d", len(updatedTodos))
			}

			if updatedTodos[0].Status != tt.expectedNext {
				t.Errorf("Expected status %s, got %s", tt.expectedNext, updatedTodos[0].Status)
			}
		})
	}
}

func TestCycleStatusBackwardCmd(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	tb.AddProject("test-project")

	todoFile := filepath.Join(tb.ActiveDirPath, "test-project", "todo.md")
	content := `# Test Project

## Active

- [ ] Test task

## Done
`
	if err := os.WriteFile(todoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write todo file: %v", err)
	}

	// Parse the todo
	todos, err := api.ParseAllTodos(tb.ActiveDirPath, false)
	if err != nil {
		t.Fatalf("Failed to parse todos: %v", err)
	}

	if len(todos) != 1 {
		t.Fatalf("Expected 1 todo, got %d", len(todos))
	}

	tests := []struct {
		initialStatus string
		expectedPrev  string
	}{
		{"done", "blocked"},
		{"blocked", "in-progress"},
		{"in-progress", "open"},
		{"open", "done"},
	}

	for _, tt := range tests {
		t.Run(tt.initialStatus+"->"+tt.expectedPrev, func(t *testing.T) {
			todo := todos[0]
			todo.Status = tt.initialStatus

			// Execute the command
			cmd := cycleStatusBackwardCmd(todo)
			msg := cmd()

			// Should return DataRefreshNeededMsg
			if _, ok := msg.(DataRefreshNeededMsg); !ok {
				t.Errorf("Expected DataRefreshNeededMsg, got %T", msg)
			}

			// Re-read the file to verify the status changed
			updatedTodos, err := api.ParseAllTodos(tb.ActiveDirPath, true)
			if err != nil {
				t.Fatalf("Failed to parse updated todos: %v", err)
			}

			if len(updatedTodos) != 1 {
				t.Fatalf("Expected 1 todo after update, got %d", len(updatedTodos))
			}

			if updatedTodos[0].Status != tt.expectedPrev {
				t.Errorf("Expected status %s, got %s", tt.expectedPrev, updatedTodos[0].Status)
			}
		})
	}
}

func TestModel_UpdateSidebar_NavigateWithArrows(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	_ = tb // tb sets up environment variables

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	m := NewModel(cfg)
	m.focusedArea = FocusSidebar
	m.sidebarSection = SidebarBrains

	// Navigate right to projects section
	keyMsg := tea.KeyMsg{Type: tea.KeyRight}
	newModel, _ := m.updateSidebar(keyMsg)
	m = newModel.(Model)

	if m.sidebarSection != SidebarProjects {
		t.Errorf("Expected sidebar section to be SidebarProjects, got %v", m.sidebarSection)
	}

	// Navigate left back to brains section
	keyMsg = tea.KeyMsg{Type: tea.KeyLeft}
	newModel, _ = m.updateSidebar(keyMsg)
	m = newModel.(Model)

	if m.sidebarSection != SidebarBrains {
		t.Errorf("Expected sidebar section to be SidebarBrains, got %v", m.sidebarSection)
	}
}

func TestModel_UpdateSidebar_ProjectSelection(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	tb.AddProject("project-1")
	tb.AddProject("project-2")

	// Load config
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	m := NewModel(cfg)
	m.projects = []api.ProjectInfo{
		{Name: "project-1"},
		{Name: "project-2"},
	}
	m.updateProjectList()

	// Set focus to sidebar and projects section
	m.focusedArea = FocusSidebar
	m.sidebarSection = SidebarProjects

	// Simulate selecting second project
	m.projectList.Select(1)

	// Create Enter key message
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}

	// Update should handle the selection
	newModel, _ := m.updateSidebar(keyMsg)
	m = newModel.(Model)

	// Current project should be updated
	if m.currentProject != "project-2" {
		t.Errorf("Expected current project to be 'project-2', got '%s'", m.currentProject)
	}
}

func TestModel_ToggleFocus(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	_ = tb // tb sets up environment variables

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	m := NewModel(cfg)
	m.focusedArea = FocusSidebar

	// Press Tab to toggle focus
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	newModel, _ := m.Update(keyMsg)
	m = newModel.(Model)

	if m.focusedArea != FocusContent {
		t.Errorf("Expected focus area to be FocusContent, got %v", m.focusedArea)
	}

	// Press Tab again to toggle back
	newModel, _ = m.Update(keyMsg)
	m = newModel.(Model)

	if m.focusedArea != FocusSidebar {
		t.Errorf("Expected focus area to be FocusSidebar, got %v", m.focusedArea)
	}
}

func TestModel_ViewSwitching(t *testing.T) {
	tb := testutil.SetupTestBrain(t)
	_ = tb // tb sets up environment variables

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	m := NewModel(cfg)

	tests := []struct {
		key          string
		expectedView ViewType
	}{
		{"1", ViewNotes},
		{"2", ViewTodosList},
		{"3", ViewTodosKanban},
	}

	for _, tt := range tests {
		t.Run("Key "+tt.key, func(t *testing.T) {
			keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			newModel, _ := m.Update(keyMsg)
			m = newModel.(Model)

			if m.activeView != tt.expectedView {
				t.Errorf("Expected active view %v, got %v", tt.expectedView, m.activeView)
			}
		})
	}
}
