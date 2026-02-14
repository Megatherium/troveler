package tui

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/internal/platform"
	"troveler/tui/panels"
)

type batchInstallStartMsg struct {
	tools []db.SearchResult
}

type batchInstallProgressMsg struct {
	toolID  string
	output  string
	err     error
	skipped bool
}

type batchInstallCompleteMsg struct{}

func (m *Model) startBatchInstall() tea.Cmd {
	markedTools := m.toolsPanel.GetMarkedTools()
	if len(markedTools) == 0 {
		return nil
	}

	m.batchProgress = NewBatchInstallProgress(markedTools)
	m.showInstallModal = true
	m.executing = true

	return func() tea.Msg {
		return batchInstallStartMsg{tools: markedTools}
	}
}

func (m *Model) handleAltIKey() (tea.Model, tea.Cmd, bool) {
	if m.toolsPanel.GetMarkedCount() > 0 {
		m.batchConfig = NewBatchInstallConfig()
		m.showBatchConfigModal = true

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
		m.batchConfig = NewBatchInstallConfig()
		m.batchConfig.UseMise = true
		m.showBatchConfigModal = true

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

func (m *Model) processBatchTool(index int) tea.Cmd {
	if m.batchProgress == nil || index >= len(m.batchProgress.Tools) {
		return func() tea.Msg { return batchInstallCompleteMsg{} }
	}

	tool := m.batchProgress.Tools[index]
	config := m.batchConfig

	return func() tea.Msg {
		installs, err := m.db.GetInstallInstructions(tool.ID)
		if err != nil || len(installs) == 0 {
			if config != nil && config.SkipIfBlind {
				return batchInstallProgressMsg{
					toolID:  tool.ID,
					skipped: true,
				}
			}

			return batchInstallProgressMsg{
				toolID: tool.ID,
				err:    fmt.Errorf("no install instructions found"),
			}
		}

		selector := install.NewPlatformSelector("", "", "", tool.Language)
		osInfo, _ := platform.DetectOS()
		detectedOS := ""
		if osInfo != nil {
			detectedOS = osInfo.ID
		}
		platform := selector.SelectPlatform(detectedOS)

		filtered, _ := install.FilterCommands(installs, platform, tool.Language)
		if len(filtered) == 0 {
			if config != nil && config.SkipIfBlind {
				return batchInstallProgressMsg{
					toolID:  tool.ID,
					skipped: true,
				}
			}

			return batchInstallProgressMsg{
				toolID: tool.ID,
				err:    fmt.Errorf("no compatible install method"),
			}
		}

		defaultCmd := install.SelectDefaultCommand(filtered, false, detectedOS)
		cmd := filtered[0].Command
		if defaultCmd != nil {
			cmd = defaultCmd.Command
		}

		if config != nil && config.UseMise {
			cmd = install.TransformToMise(cmd)
		}

		if config != nil && config.UseSudo {
			isSystemPM := isSystemPackageManager(filtered[0].Platform)
			if !config.SudoOnlySystem || isSystemPM {
				cmd = "sudo " + cmd
			}
		}

		execCmd := exec.Command("sh", "-c", cmd) //nolint:noctx //nolint:gosec // G204: user install
		output, err := execCmd.CombinedOutput()

		return batchInstallProgressMsg{
			toolID: tool.ID,
			output: string(output),
			err:    err,
		}
	}
}

func isSystemPackageManager(platform string) bool {
	switch platform {
	case "apt", "dnf", "yum", "pacman", "apk", "zypper", "nix":
		return true
	default:
		return false
	}
}
