package tui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/internal/update"
)

func (m *Model) executeInstallCommand(command string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("sh", "-c", command) //nolint:noctx //nolint:gosec // G204
		output, err := cmd.CombinedOutput()

		return installCompleteMsg{
			output: string(output),
			err:    err,
		}
	}
}

func (m *Model) openRepositoryURL() tea.Cmd {
	return func() tea.Msg {
		if m.selectedTool == nil || m.selectedTool.CodeRepository == "" {
			return nil
		}

		var cmd *exec.Cmd
		url := m.selectedTool.CodeRepository

		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url) //nolint:noctx //nolint:gosec // G204
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) //nolint:noctx //nolint:gosec // G204
		default:
			cmd = exec.Command("xdg-open", url) //nolint:noctx //nolint:gosec // G204
		}

		_ = cmd.Start()

		return nil
	}
}

func (m *Model) closeUpdateModal() (tea.Model, tea.Cmd) {
	m.showUpdateModal = false
	m.updating = false
	if m.updateCancel != nil {
		m.updateCancel()
		m.updateCancel = nil
	}
	m.updateProgress = nil

	return m, nil
}

func (m *Model) closeInstallModal() (tea.Model, tea.Cmd) {
	if !m.executing {
		m.showInstallModal = false
		m.executeOutput = ""
		m.err = nil
		m.batchProgress = nil
		m.batchConfig = nil
	}

	return m, nil
}

func (m *Model) startUpdate() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.updateCancel = cancel

	opts := update.UpdateOptions{
		Limit:    0,
		Progress: m.updateProgress,
	}

	go func() {
		_ = m.updateService.FetchAndUpdate(ctx, opts)
	}()

	return m.listenForUpdates()
}

func (m *Model) listenForUpdates() tea.Cmd {
	return func() tea.Msg {
		if m.updateProgress == nil {
			return nil
		}
		upd, ok := <-m.updateProgress
		if !ok {
			return updateProgressMsg{Type: "complete"}
		}

		return updateProgressMsg(upd)
	}
}

func (m *Model) tickSlugWave() tea.Cmd {
	return tea.Tick(time.Millisecond*33, func(_ time.Time) tea.Msg {
		return slugTickMsg{}
	})
}

type updateProgressMsg update.ProgressUpdate
