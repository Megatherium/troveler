package tui

import (
	"fmt"

	"troveler/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderHelpModal() string {
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
		Width(min(60, m.width-4)).
		Render(helpText)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}

func (m *Model) renderInfoModal() string {
	if m.selectedTool == nil {
		content := styles.MutedStyle.Render("No tool selected\n\nNavigate to a tool first, then press 'i'")
		modalBox := styles.BorderStyle.
			BorderForeground(lipgloss.Color("#00FFFF")).
			Padding(1, 2).
			Width(min(60, m.width-4)).
			Render(content)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalBox)
	}

	var content string
	content += styles.TitleStyle.Render(m.selectedTool.Name)
	content += "\n\n"

	if m.selectedTool.Tagline != "" {
		content += styles.SubtitleStyle.Render(m.selectedTool.Tagline) + "\n\n"
	}

	if m.selectedTool.Description != "" {
		content += styles.HighlightStyle.Render("Description:") + "\n"
		content += m.selectedTool.Description + "\n\n"
	}

	content += styles.HighlightStyle.Render("Details:") + "\n"
	if m.selectedTool.Language != "" {
		content += styles.MutedStyle.Render("  Language: ") + m.selectedTool.Language + "\n"
	}
	if m.selectedTool.License != "" {
		content += styles.MutedStyle.Render("  License: ") + m.selectedTool.License + "\n"
	}
	if m.selectedTool.DatePublished != "" {
		content += styles.MutedStyle.Render("  Published: ") + m.selectedTool.DatePublished + "\n"
	}

	if len(m.installs) > 0 {
		content += fmt.Sprintf("\n💾 %d install options available\n", len(m.installs))
	}

	content += "\n" + styles.HelpStyle.Render("Press Esc to close")

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(80, m.width-4)).
		Height(min(30, m.height-4)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}

func (m *Model) renderUpdateModal() string {
	var content string

	if m.updating && m.updateSlugWave != nil {
		content = styles.TitleStyle.Render("🔄 Database Update") + "\n\n"
		content += m.updateSlugWave.RenderWithProgress() + "\n\n"
		content += styles.HelpStyle.Render("Press Esc to cancel")
	} else {
		content = styles.TitleStyle.Render("Database Update") + "\n\n"
		content += "Updating database...\n\n"
		content += styles.HelpStyle.Render("Press Esc to close")
	}

	modalBox := styles.BorderStyle.
		BorderForeground(lipgloss.Color("#00FFFF")).
		Padding(1, 2).
		Width(min(80, m.width-4)).
		Height(min(20, m.height-4)).
		Render(content)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modalBox,
	)
}
