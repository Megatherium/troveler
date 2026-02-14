package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/update"
	"troveler/tui/panels"
)

func (m *Model) handleSearchTriggered(msg panels.SearchTriggeredMsg) (tea.Model, tea.Cmd) {
	m.searching = true

	return m, m.performSearch(msg.Query)
}

func (m *Model) handleSearchResult(msg searchResultMsg) (tea.Model, tea.Cmd) {
	m.tools = msg.tools
	m.toolsPanel.SetTools(msg.tools)
	m.searching = false

	m.toolsPanel.UpdateAllInstalledStatus(m.db.GetInstallInstructions)

	if len(msg.tools) > 0 {
		firstTool := &msg.tools[0].Tool
		m.selectedTool = firstTool
		m.infoPanel.SetTool(firstTool, []db.InstallInstruction{})

		installs, err := m.db.GetInstallInstructions(firstTool.ID)
		if err == nil {
			m.installs = installs
			m.infoPanel.SetTool(firstTool, installs)
			m.installPanel.SetTool(firstTool, installs)
		}
	}

	return m, nil
}

func (m *Model) handleSearchError(msg searchErrorMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	m.searching = false

	return m, nil
}

func (m *Model) handleToolCursorChanged(msg panels.ToolCursorChangedMsg) (tea.Model, tea.Cmd) {
	m.selectedTool = &msg.Tool.Tool

	m.infoPanel.SetTool(m.selectedTool, []db.InstallInstruction{})

	installs, err := m.db.GetInstallInstructions(m.selectedTool.ID)
	if err == nil {
		m.installs = installs
		m.infoPanel.SetTool(m.selectedTool, installs)
		m.installPanel.SetTool(m.selectedTool, installs)
	}

	return m, nil
}

func (m *Model) handleToolSelected() (tea.Model, tea.Cmd) {
	m.toolsPanel.Blur()
	m.activePanel = PanelInstall
	m.installPanel.Focus()

	return m, nil
}

func (m *Model) handleInstallExecute(msg panels.InstallExecuteMsg) (tea.Model, tea.Cmd) {
	m.showInstallModal = true
	m.executing = true
	m.executeOutput = ""

	return m, m.executeInstallCommand(msg.Command)
}

func (m *Model) handleInstallExecuteMise(msg panels.InstallExecuteMiseMsg) (tea.Model, tea.Cmd) {
	m.showInstallModal = true
	m.executing = true
	m.executeOutput = ""

	return m, m.executeInstallCommand(msg.Command)
}

func (m *Model) handleInstallComplete(msg installCompleteMsg) (tea.Model, tea.Cmd) {
	m.executing = false
	m.executeOutput = msg.output
	if msg.err != nil {
		m.err = msg.err
	}

	return m, nil
}

func (m *Model) handleBatchInstallProgress(msg batchInstallProgressMsg) (tea.Model, tea.Cmd) {
	if m.batchProgress == nil {
		return m, nil
	}

	if msg.skipped {
		m.batchProgress.Skipped = append(m.batchProgress.Skipped, msg.toolID)
	} else if msg.err != nil {
		m.batchProgress.Failed = append(m.batchProgress.Failed, msg.toolID)
	} else {
		m.batchProgress.Completed = append(m.batchProgress.Completed, msg.toolID)
	}
	m.batchProgress.CurrentOutput = msg.output
	m.batchProgress.CurrentError = msg.err
	m.batchProgress.CurrentIndex++

	if m.batchProgress.CurrentIndex < len(m.batchProgress.Tools) {
		return m, m.processBatchTool(m.batchProgress.CurrentIndex)
	}
	m.batchProgress.IsComplete = true
	m.executing = false
	m.toolsPanel.ClearMarks()

	return m, nil
}

func (m *Model) handleBatchInstallComplete() (tea.Model, tea.Cmd) {
	m.executing = false
	if m.batchProgress != nil {
		m.batchProgress.IsComplete = true
	}

	return m, nil
}

func (m *Model) handleUpdateProgress(msg updateProgressMsg) (tea.Model, tea.Cmd) {
	upd := update.ProgressUpdate(msg)

	if upd.Type == "progress" && upd.Total > 0 && m.updateSlugWave == nil {
		m.updateSlugWave = update.NewSlugWave(upd.Total)
	}

	if upd.Type == "slug" && m.updateSlugWave != nil {
		m.updateSlugWave.AddSlug(upd.Slug)
		m.updateSlugWave.IncProcessed()
	}

	if upd.Type == "complete" {
		m.updating = false

		return m, nil
	}

	if upd.Type == "error" {
		m.updating = false
		m.err = upd.Error

		return m, nil
	}

	return m, m.listenForUpdates()
}

func (m *Model) handleSlugTick() (tea.Model, tea.Cmd) {
	if m.updating && m.updateSlugWave != nil {
		m.updateSlugWave.AdvanceFrame()

		return m, m.tickSlugWave()
	}

	return m, nil
}
