package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"troveler/tui/styles"
)

// View renders the TUI
func (m *Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	// Show modals if active
	if m.showHelp {
		return m.renderHelpModal()
	}

	if m.showInfoModal {
		return m.renderInfoModal()
	}

	if m.showUpdateModal {
		return m.renderUpdateModal()
	}

	if m.showInstallModal {
		return m.renderInstallModal()
	}

	// Render main layout
	return m.renderMainLayout()
}

// renderMainLayout renders the 4-panel layout
func (m *Model) renderMainLayout() string {
	// Reserve space for status bar (2 lines) and top margin (1 line)
	statusBarHeight := 2
	topMargin := 1
	availableHeight := m.height - statusBarHeight - topMargin

	// Calculate panel dimensions
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth

	// Left side: fixed proportions
	searchHeight := max(5, availableHeight/10)
	toolsHeight := availableHeight - searchHeight

	// Right side: dynamic proportions based on install count
	// More installs = larger install panel, smaller info panel
	installLineCount := len(m.installs) + 4 // +4 for title, borders, help text, padding
	if installLineCount < 8 {
		installLineCount = 8 // minimum height
	}
	
	maxInstallHeight := (availableHeight * 2) / 3 // max 67% of height
	installHeight := installLineCount
	if installHeight > maxInstallHeight {
		installHeight = maxInstallHeight
	}
	
	// Info panel gets remaining space
	infoHeight := availableHeight - installHeight

	// Render panels
	searchPanel := m.renderSearchPanel(leftWidth, searchHeight)
	toolsPanel := m.renderToolsPanel(leftWidth, toolsHeight)
	infoPanel := m.renderInfoPanel(rightWidth, infoHeight)
	installPanel := m.renderInstallPanel(rightWidth, installHeight)

	// Combine left and right columns
	leftColumn := lipgloss.JoinVertical(lipgloss.Left, searchPanel, toolsPanel)
	rightColumn := lipgloss.JoinVertical(lipgloss.Left, infoPanel, installPanel)

	// Join columns horizontally
	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftColumn, rightColumn)

	// Add top margin and status bar
	topSpace := "\n"
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, topSpace, layout, statusBar)
}

// renderPanel renders a single panel with border
func (m *Model) renderPanel(id PanelID, title string, width, height int, placeholder string) string {
	// Determine if this panel is active
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if id == m.activePanel {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
	}

	// Create title with highlighting
	titleStr := titleStyle.Render(fmt.Sprintf(" %s ", title))

	// Content (placeholder for now)
	content := placeholder

	// Calculate content area (account for borders and title)
	contentHeight := max(1, height-4) // 2 for top/bottom border, 1 for title, 1 for padding
	contentWidth := max(1, width-4)   // 2 for left/right border, 2 for padding

	// Create bordered box
	boxStyle := borderStyle.
		Width(contentWidth).
		Height(contentHeight).
		Padding(0, 1)

	panel := boxStyle.Render(content)

	// Wrap in outer border with title
	return borderStyle.
		Width(width-1).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, titleStr, panel))
}

// renderSearchPanel renders the search panel with live input
func (m *Model) renderSearchPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelSearch {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
		m.searchPanel.SetStyles()
	}

	title := titleStyle.Render(" Search ")

	// Add searching indicator
	if m.searching {
		title += styles.MutedStyle.Render(" [searching...]")
	}

	// Render search input
	content := m.searchPanel.View(width-4, height-4)

	// Wrap in border
	return borderStyle.
		Width(width-1).
		Height(height-1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderToolsPanel renders the tools list
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

	// Render tools table
	content := m.toolsPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderInfoPanel renders the info panel
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

	// Render info content
	content := m.infoPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, titleRendered, content))
}

// renderInstallPanel renders the install options panel
func (m *Model) renderInstallPanel(width, height int) string {
	borderStyle := styles.InactiveBorder
	titleStyle := styles.MutedStyle

	if m.activePanel == PanelInstall {
		borderStyle = styles.ActiveBorder
		titleStyle = styles.TitleStyle
	}

	title := titleStyle.Render(" Install ")

	// Render install options
	content := m.installPanel.View(width-4, height-4)

	return borderStyle.
		Width(width-1).
		Height(height-1).
		Padding(0, 1).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	// Show keybindings
	help := styles.HelpStyle.Render("Tab: panels | Alt+I: install | Alt+M: install via mise | Alt+R: repo | Alt+U: update | Alt+Q: quit | ?: help")

	// Show error if present
	if m.err != nil {
		errMsg := styles.ErrorStyle.Render(fmt.Sprintf(" Error: %v ", m.err))
		return lipgloss.JoinHorizontal(lipgloss.Left, errMsg, " | ", help)
	}

	// Show tool count
	if len(m.tools) > 0 {
		count := styles.StatusBarStyle.Render(fmt.Sprintf(" %d tools ", len(m.tools)))
		return lipgloss.JoinHorizontal(lipgloss.Left, count, " | ", help)
	}

	return help
}

// renderHelpModal renders the help overlay
func (m *Model) renderHelpModal() string {
	helpText := `
Troveler TUI - Keyboard Shortcuts

Navigation:
  ↑/k, ↓/j     Move cursor up/down
  ←/h, →/l     Select column (in tools table)
  Tab          Cycle between panels
  Enter        Select tool / jump to install panel

Actions:
  Alt+I        Execute install command
  Alt+M        Execute install via mise
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

	// Center the help box
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		helpBox,
	)
}

// renderInfoModal renders the full-screen info modal
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

	// Build full info display
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

// renderUpdateModal renders the update progress modal with slug wave
func (m *Model) renderUpdateModal() string {
	var content string

	if m.updating && m.updateSlugWave != nil {
		// Show slug wave animation
		content = styles.TitleStyle.Render("🔄 Database Update") + "\n\n"
		content += m.updateSlugWave.RenderWithProgress() + "\n\n"
		content += styles.HelpStyle.Render("Press Esc to cancel")
	} else {
		// Update complete or not started
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

// renderInstallModal renders the install execution modal
func (m *Model) renderInstallModal() string {
	var content string

	if m.executing {
		// Show execution in progress
		content = styles.TitleStyle.Render("💻 Executing Install Command") + "\n\n"
		content += "Running install command...\n\n"
		content += styles.MutedStyle.Render("This may take a moment depending on your package manager") + "\n\n"
		content += styles.HelpStyle.Render("Please wait...")
	} else {
		// Show execution results
		content = styles.TitleStyle.Render("💻 Install Complete") + "\n\n"

		if m.err != nil {
			// Show error
			content += styles.ErrorStyle.Render(fmt.Sprintf("Error: %v\n\n", m.err))
			content += styles.HighlightStyle.Render("Output:\n")
			content += styles.MutedStyle.Render(m.executeOutput)
		} else {
			// Show success
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

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
