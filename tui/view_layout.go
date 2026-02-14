package tui

import (
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) renderMainLayout() string {
	statusBarHeight := 2
	topMargin := 1
	availableHeight := m.height - statusBarHeight - topMargin

	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth

	searchHeight := max(5, availableHeight/10)
	toolsHeight := availableHeight - searchHeight

	installLineCount := len(m.installs) + 4
	if installLineCount < 8 {
		installLineCount = 8
	}

	maxInstallHeight := (availableHeight * 2) / 3
	installHeight := installLineCount
	if installHeight > maxInstallHeight {
		installHeight = maxInstallHeight
	}

	infoHeight := availableHeight - installHeight

	searchPanel := m.renderSearchPanel(leftWidth, searchHeight)
	toolsPanel := m.renderToolsPanel(leftWidth, toolsHeight)
	infoPanel := m.renderInfoPanel(rightWidth, infoHeight)
	installPanel := m.renderInstallPanel(rightWidth, installHeight)

	leftColumn := lipgloss.JoinVertical(lipgloss.Left, searchPanel, toolsPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, infoPanel, installPanel)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	topSpace := "\n"
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, topSpace, layout, statusBar)
}
