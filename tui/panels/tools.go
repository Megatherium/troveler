package panels

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"troveler/db"
	"troveler/tui/styles"
)

// ToolsPanel displays the tools list in a table
type ToolsPanel struct {
	tools         []db.SearchResult
	cursor        int
	selectedCol   int // 0=slug, 1=tagline, 2=language
	sortCol       int
	sortAscending bool
	focused       bool
	width         int
	height        int
	scrollOffset  int
}

// ToolSelectedMsg is sent when a tool is selected
type ToolSelectedMsg struct {
	Tool db.SearchResult
}

// NewToolsPanel creates a new tools panel
func NewToolsPanel() *ToolsPanel {
	return &ToolsPanel{
		tools:         []db.SearchResult{},
		cursor:        0,
		selectedCol:   0,
		sortCol:       0,
		sortAscending: true,
		focused:       false,
	}
}

// SetTools updates the tools list
func (p *ToolsPanel) SetTools(tools []db.SearchResult) {
	p.tools = tools
	p.cursor = 0
	p.scrollOffset = 0
}

// Update handles messages
func (p *ToolsPanel) Update(msg tea.Msg) tea.Cmd {
	if !p.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("down", "j"))):
			if p.cursor < len(p.tools)-1 {
				p.cursor++
				p.adjustScroll()
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if p.cursor > 0 {
				p.cursor--
				p.adjustScroll()
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if p.selectedCol > 0 {
				p.selectedCol--
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if p.selectedCol < 2 {
				p.selectedCol++
			}
			return nil

		case msg.Alt && (msg.Type == tea.KeyRunes && len(msg.Runes) > 0 && msg.Runes[0] == 's'):
			// Toggle sort
			if p.sortCol == p.selectedCol {
				p.sortAscending = !p.sortAscending
			} else {
				p.sortCol = p.selectedCol
				p.sortAscending = true
			}
			p.sortTools()
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			// Select tool
			if p.cursor < len(p.tools) {
				return func() tea.Msg {
					return ToolSelectedMsg{Tool: p.tools[p.cursor]}
				}
			}
			return nil
		}
	}

	return nil
}

// adjustScroll adjusts scroll offset to keep cursor visible
func (p *ToolsPanel) adjustScroll() {
	visibleRows := p.height - 4 // Account for header and borders
	if visibleRows < 1 {
		visibleRows = 1
	}

	if p.cursor < p.scrollOffset {
		p.scrollOffset = p.cursor
	} else if p.cursor >= p.scrollOffset+visibleRows {
		p.scrollOffset = p.cursor - visibleRows + 1
	}
}

// sortTools sorts the tools based on current sort settings
func (p *ToolsPanel) sortTools() {
	// Simple bubble sort for now (good enough for UI)
	for i := 0; i < len(p.tools)-1; i++ {
		for j := 0; j < len(p.tools)-i-1; j++ {
			var swap bool
			switch p.sortCol {
			case 0: // slug/name
				if p.sortAscending {
					swap = p.tools[j].Name > p.tools[j+1].Name
				} else {
					swap = p.tools[j].Name < p.tools[j+1].Name
				}
			case 1: // tagline
				if p.sortAscending {
					swap = p.tools[j].Tagline > p.tools[j+1].Tagline
				} else {
					swap = p.tools[j].Tagline < p.tools[j+1].Tagline
				}
			case 2: // language
				if p.sortAscending {
					swap = p.tools[j].Language > p.tools[j+1].Language
				} else {
					swap = p.tools[j].Language < p.tools[j+1].Language
				}
			}
			if swap {
				p.tools[j], p.tools[j+1] = p.tools[j+1], p.tools[j]
			}
		}
	}
}

// View renders the tools table
func (p *ToolsPanel) View(width, height int) string {
	p.width = width
	p.height = height

	if len(p.tools) == 0 {
		return styles.MutedStyle.Render("No tools found")
	}

	var b strings.Builder

	// Calculate column widths
	nameWidth := 20
	taglineWidth := width - nameWidth - 15 - 10 // Adjust for language and borders
	langWidth := 10

	// Render header
	headers := []string{
		p.renderHeader("Name", 0, nameWidth),
		p.renderHeader("Tagline", 1, taglineWidth),
		p.renderHeader("Language", 2, langWidth),
	}
	b.WriteString(strings.Join(headers, " │ "))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", width-4))
	b.WriteString("\n")

	// Render rows
	visibleRows := height - 4
	if visibleRows < 1 {
		visibleRows = 1
	}

	end := p.scrollOffset + visibleRows
	if end > len(p.tools) {
		end = len(p.tools)
	}

	for i := p.scrollOffset; i < end; i++ {
		tool := p.tools[i]
		row := p.renderRow(i, tool, nameWidth, taglineWidth, langWidth)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

// renderHeader renders a column header with sort indicator
func (p *ToolsPanel) renderHeader(title string, col int, width int) string {
	// Add sort indicator
	indicator := " "
	if p.sortCol == col {
		if p.sortAscending {
			indicator = "▲"
		} else {
			indicator = "▼"
		}
	}

	// Add selection indicator
	style := styles.TitleStyle
	if col == p.selectedCol && p.focused {
		style = styles.HighlightStyle
	}

	text := fmt.Sprintf("%s %s", title, indicator)
	if len(text) > width {
		text = text[:width]
	}

	return style.Render(lipgloss.NewStyle().Width(width).Render(text))
}

// renderRow renders a single tool row
func (p *ToolsPanel) renderRow(idx int, tool db.SearchResult, nameWidth, taglineWidth, langWidth int) string {
	// Truncate fields to fit
	name := tool.Name
	if len(name) > nameWidth {
		name = name[:nameWidth-3] + "..."
	}

	tagline := tool.Tagline
	if len(tagline) > taglineWidth {
		tagline = tagline[:taglineWidth-3] + "..."
	}

	lang := tool.Language
	if len(lang) > langWidth {
		lang = lang[:langWidth-3] + "..."
	}

	// Apply gradient color
	gradient := styles.GetGradientColor(idx)

	// Highlight if selected
	if idx == p.cursor && p.focused {
		nameStyle := styles.SelectedStyle.Foreground(gradient)
		taglineStyle := styles.SelectedStyle.Foreground(gradient)
		langStyle := styles.SelectedStyle.Foreground(gradient)

		return fmt.Sprintf("%s │ %s │ %s",
			nameStyle.Width(nameWidth).Render(name),
			taglineStyle.Width(taglineWidth).Render(tagline),
			langStyle.Width(langWidth).Render(lang),
		)
	}

	// Normal style with gradient
	style := lipgloss.NewStyle().Foreground(gradient)

	return fmt.Sprintf("%s │ %s │ %s",
		style.Width(nameWidth).Render(name),
		style.Width(taglineWidth).Render(tagline),
		style.Width(langWidth).Render(lang),
	)
}

// Focus focuses the panel
func (p *ToolsPanel) Focus() {
	p.focused = true
}

// Blur unfocuses the panel
func (p *ToolsPanel) Blur() {
	p.focused = false
}

// IsFocused returns focus state
func (p *ToolsPanel) IsFocused() bool {
	return p.focused
}

// GetSelectedTool returns the currently selected tool
func (p *ToolsPanel) GetSelectedTool() *db.SearchResult {
	if p.cursor >= 0 && p.cursor < len(p.tools) {
		return &p.tools[p.cursor]
	}
	return nil
}
