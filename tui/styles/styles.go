package styles

import (
	"troveler/pkg/ui"

	"github.com/charmbracelet/lipgloss"
)

// Theme colors and styles for the TUI
var (
	// Borders
	BorderStyle    = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())
	ActiveBorder   = BorderStyle.BorderForeground(lipgloss.Color("#00FFFF"))
	InactiveBorder = BorderStyle.BorderForeground(lipgloss.Color("#444444"))

	// Titles
	TitleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))
	SubtitleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Italic(true)
	HighlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))
	ErrorStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000"))
	MutedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Interactive elements
	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FFFF")).
			Background(lipgloss.Color("#003333"))
	UnselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	CursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))
	MarkedStyle     = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Background(lipgloss.Color("#333300"))
	MarkedSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFF00")).
				Background(lipgloss.Color("#555500"))

	// Status
	StatusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	HelpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

	// Gradient colors (reuse from existing gradient)
	GradientColors = ui.GradientColors
)

// GetGradientColor returns a gradient color based on position
func GetGradientColor(index int) lipgloss.Color {
	return lipgloss.Color(ui.GetGradientColorSimple(index))
}

// ApplyGradient applies gradient color to a style based on row index
func ApplyGradient(style lipgloss.Style, index int) lipgloss.Style {
	return style.Foreground(GetGradientColor(index))
}
