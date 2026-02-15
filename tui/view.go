package tui

import (
	"fmt"
)

// View renders the full TUI layout.
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	minWidth := 80
	if m.width < minWidth {
		return fmt.Sprintf("Terminal too narrow (%d columns).\nPlease resize to at least %d columns.\n", m.width, minWidth)
	}

	minHeight := 20
	if m.height < minHeight {
		return fmt.Sprintf("Terminal too short (%d rows).\nPlease resize to at least %d rows.\n", m.height, minHeight)
	}

	if m.showHelp {
		return m.renderHelpModal()
	}

	if m.showInfoModal {
		return m.renderInfoModal()
	}

	if m.showUpdateModal {
		return m.renderUpdateModal()
	}

	if m.showInstallModal {
		return m.renderInstallModal()
	}

	if m.showBatchConfigModal {
		return m.renderBatchConfigModal()
	}

	return m.renderMainLayout()
}
