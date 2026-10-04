package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/config"
	"troveler/db"
)

// Run starts the TUI application
func Run(database *db.SQLiteDB, cfg *config.Config) error {
	// Create model
	m := NewModel(database, cfg)
	if m.err != nil {
		return fmt.Errorf("invalid TUI config: %w", m.err)
	}

	// Create program with alt screen
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Even though we don't use it, helps with compatibility
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
