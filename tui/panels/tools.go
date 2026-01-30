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
	selectedCol   int // 0=slug, 1=tagline, 2=language, 3=installed
	sortCol       int
	sortAscending bool
	focused       bool
	width         int
	height        int
	scrollOffset  int
	installedMap  map[string]bool // Cache of installed status by tool ID
}

// ToolSelectedMsg is sent when a tool is selected (Enter pressed)
type ToolSelectedMsg struct {
	Tool db.SearchResult
}

// ToolCursorChangedMsg is sent when the cursor moves to a different tool
type ToolCursorChangedMsg struct {
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
		installedMap:  make(map[string]bool),
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
				// Notify that cursor changed
				if p.cursor < len(p.tools) {
					return func() tea.Msg {
						return ToolCursorChangedMsg{Tool: p.tools[p.cursor]}
					}
				}
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("up", "k"))):
			if p.cursor > 0 {
				p.cursor--
				p.adjustScroll()
				// Notify that cursor changed
				if p.cursor < len(p.tools) {
					return func() tea.Msg {
						return ToolCursorChangedMsg{Tool: p.tools[p.cursor]}
					}
				}
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
			if p.selectedCol > 0 {
				p.selectedCol--
			}
			return nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
			if p.selectedCol < 3 {
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
			case 3: // installed (installed first when ascending)
				if p.sortAscending {
					swap = !p.tools[j].Installed && p.tools[j+1].Installed
				} else {
					swap = p.tools[j].Installed && !p.tools[j+1].Installed
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

	installedWidth := 9
	nameWidth := 25
	taglineWidth := width - nameWidth - installedWidth - 15 - 10
	langWidth := 10

	// Render header
	headers := []string{
		p.renderHeader("Name", 0, nameWidth),
		p.renderHeader("Tagline", 1, taglineWidth),
		p.renderHeader("Language", 2, langWidth),
		p.renderHeader("Installed", 3, installedWidth),
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
		row := p.renderRow(i, tool, nameWidth, taglineWidth, langWidth, installedWidth)
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
func (p *ToolsPanel) renderRow(idx int, tool db.SearchResult, nameWidth, taglineWidth, langWidth, installedWidth int) string {
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

	installed := ""
	if tool.Installed {
		installed = "✓"
	}

	// Apply gradient color
	gradient := styles.GetGradientColor(idx)

	// Highlight if selected
	if idx == p.cursor && p.focused {
		nameStyle := styles.SelectedStyle.Foreground(gradient)
		taglineStyle := styles.SelectedStyle.Foreground(gradient)
		langStyle := styles.SelectedStyle.Foreground(gradient)
		installedStyle := styles.SelectedStyle.Foreground(gradient)

		return fmt.Sprintf("%s │ %s │ %s │ %s",
			nameStyle.Width(nameWidth).Render(name),
			taglineStyle.Width(taglineWidth).Render(tagline),
			langStyle.Width(langWidth).Render(lang),
			installedStyle.Width(installedWidth).Render(installed),
		)
	}

	// Normal style with gradient
	style := lipgloss.NewStyle().Foreground(gradient)

	return fmt.Sprintf("%s │ %s │ %s │ %s",
		style.Width(nameWidth).Render(name),
		style.Width(taglineWidth).Render(tagline),
		style.Width(langWidth).Render(lang),
		style.Width(installedWidth).Render(installed),
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

// UpdateInstalledStatus updates the installed status for all tools
// This should be called when install instructions are available
func (p *ToolsPanel) UpdateInstalledStatus(installs []db.InstallInstruction) {
	for i := range p.tools {
		if _, ok := p.installedMap[p.tools[i].ID]; !ok {
			isInstalled := db.IsInstalled(&p.tools[i].Tool, installs)
			p.installedMap[p.tools[i].ID] = isInstalled
			p.tools[i].Installed = isInstalled
		}
	}
}

// UpdateToolInstalledStatus updates the installed status for a specific tool
func (p *ToolsPanel) UpdateToolInstalledStatus(toolID string, isInstalled bool) {
	p.installedMap[toolID] = isInstalled
	for i := range p.tools {
		if p.tools[i].ID == toolID {
			p.tools[i].Installed = isInstalled
			break
		}
	}
}

// UpdateAllInstalledStatus updates the installed status for all tools using their install instructions
func (p *ToolsPanel) UpdateAllInstalledStatus(getInstalls func(string) ([]db.InstallInstruction, error)) {
	for i := range p.tools {
		installs, err := getInstalls(p.tools[i].ID)
		if err == nil {
			p.tools[i].Installed = db.IsInstalled(&p.tools[i].Tool, installs)
			p.installedMap[p.tools[i].ID] = p.tools[i].Installed
		}
	}
}

// GetTool returns a tool by index
func (p *ToolsPanel) GetTool(idx int) *db.SearchResult {
	if idx >= 0 && idx < len(p.tools) {
		return &p.tools[idx]
	}
	return nil
}
