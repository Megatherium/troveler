package styles

import (
	"troveler/pkg/ui"

	"github.com/charmbracelet/lipgloss"
)

// Theme colors and styles for the TUI.

// BorderStyle is the base style for borders.
var BorderStyle = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder())

// ActiveBorder is the border style for the active panel.
var ActiveBorder = BorderStyle.BorderForeground(lipgloss.Color("#00FFFF"))

// InactiveBorder is the border style for inactive panels.
var InactiveBorder = BorderStyle.BorderForeground(lipgloss.Color("#444444"))

// TitleStyle is used for panel titles.
var TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FFFF"))

// SubtitleStyle is used for secondary text like taglines.
var SubtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Italic(true)

// HighlightStyle is used for emphasized text.
var HighlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF00"))

// ErrorStyle is used for error messages.
var ErrorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF0000"))

// MutedStyle is used for less prominent text.
var MutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

// SelectedStyle is used for the selected row in a panel.
var SelectedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#00FFFF")).
	Background(lipgloss.Color("#003333"))

// UnselectedStyle is used for non-selected interactive elements.
var UnselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

// CursorStyle is used for the cursor in input fields.
var CursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF"))

// MarkedStyle is used for marked (selected for batch) items.
var MarkedStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#FFFF00")).
	Background(lipgloss.Color("#333300"))

// MarkedSelectedStyle is used for marked items when focused.
var MarkedSelectedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FFFF00")).
	Background(lipgloss.Color("#555500"))

// StatusBarStyle is used for the status bar text.
var StatusBarStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))

// HelpStyle is used for help text.
var HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))

// GradientColors is the list of gradient colors (reused from ui.GradientColors).
var GradientColors = ui.GradientColors

// GetGradientColor returns a gradient color based on position.
func GetGradientColor(index int) lipgloss.Color {
	return lipgloss.Color(ui.GetGradientColorSimple(index))
}

// ApplyGradient applies gradient color to a style based on row index.
func ApplyGradient(style lipgloss.Style, index int) lipgloss.Style {
	return style.Foreground(GetGradientColor(index))
}
