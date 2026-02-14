package tui

import (
	"fmt"

	"troveler/tui/styles"

	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderSearchPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelSearch {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
		m.searchPanel.SetStyles()
	}

	title := titleStyle.Render(" Search ")

	if m.searching {
		title += styles.MutedStyle.Render(" [searching...]")
	}

	content := m.searchPanel.View(width-4, height-4)

	return borderStyle.
		Width(width - 1).
		Height(height - 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *Model) renderToolsPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelTools {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
	}

	var title string
	if m.searching {
		title = titleStyle.Render(" Tools (searching...) ")
	} else {
		title = titleStyle.Render(fmt.Sprintf(" Tools (%d) ", len(m.tools)))
	}

	content := m.toolsPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *Model) renderInfoPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelInfo {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
	}

	title := " Info "
	if m.selectedTool != nil {
		title = " " + m.selectedTool.Name + " "
	}
	titleRendered := titleStyle.Render(title)

	content := m.infoPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, titleRendered, content))
}

func (m *Model) renderInstallPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelInstall {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
	}

	title := titleStyle.Render(" Install ")

	content := m.installPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

func (m *Model) renderStatusBar() string {
	help := styles.HelpStyle.Render(
		"Tab: panels | m: mark | Alt+I: install | Alt+M: mise | Alt+R: repo | Alt+U: update | Alt+Q: quit | ?: help")

	if m.err != nil {
		errMsg := styles.ErrorStyle.Render(fmt.Sprintf(" Error: %v ", m.err))

		return lipgloss.JoinHorizontal(lipgloss.Left, errMsg, " | ", help)
	}

	var statusParts []string

	if len(m.tools) > 0 {
		statusParts = append(statusParts, fmt.Sprintf("%d tools", len(m.tools)))
	}

	markedCount := m.toolsPanel.GetMarkedCount()
	if markedCount > 0 {
		statusParts = append(statusParts, styles.HighlightStyle.Render(fmt.Sprintf("%d marked", markedCount)))
	}

	if len(statusParts) > 0 {
		status := styles.StatusBarStyle.Render(" " + statusParts[0] + " ")
		for i := 1; i < len(statusParts); i++ {
			status = lipgloss.JoinHorizontal(lipgloss.Left, status, " | ", statusParts[i])
		}

		return lipgloss.JoinHorizontal(lipgloss.Left, status, " | ", help)
	}

	return help
}
