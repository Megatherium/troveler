package tui

import (
	"os/exec"

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
		
		// Auto-select first tool to populate info/install panels
		if len(msg.tools) > 0 {
			firstTool := &msg.tools[0].Tool
			m.selectedTool = firstTool
			m.infoPanel.SetTool(firstTool, []db.InstallInstruction{})
			
			// Load install instructions for first tool
			installs, err := m.db.GetInstallInstructions(firstTool.ID)
			if err == nil {
				m.installs = installs
				m.infoPanel.SetTool(firstTool, installs)
				m.installPanel.SetTool(firstTool, installs)
			}
		}
		return m, nil

	case searchErrorMsg:
		// Search error
		m.err = msg.err
		m.searching = false
		return m, nil

	case panels.ToolCursorChangedMsg:
		// Cursor moved to a different tool - update info/install panels
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
		return m, nil

	case panels.ToolSelectedMsg:
		// Tool was selected (Enter pressed in tools panel)
		// Info/install panels already populated by cursor change, just jump to install panel
		m.toolsPanel.Blur()
		m.activePanel = PanelInstall
		m.installPanel.Focus()
		return m, nil

	case panels.InstallExecuteMsg:
		// User wants to execute install command
		m.executing = true
		m.executeOutput = ""
		return m, m.executeInstallCommand(msg.Command)

	case installCompleteMsg:
		// Install finished
		m.executing = false
		m.executeOutput = msg.output
		if msg.err != nil {
			m.err = msg.err
		}
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
		if m.executeOutput != "" {
			m.executeOutput = ""
			m.err = nil
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
		// TODO: Wire up actual update - for now just show placeholder
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

// installCompleteMsg is sent when install completes
type installCompleteMsg struct {
	output string
	err    error
}

// executeInstallCommand runs the install command
func (m *Model) executeInstallCommand(command string) tea.Cmd {
	return func() tea.Msg {
		// Execute command using shell
		cmd := exec.Command("sh", "-c", command)
		output, err := cmd.CombinedOutput()

		return installCompleteMsg{
			output: string(output),
			err:    err,
		}
	}
}


