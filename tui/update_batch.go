package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"troveler/internal/install"
	"troveler/tui/panels"
)

func (m *Model) startBatchInstall() tea.Cmd {
	markedTools := m.toolsPanel.GetMarkedTools()
	if len(markedTools) == 0 {
		return nil
	}

	m.executing = true
	m.modals.ShowInstall()
	return m.batch.StartInstall(markedTools)
}

func (m *Model) handleAltIKey() (tea.Model, tea.Cmd, bool) {
	if m.toolsPanel.GetMarkedCount() > 0 {
		m.batch.StartBatchConfig(false)
		m.modals.ShowBatchConfig()

		return m, nil, true
	}
	if m.installPanel.HasCommands() {
		cmd := m.installPanel.GetSelectedCommand()
		if cmd != "" {
			return m, func() tea.Msg {
				return panels.InstallExecuteMsg{Command: cmd}
			}, true
		}
	}

	return m, nil, true
}

func (m *Model) handleAltMKey() (tea.Model, tea.Cmd, bool) {
	if m.toolsPanel.GetMarkedCount() > 0 {
		m.batch.StartBatchConfig(true)
		m.modals.ShowBatchConfig()

		return m, nil, true
	}
	if m.installPanel.HasCommands() {
		cmd := m.installPanel.GetSelectedCommand()
		if cmd != "" {
			transformedCmd := install.TransformToMise(cmd)

			return m, func() tea.Msg {
				return panels.InstallExecuteMiseMsg{Command: transformedCmd}
			}, true
		}
	}

	return m, nil, true
}

func isSystemPackageManager(platform string) bool {
	switch platform {
	case "apt", "dnf", "yum", "pacman", "apk", "zypper", "nix":
		return true
	default:
		return false
	}
}
