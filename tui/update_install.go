package tui

import (
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/internal/update"
)

func (m *Model) executeInstallCommand(command string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("sh", "-c", command) //nolint:noctx,gosec // G204: user install command
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
			cmd = exec.Command("open", url) //nolint:noctx,gosec // G204: user URL open
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) //nolint:noctx,gosec // G204: user URL open
		default:
			cmd = exec.Command("xdg-open", url) //nolint:noctx,gosec // G204: user URL open
		}

		_ = cmd.Start()

		return nil
	}
}

type updateProgressMsg update.ProgressUpdate
