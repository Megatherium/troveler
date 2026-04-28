package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/internal/update"
	"troveler/tui/panels"
)

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Layer 1: Global keys (quit, help, tab)
	if result, cmd, handled := m.handleGlobalKeys(msg); handled {
		return result, cmd
	}

	// Layer 2: Batch config modal input
	if m.showBatchConfigModal && m.batchConfig != nil {
		return m.handleBatchConfigInput(msg)
	}

	// Layer 3: Search panel text input (unmodified rune keys)
	if m.activePanel == PanelSearch && !msg.Alt && msg.Type == tea.KeyRunes {
		return m.delegateToSearchPanel(msg)
	}

	// Layer 4: Escape key
	if key.Matches(msg, m.keys.Escape) {
		return m.handleEscapeKey()
	}

	// Layer 5: Action keys (alt+u, i, alt+i, alt+m, alt+r)
	if result, cmd, handled := m.handleActionKeys(msg); handled {
		return result, cmd
	}

	// Layer 6: Panel delegation
	return m.delegateToActivePanel(msg)
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
		newModel, cmd := m.searchPanel.Update(m.keys.Escape)
		if p, ok := newModel.(*panels.SearchPanel); ok {
			m.searchPanel = p
		}

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

// delegateToActivePanel forwards a message to the currently focused panel.
func (m *Model) delegateToActivePanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.activePanel {
	case PanelSearch:
		return m.delegateToSearchPanel(msg)
	case PanelTools:
		newModel, cmd := m.toolsPanel.Update(msg)
		if p, ok := newModel.(*panels.ToolsPanel); ok {
			m.toolsPanel = p
		}

		return m, cmd
	case PanelInstall:
		newModel, cmd := m.installPanel.Update(msg)
		if p, ok := newModel.(*panels.InstallPanel); ok {
			m.installPanel = p
		}

		return m, cmd
	}

	return m, nil
}

// delegateToSearchPanel forwards a message to the search panel.
func (m *Model) delegateToSearchPanel(msg tea.Msg) (tea.Model, tea.Cmd) {
	newModel, cmd := m.searchPanel.Update(msg)
	if p, ok := newModel.(*panels.SearchPanel); ok {
		m.searchPanel = p
	}

	return m, cmd
}

// handleBatchConfigInput processes key input when the batch config modal is open.
func (m *Model) handleBatchConfigInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) == 0 {
		return m, nil
	}

	r := msg.Runes[0]
	if r != '1' && r != '2' {
		return m, nil
	}

	optionIndex := int(r - '1')
	m.batchConfig.SetCurrentStepValue(optionIndex)
	if !m.batchConfig.NextStep() {
		m.showBatchConfigModal = false

		return m, m.startBatchInstall()
	}

	return m, nil
}
