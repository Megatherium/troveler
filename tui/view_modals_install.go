package tui

import (
	"fmt"
	"strings"

	"troveler/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderInstallModal() string {
	var content string

	if m.batchProgress != nil {
		return m.renderBatchInstallModal()
	}

	if m.executing {
		content = styles.TitleStyle.Render("💻 Executing Install Command") + "\n\n"
		content += "Running install command...\n\n"
		content += styles.MutedStyle.Render("This may take a moment depending on your package manager") + "\n\n"
		content += styles.HelpStyle.Render("Please wait...")
	} else {
		content = styles.TitleStyle.Render("💻 Install Complete") + "\n\n"

		if m.err != nil {
			content += styles.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.err))
			content += styles.HighlightStyle.Render("Output:\n")
			content += styles.MutedStyle.Render(m.executeOutput)
		} else {
			content += styles.HighlightStyle.Render("✓ Command executed successfully\n\n")
			if m.executeOutput != "" {
				content += styles.HighlightStyle.Render("Output:\n")
				content += styles.MutedStyle.Render(m.executeOutput)
			}
		}
		content += "\n\n" + styles.HelpStyle.Render("Press Esc to close")
	}

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(100, m.width-4)).
		Height(min(30, m.height-4)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

func (m *Model) renderBatchInstallModal() string {
	var content string
	bp := m.batchProgress

	if bp.IsComplete {
		content = styles.TitleStyle.Render("🔧 Batch Install Complete") + "\n\n"

		if len(bp.Completed) > 0 {
			content += styles.HighlightStyle.Render(fmt.Sprintf("✓ Completed: %d\n", len(bp.Completed)))
		}
		if len(bp.Failed) > 0 {
			content += styles.ErrorStyle.Render(fmt.Sprintf("✗ Failed: %d\n", len(bp.Failed)))
		}
		if len(bp.Skipped) > 0 {
			content += styles.MutedStyle.Render(fmt.Sprintf("○ Skipped: %d\n", len(bp.Skipped)))
		}

		content += "\n" + styles.HelpStyle.Render("Press Esc to close")
	} else {
		current := bp.CurrentIndex + 1
		total := len(bp.Tools)
		content = styles.TitleStyle.Render(fmt.Sprintf("🔧 Batch Install (%d/%d)", current, total)) + "\n\n"

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
		Width(min(70, m.width-4)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

func (m *Model) renderBatchConfigModal() string {
	if m.batchConfig == nil {
		return ""
	}

	var content string
	markedCount := m.toolsPanel.GetMarkedCount()
	content += styles.TitleStyle.Render(fmt.Sprintf("🔧 Batch Install Configuration (%d tools)", markedCount)) + "\n\n"

	stepNum := m.batchConfig.ConfigStep + 1
	totalSteps := m.batchConfig.ConfigStepCount()
	content += styles.MutedStyle.Render(fmt.Sprintf("Step %d of %d", stepNum, totalSteps)) + "\n\n"

	content += styles.HighlightStyle.Render(m.batchConfig.GetCurrentStepTitle()) + "\n\n"

	options := m.batchConfig.GetCurrentStepOptions()
	for i, opt := range options {
		key := fmt.Sprintf("[%d]", i+1)
		content += styles.SubtitleStyle.Render(key) + " " + opt + "\n"
	}

	content += "\n" + styles.HelpStyle.Render("Press 1 or 2 to select | Esc to cancel")

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(70, m.width-4)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}
