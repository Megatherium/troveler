package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/tui/panels"
)

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case panels.SearchTriggeredMsg:
		// Search was triggered (debounced or Enter)
		m.searching = true
		return m, m.performSearch(msg.Query)

	case searchResultMsg:
		// Search results received
		m.tools = msg.tools
		m.toolsPanel.SetTools(msg.tools)
		m.searching = false
		return m, nil

	case searchErrorMsg:
		// Search error
		m.err = msg.err
		m.searching = false
		return m, nil

	case panels.ToolSelectedMsg:
		// Tool was selected (Enter pressed in tools panel)
		m.selectedTool = &msg.Tool.Tool
		// Jump to install panel
		m.toolsPanel.Blur()
		m.activePanel = PanelInstall
		return m, nil

	case tea.MouseMsg:
		// Mouse support disabled per spec
		return m, nil
	}

	// Forward to active panel
	switch m.activePanel {
	case PanelSearch:
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		cmds = append(cmds, cmd)
	case PanelTools:
		cmd := m.toolsPanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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

	// Forward to search panel if active and no global key matched
	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		return m, cmd
	}

	// TODO: Forward to other panels when implemented

	return m, nil
}
