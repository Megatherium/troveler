// Package panels implements the TUI panel components.
package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/info"
	"troveler/tui/styles"
)

// InfoPanel displays tool information
type InfoPanel struct {
	viewport viewport.Model
	tool     *info.ToolInfo
	focused  bool
	ready    bool
	width    int
	height   int
}

// NewInfoPanel creates a new info panel
func NewInfoPanel() *InfoPanel {
	vp := viewport.New(40, 20)

	return &InfoPanel{
		viewport: vp,
		focused:  false,
		ready:    false,
	}
}

// SetTool updates the displayed tool
func (p *InfoPanel) SetTool(tool *db.Tool, installs []db.InstallInstruction) {
	p.tool = info.FormatTool(tool, installs)
	p.updateContent()
}

// Clear clears the displayed tool
func (p *InfoPanel) Clear() {
	p.tool = nil
	p.updateContent()
}

// updateContent renders the tool info into the viewport
func (p *InfoPanel) updateContent() {
	if p.tool == nil {
		p.viewport.SetContent(styles.MutedStyle.Render("Select a tool to view details"))

		return
	}

	var b strings.Builder

	// Name
	b.WriteString(styles.TitleStyle.Render(p.tool.Name))
	b.WriteString("\n\n")

	// Tagline
	if p.tool.Tagline != "" {
		b.WriteString(styles.SubtitleStyle.Render(p.tool.Tagline))
		b.WriteString("\n\n")
	}

	// Description
	if p.tool.Description != "" {
		b.WriteString(styles.HighlightStyle.Render("Description:"))
		b.WriteString("\n")
		b.WriteString(info.WrapText(p.tool.Description, p.viewport.Width-4))
		b.WriteString("\n\n")
	}

	// Metadata
	b.WriteString(styles.HighlightStyle.Render("Details:"))
	b.WriteString("\n")

	pairs := p.tool.GetKeyValuePairs()
	for _, pair := range pairs {
		key := styles.MutedStyle.Render(pair[0] + ":")
		val := pair[1]
		b.WriteString("  " + key + " " + val + "\n")
	}

	p.viewport.SetContent(b.String())
}

// Update handles messages
func (p *InfoPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !p.focused || !p.ready {
		return p, nil
	}

	var cmd tea.Cmd
	p.viewport, cmd = p.viewport.Update(msg)

	return p, cmd
}

// View renders the info panel
func (p *InfoPanel) View() string {
	if !p.ready || p.viewport.Width != p.width || p.viewport.Height != p.height {
		p.viewport.Width = p.width
		p.viewport.Height = p.height
		p.ready = true
		p.updateContent()
	}

	return p.viewport.View()
}

// SetSize stores the panel dimensions for use in View()
func (p *InfoPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Init satisfies tea.Model
func (p *InfoPanel) Init() tea.Cmd {
	return nil
}

// Focus focuses the panel
func (p *InfoPanel) Focus() {
	p.focused = true
}

// Blur unfocuses the panel
func (p *InfoPanel) Blur() {
	p.focused = false
}

// IsFocused returns focus state
func (p *InfoPanel) IsFocused() bool {
	return p.focused
}
