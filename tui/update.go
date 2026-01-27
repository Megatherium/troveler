package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
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
		
		// Update info panel
		m.infoPanel.SetTool(m.selectedTool, []db.InstallInstruction{})
		
		// Load install instructions
		installs, err := m.db.GetInstallInstructions(m.selectedTool.ID)
		if err == nil {
			m.installs = installs
			m.infoPanel.SetTool(m.selectedTool, installs)
			m.installPanel.SetTool(m.selectedTool, installs)
		}
		
		// Jump to install panel
		m.toolsPanel.Blur()
		m.activePanel = PanelInstall
		m.installPanel.Focus()
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
	case PanelInstall:
		cmd := m.installPanel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleKeyPress processes keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Check critical global keys first (help, quit, etc.)
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		return m, nil
	case key.Matches(msg, m.keys.Tab):
		m.NextPanel()
		return m, nil
	}

	// CRITICAL: Forward regular keys to search panel FIRST when it's focused
	// This allows typing characters like 'i', 'u', etc. in the search box
	if m.activePanel == PanelSearch && !msg.Alt && msg.Type == tea.KeyRunes {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		return m, cmd
	}

	// Other global keybindings
	switch {
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
		// Also forward ESC to active panel
		if m.activePanel == PanelSearch {
			cmd, updatedPanel := m.searchPanel.Update(msg)
			m.searchPanel = updatedPanel
			return m, cmd
		}
		return m, nil

	case key.Matches(msg, m.keys.Update):
		m.showUpdateModal = true
		// TODO: Start update process
		return m, nil

	case key.Matches(msg, m.keys.InfoModal):
		// Only show modal if NOT typing in search
		if m.selectedTool != nil && m.activePanel != PanelSearch {
			m.showInfoModal = true
		}
		return m, nil
	}

	// Forward to search panel for non-rune keys (Enter, Backspace, etc.)
	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel
		return m, cmd
	}

	// Forward to other panels
	switch m.activePanel {
	case PanelTools:
		cmd := m.toolsPanel.Update(msg)
		return m, cmd
	case PanelInstall:
		cmd := m.installPanel.Update(msg)
		return m, cmd
	}

	return m, nil
}
