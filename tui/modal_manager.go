package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"troveler/db"
	"troveler/internal/update"
	"troveler/tui/styles"
)

// ModalManager owns the boolean flags for all modal dialogs and handles
// their rendering and escape-key dismissal.
type ModalManager struct {
	showHelp             bool
	showInfoModal        bool
	showUpdateModal      bool
	showInstallModal     bool
	showBatchConfigModal bool
}

// NewModalManager creates a ModalManager.
func NewModalManager() *ModalManager {
	return &ModalManager{}
}

// Active returns which modal type is currently shown.
type ModalType int

const (
	ModalNone ModalType = iota
	ModalHelp
	ModalInfo
	ModalUpdate
	ModalInstall
	ModalBatchConfig
)

// ActiveModalType returns the type of the currently active modal.
func (mm *ModalManager) ActiveModalType() ModalType {
	switch {
	case mm.showHelp:
		return ModalHelp
	case mm.showInfoModal:
		return ModalInfo
	case mm.showUpdateModal:
		return ModalUpdate
	case mm.showInstallModal:
		return ModalInstall
	case mm.showBatchConfigModal:
		return ModalBatchConfig
	}
	return ModalNone
}

// Show methods update the boolean state.
func (mm *ModalManager) ShowHelp()    { mm.showHelp = true }
func (mm *ModalManager) ShowInfo()    { mm.showInfoModal = true }
func (mm *ModalManager) ShowUpdate()  { mm.showUpdateModal = true }
func (mm *ModalManager) ShowInstall() { mm.showInstallModal = true }
func (mm *ModalManager) ShowBatchConfig() { mm.showBatchConfigModal = true }

// Getters for Model-level access (e.g. handleKeyPress checks these).
func (mm *ModalManager) IsHelpShown()          bool { return mm.showHelp }
func (mm *ModalManager) IsInfoShown()          bool { return mm.showInfoModal }
func (mm *ModalManager) IsUpdateShown()        bool { return mm.showUpdateModal }
func (mm *ModalManager) IsInstallShown()       bool { return mm.showInstallModal }
func (mm *ModalManager) IsBatchConfigShown()   bool { return mm.showBatchConfigModal }

// ToggleHelp toggles the help modal.
func (mm *ModalManager) ToggleHelp() {
	mm.showHelp = !mm.showHelp
}

// HandleEscape processes an Escape key press, closing the topmost modal.
// Returns the ModalType that was closed (ModalNone if no modal was active).
//
// closeUpdateFunc is called when closing the update modal to cancel the update.
// closeInstallAllowed returns true when the install modal may be closed.
func (mm *ModalManager) HandleEscape(
	closeUpdateFunc func(),  // cancels the update (nil if not updating)
	executing bool,           // true while an install is running
) (closed ModalType, cmd tea.Cmd) {
	if mm.showHelp {
		mm.showHelp = false
		return ModalHelp, nil
	}
	if mm.showInfoModal {
		mm.showInfoModal = false
		return ModalInfo, nil
	}
	if mm.showUpdateModal {
		mm.showUpdateModal = false
		if closeUpdateFunc != nil {
			closeUpdateFunc()
		}
		return ModalUpdate, nil
	}
	if mm.showInstallModal {
		if !executing {
			mm.showInstallModal = false
			return ModalInstall, nil
		}
		return ModalNone, nil // executing — cannot close
	}
	if mm.showBatchConfigModal {
		mm.showBatchConfigModal = false
		return ModalBatchConfig, nil
	}
	return ModalNone, nil
}

// CloseBatchConfig marks the batch config as done and clears its reference
// (the caller is responsible for clearing m.batchConfig).
func (mm *ModalManager) CloseBatchConfig() {
	mm.showBatchConfigModal = false
}

// --- View methods (rendering) ------------------------------------------------

// ViewHelp renders the keyboard shortcuts help overlay.
func (mm *ModalManager) ViewHelp(width, height int) string {
	helpText := `
Troveler TUI - Keyboard Shortcuts

Navigation:
  ↑/k, ↓/j     Move cursor up/down
  ←/h, →/l     Select column (in tools table)
  Tab          Cycle between panels
  Enter        Select tool / jump to install panel

Selection:
  m            Mark/unmark tool for batch install
  Alt+I        Install (single or batch if marked)
  Alt+M        Install via mise (single or batch)

Actions:
  Alt+R        Open repository URL in browser
  Alt+U        Update database
  Alt+S        Toggle sort order
  i            Show full info modal

Other:
  ?            Show/hide this help
  Esc          Cancel / close modal
  Alt+Q        Quit

Press Esc to close this help
`

	helpBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(60, width-4)).
		Render(helpText)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}

// ViewInfo renders the full-screen tool information modal.
func (mm *ModalManager) ViewInfo(width, height int, tool *db.Tool, installs []db.InstallInstruction) string {
	if tool == nil {
		content := styles.MutedStyle.Render("No tool selected\n\nNavigate to a tool first, then press 'i'")
		modalBox := styles.BorderStyle.
			BorderForeground(lipgloss.Color("#00FFFF")).
			Padding(1, 2).
			Width(min(60, width-4)).
			Render(content)

		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modalBox)
	}

	var content string
	content += styles.TitleStyle.Render(tool.Name)
	content += "\n\n"

	if tool.Tagline != "" {
		content += styles.SubtitleStyle.Render(tool.Tagline) + "\n\n"
	}

	if tool.Description != "" {
		content += styles.HighlightStyle.Render("Description:") + "\n"
		content += tool.Description + "\n\n"
	}

	content += styles.HighlightStyle.Render("Details:") + "\n"
	if tool.Language != "" {
		content += styles.MutedStyle.Render("  Language: ") + tool.Language + "\n"
	}
	if tool.License != "" {
		content += styles.MutedStyle.Render("  License: ") + tool.License + "\n"
	}
	if tool.DatePublished != "" {
		content += styles.MutedStyle.Render("  Published: ") + tool.DatePublished + "\n"
	}

	if len(installs) > 0 {
		content += fmt.Sprintf("\n%d install options available\n", len(installs))
	}

	content += "\n" + styles.HelpStyle.Render("Press Esc to close")

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(80, width-4)).
		Height(min(30, height-4)).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

// ViewUpdate renders the database update progress modal.
func (mm *ModalManager) ViewUpdate(width, height int, updating bool, slugWave *update.SlugWave) string {
	var content string
	if updating && slugWave != nil {
		content = styles.TitleStyle.Render("Database Update") + "\n\n"
		content += slugWave.RenderWithProgress() + "\n\n"
		content += styles.HelpStyle.Render("Press Esc to cancel")
	} else {
		content = styles.TitleStyle.Render("Database Update") + "\n\n"
		content += "Updating database...\n\n"
		content += styles.HelpStyle.Render("Press Esc to close")
	}

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(80, width-4)).
		Height(min(20, height-4)).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

// ViewInstall renders the install execution modal.
func (mm *ModalManager) ViewInstall(width, height int, executing bool, output string, err error, batchProgress *BatchInstallProgress) string {
	if batchProgress != nil {
		return mm.viewBatchInstall(width, height, batchProgress)
	}

	var content string

	if executing {
		content = styles.TitleStyle.Render("Executing Install Command") + "\n\n"
		content += "Running install command...\n\n"
		content += styles.MutedStyle.Render("This may take a moment depending on your package manager") + "\n\n"
		content += styles.HelpStyle.Render("Please wait...")
	} else {
		content = styles.TitleStyle.Render("Install Complete") + "\n\n"

		if err != nil {
			content += styles.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\n", err))
			content += styles.HighlightStyle.Render("Output:\n")
			content += styles.MutedStyle.Render(output)
		} else {
			content += styles.HighlightStyle.Render("Command executed successfully\n\n")
			if output != "" {
				content += styles.HighlightStyle.Render("Output:\n")
				content += styles.MutedStyle.Render(output)
			}
		}
		content += "\n\n" + styles.HelpStyle.Render("Press Esc to close")
	}

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(100, width-4)).
		Height(min(30, height-4)).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

// viewBatchInstall renders the batch install progress modal.
func (mm *ModalManager) viewBatchInstall(width, height int, bp *BatchInstallProgress) string {
	var content string

	if bp.IsComplete {
		content = styles.TitleStyle.Render("Batch Install Complete") + "\n\n"

		if len(bp.Completed) > 0 {
			content += styles.HighlightStyle.Render(fmt.Sprintf("Completed: %d\n", len(bp.Completed)))
		}
		if len(bp.Failed) > 0 {
			content += styles.ErrorStyle.Render(fmt.Sprintf("Failed: %d\n", len(bp.Failed)))
		}
		if len(bp.Skipped) > 0 {
			content += styles.MutedStyle.Render(fmt.Sprintf("Skipped: %d\n", len(bp.Skipped)))
		}

		content += "\n" + styles.HelpStyle.Render("Press Esc to close")
	} else {
		current := bp.CurrentIndex + 1
		total := len(bp.Tools)
		content = styles.TitleStyle.Render(fmt.Sprintf("Batch Install (%d/%d)", current, total)) + "\n\n"

		if bp.CurrentIndex < len(bp.Tools) {
			tool := bp.Tools[bp.CurrentIndex]
			content += styles.HighlightStyle.Render("Installing: ") + tool.Name + "\n\n"
		}

		content += styles.MutedStyle.Render("Please wait...") + "\n"

		progressWidth := 40
		completed := bp.CurrentIndex
		pct := float64(completed) / float64(total)
		filled := int(pct * float64(progressWidth))
		empty := progressWidth - filled
		bar := "[" + strings.Repeat("=", filled) + strings.Repeat(" ", empty) + "]"
		content += styles.SubtitleStyle.Render(bar) + "\n"
	}

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(70, width-4)).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

// ViewBatchConfig renders the batch install configuration wizard modal.
func (mm *ModalManager) ViewBatchConfig(width, height int, config *BatchInstallConfig, markedCount int) string {
	if config == nil {
		return ""
	}

	var content string
	content += styles.TitleStyle.Render(fmt.Sprintf("Batch Install Configuration (%d tools)", markedCount)) + "\n\n"

	stepNum := config.ConfigStep + 1
	totalSteps := config.ConfigStepCount()
	content += styles.MutedStyle.Render(fmt.Sprintf("Step %d of %d", stepNum, totalSteps)) + "\n\n"

	content += styles.HighlightStyle.Render(config.GetCurrentStepTitle()) + "\n\n"

	options := config.GetCurrentStepOptions()
	for i, opt := range options {
		key := fmt.Sprintf("[%d]", i+1)
		content += styles.SubtitleStyle.Render(key) + " " + opt + "\n"
	}

	content += "\n" + styles.HelpStyle.Render("Press 1 or 2 to select | Esc to cancel")

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(70, width-4)).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}
