package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/internal/update"
)

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if result, cmd, handled := m.handleGlobalKeys(msg); handled {
		return result, cmd
	}

	if m.showBatchConfigModal && m.batchConfig != nil && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
		r := msg.Runes[0]
		if r == '1' || r == '2' {
			optionIndex := int(r - '1')
			m.batchConfig.SetCurrentStepValue(optionIndex)
			if !m.batchConfig.NextStep() {
				m.showBatchConfigModal = false

				return m, m.startBatchInstall()
			}
		}

		return m, nil
	}

	if m.activePanel == PanelSearch && !msg.Alt && msg.Type == tea.KeyRunes {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel

		return m, cmd
	}

	if key.Matches(msg, m.keys.Escape) {
		return m.handleEscapeKey()
	}

	if result, cmd, handled := m.handleActionKeys(msg); handled {
		return result, cmd
	}

	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(msg)
		m.searchPanel = updatedPanel

		return m, cmd
	}

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

func (m *Model) handleGlobalKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit, true
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp

		return m, nil, true
	case key.Matches(msg, m.keys.Tab):
		m.NextPanel()

		return m, nil, true
	}

	return m, nil, false
}

func (m *Model) handleEscapeKey() (tea.Model, tea.Cmd) {
	if m.showHelp {
		m.showHelp = false

		return m, nil
	}
	if m.showInfoModal {
		m.showInfoModal = false

		return m, nil
	}
	if m.showUpdateModal {
		return m.closeUpdateModal()
	}
	if m.showInstallModal {
		return m.closeInstallModal()
	}
	if m.showBatchConfigModal {
		m.showBatchConfigModal = false
		m.batchConfig = nil

		return m, nil
	}
	if m.executeOutput != "" {
		m.executeOutput = ""
		m.err = nil

		return m, nil
	}

	if m.activePanel == PanelSearch {
		cmd, updatedPanel := m.searchPanel.Update(m.keys.Escape)
		m.searchPanel = updatedPanel

		return m, cmd
	}

	return m, nil
}

func (m *Model) handleActionKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	switch {
	case key.Matches(msg, m.keys.Update):
		return m.handleUpdateKey()
	case key.Matches(msg, m.keys.InfoModal):
		return m.handleInfoModalKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'i':
		return m.handleAltIKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'm':
		return m.handleAltMKey()
	case msg.Alt && msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 'r':
		return m, m.openRepositoryURL(), true
	}

	return m, nil, false
}

func (m *Model) handleUpdateKey() (tea.Model, tea.Cmd, bool) {
	m.showUpdateModal = true
	m.updating = true
	m.updateProgress = make(chan update.ProgressUpdate, 100)

	return m, tea.Batch(
		m.startUpdate(),
		m.tickSlugWave(),
	), true
}

func (m *Model) handleInfoModalKey() (tea.Model, tea.Cmd, bool) {
	if m.selectedTool != nil && m.activePanel != PanelSearch {
		m.showInfoModal = true
	}

	return m, nil, true
}
