package tui

import (
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
		return m.handleSearchTriggered(msg)

	case searchResultMsg:
		return m.handleSearchResult(msg)

	case searchErrorMsg:
		return m.handleSearchError(msg)

	case panels.ToolMarkedMsg:
		return m, nil

	case panels.ToolCursorChangedMsg:
		return m.handleToolCursorChanged(msg)

	case panels.ToolSelectedMsg:
		return m.handleToolSelected()

	case panels.InstallExecuteMsg:
		return m.handleInstallExecute(msg)

	case panels.InstallExecuteMiseMsg:
		return m.handleInstallExecuteMise(msg)

	case installCompleteMsg:
		return m.handleInstallComplete(msg)

	case batchInstallStartMsg:
		return m, m.processBatchTool(0)

	case batchInstallProgressMsg:
		return m.handleBatchInstallProgress(msg)

	case batchInstallCompleteMsg:
		return m.handleBatchInstallComplete()

	case updateProgressMsg:
		return m.handleUpdateProgress(msg)

	case slugTickMsg:
		return m.handleSlugTick()

	case tea.MouseMsg:
		return m, nil
	}

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

type installCompleteMsg struct {
	output string
	err    error
}

type slugTickMsg struct{}
