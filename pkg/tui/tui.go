package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sandermoonemans/local-brain/pkg/config"
)

// Launch starts the TUI application
func Launch() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	m := NewModel(cfg)

	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	_, err = p.Run()
	return err
}
