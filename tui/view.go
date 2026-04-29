package tui

import (
	"fmt"
)

// View renders the entire TUI layout.
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

	switch m.modals.ActiveModalType() {
	case ModalNone:
		// No modal active — render main layout
	case ModalHelp:
		return m.modals.ViewHelp(m.width, m.height)
	case ModalInfo:
		return m.modals.ViewInfo(m.width, m.height, m.selectedTool, m.installs)
	case ModalUpdate:
		return m.modals.ViewUpdate(m.width, m.height, m.update.IsRunning(), m.update.SlugWave())
	case ModalInstall:
		return m.modals.ViewInstall(m.width, m.height, m.executing, m.executeOutput, m.err, m.batch.Progress())
	case ModalBatchConfig:
		return m.modals.ViewBatchConfig(m.width, m.height, m.batch.Config(), m.toolsPanel.GetMarkedCount())
	}

	return m.renderMainLayout()
}
