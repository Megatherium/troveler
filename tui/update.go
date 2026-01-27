package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.MouseMsg:
		// Mouse support disabled per spec
		return m, nil
	}

	// Forward to active panel
	if panel := m.GetActivePanel(); panel != nil {
		updatedPanel, cmd := panel.Update(msg)
		m.panels[m.activePanel] = updatedPanel
		return m, cmd
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keybindings (work from any panel)
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil

	case key.Matches(msg, m.keys.Escape):
		// Close modals if open
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if m.showInfoModal {
			m.showInfoModal = false
			return m, nil
		}
		if m.showUpdateModal {
			m.showUpdateModal = false
			// TODO: Cancel update
			return m, nil
		}
		return m, nil

	case key.Matches(msg, m.keys.Tab):
		m.NextPanel()
		return m, nil

	case key.Matches(msg, m.keys.Update):
		m.showUpdateModal = true
		// TODO: Start update process
		return m, nil

	case key.Matches(msg, m.keys.InfoModal):
		if m.selectedTool != nil {
			m.showInfoModal = true
		}
		return m, nil
	}

	// Forward to active panel if no global key matched
	if panel := m.GetActivePanel(); panel != nil {
		updatedPanel, cmd := panel.Update(msg)
		m.panels[m.activePanel] = updatedPanel
		return m, cmd
	}

	return m, nil
}
