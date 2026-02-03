package tui

import (
	"testing"

	"github.com/sandermoonemans/local-brain/pkg/config"
)

func TestModel_ShouldIncludeCompleted(t *testing.T) {
	tests := []struct {
		name           string
		activeView     ViewType
		showCompleted  bool
		expectedResult bool
	}{
		{
			name:           "Kanban view always includes completed",
			activeView:     ViewTodosKanban,
			showCompleted:  false,
			expectedResult: true,
		},
		{
			name:           "Kanban view with showCompleted true",
			activeView:     ViewTodosKanban,
			showCompleted:  true,
			expectedResult: true,
		},
		{
			name:           "List view with showCompleted false",
			activeView:     ViewTodosList,
			showCompleted:  false,
			expectedResult: false,
		},
		{
			name:           "List view with showCompleted true",
			activeView:     ViewTodosList,
			showCompleted:  true,
			expectedResult: true,
		},
		{
			name:           "Notes view with showCompleted false",
			activeView:     ViewNotes,
			showCompleted:  false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal config for testing
			cfg := &config.Config{}
			m := NewModel(cfg)
			m.activeView = tt.activeView
			m.showCompleted = tt.showCompleted

			result := m.shouldIncludeCompleted()

			if result != tt.expectedResult {
				t.Errorf("shouldIncludeCompleted() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestModel_GetViewName(t *testing.T) {
	tests := []struct {
		activeView   ViewType
		expectedName string
	}{
		{ViewNotes, "Notes"},
		{ViewTodosList, "Todos (List)"},
		{ViewTodosKanban, "Todos (Kanban)"},
	}

	cfg := &config.Config{}
	m := NewModel(cfg)

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			m.activeView = tt.activeView
			result := m.getViewName()

			if result != tt.expectedName {
				t.Errorf("getViewName() = %v, want %v", result, tt.expectedName)
			}
		})
	}
}
