package panels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"troveler/db"
	"troveler/tui/styles"
)

// ToolsPanel displays the tools list in a table.
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
	markedTools   map[string]bool // Set of marked tool IDs for batch install
}

// ToolSelectedMsg is sent when a tool is selected (Enter pressed).
type ToolSelectedMsg struct {
	Tool db.SearchResult
}

// ToolMarkedMsg is sent when a tool's mark status changes.
type ToolMarkedMsg struct {
	Tool     db.SearchResult
	IsMarked bool
}

// ToolCursorChangedMsg is sent when the cursor moves to a different tool.
type ToolCursorChangedMsg struct {
	Tool db.SearchResult
}

// NewToolsPanel creates a new tools panel.
func NewToolsPanel() *ToolsPanel {
	return &ToolsPanel{
		tools:         []db.SearchResult{},
		cursor:        0,
		selectedCol:   0,
		sortCol:       0,
		sortAscending: true,
		focused:       false,
		installedMap:  make(map[string]bool),
		markedTools:   make(map[string]bool),
	}
}

// SetTools updates the tools list.
func (p *ToolsPanel) SetTools(tools []db.SearchResult) {
	p.tools = tools
	p.cursor = 0
	p.scrollOffset = 0
}

// Update handles messages.
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

		case key.Matches(msg, key.NewBinding(key.WithKeys("m"))):
			// Toggle mark for batch install
			if p.cursor < len(p.tools) {
				tool := p.tools[p.cursor]
				isMarked := p.markedTools[tool.ID]
				if isMarked {
					delete(p.markedTools, tool.ID)
				} else {
					p.markedTools[tool.ID] = true
				}

				return func() tea.Msg {
					return ToolMarkedMsg{Tool: tool, IsMarked: !isMarked}
				}
			}

			return nil
		}
	}

	return nil
}

// adjustScroll adjusts scroll offset to keep cursor visible.
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

// sortTools sorts the tools based on current sort settings.
func (p *ToolsPanel) sortTools() {
	sort.Slice(p.tools, func(i, j int) bool {
		switch p.sortCol {
		case 0:
			if p.sortAscending {
				return p.tools[i].Name < p.tools[j].Name
			}

			return p.tools[i].Name > p.tools[j].Name
		case 1:
			if p.sortAscending {
				return p.tools[i].Tagline < p.tools[j].Tagline
			}

			return p.tools[i].Tagline > p.tools[j].Tagline
		case 2:
			if p.sortAscending {
				return p.tools[i].Language < p.tools[j].Language
			}

			return p.tools[i].Language > p.tools[j].Language
		case 3:
			if p.sortAscending {
				return p.tools[i].Installed && !p.tools[j].Installed
			}

			return !p.tools[i].Installed && p.tools[j].Installed
		}

		return false
	})
}

// View renders the tools panel.
func (p *ToolsPanel) View(width, height int) string {
	p.width = width
	p.height = height

	if len(p.tools) == 0 {
		return styles.MutedStyle.Render("No tools found")
	}

	// Minimum width needed to render table safely
	minWidth := 60
	if width < minWidth {
		return fmt.Sprintf("Terminal too narrow (%d columns).\nMinimum %d columns required.", width, minWidth)
	}

	var b strings.Builder

	installedWidth := 9
	nameWidth := 25
	taglineWidth := width - nameWidth - installedWidth - 15 - 10
	langWidth := 10

	// Ensure taglineWidth doesn't become negative or too small
	if taglineWidth < 10 {
		taglineWidth = 10
	}

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

// renderHeader renders a column header with sort indicator.
func (p *ToolsPanel) renderHeader(title string, col int, width int) string {
	// Ensure width is positive
	if width <= 0 {
		width = 1
	}

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

// renderRow renders a single tool row.
func (p *ToolsPanel) renderRow(
	idx int, tool db.SearchResult, nameWidth, taglineWidth, langWidth, installedWidth int,
) string {
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

	// Check if tool is marked for batch install
	isMarked := p.markedTools[tool.ID]
	markIndicator := "  "
	if isMarked {
		markIndicator = "● "
	}

	// Apply gradient color
	gradient := styles.GetGradientColor(idx)

	// Highlight if selected (cursor on this row)
	if idx == p.cursor && p.focused {
		var baseStyle lipgloss.Style
		if isMarked {
			baseStyle = styles.MarkedSelectedStyle.Foreground(gradient)
		} else {
			baseStyle = styles.SelectedStyle.Foreground(gradient)
		}

		return fmt.Sprintf("%s%s │ %s │ %s │ %s",
			baseStyle.Width(2).Render(markIndicator),
			baseStyle.Width(nameWidth-2).Render(name),
			baseStyle.Width(taglineWidth).Render(tagline),
			baseStyle.Width(langWidth).Render(lang),
			baseStyle.Width(installedWidth).Render(installed),
		)
	}

	// Marked but not selected
	if isMarked {
		markedStyle := styles.MarkedStyle.Foreground(gradient)

		return fmt.Sprintf("%s%s │ %s │ %s │ %s",
			markedStyle.Width(2).Render(markIndicator),
			markedStyle.Width(nameWidth-2).Render(name),
			markedStyle.Width(taglineWidth).Render(tagline),
			markedStyle.Width(langWidth).Render(lang),
			markedStyle.Width(installedWidth).Render(installed),
		)
	}

	// Normal style with gradient
	style := lipgloss.NewStyle().Foreground(gradient)

	return fmt.Sprintf("%s%s │ %s │ %s │ %s",
		style.Width(2).Render(markIndicator),
		style.Width(nameWidth-2).Render(name),
		style.Width(taglineWidth).Render(tagline),
		style.Width(langWidth).Render(lang),
		style.Width(installedWidth).Render(installed),
	)
}

// Focus focuses the tools panel.
func (p *ToolsPanel) Focus() {
	p.focused = true
}

// Blur unfocuses the tools panel.
func (p *ToolsPanel) Blur() {
	p.focused = false
}

// IsFocused returns whether the tools panel is focused.
func (p *ToolsPanel) IsFocused() bool {
	return p.focused
}

// GetSelectedTool returns the currently selected tool.
func (p *ToolsPanel) GetSelectedTool() *db.SearchResult {
	if p.cursor >= 0 && p.cursor < len(p.tools) {
		return &p.tools[p.cursor]
	}

	return nil
}

// UpdateInstalledStatus updates the installed status for all tools.
// This should be called when install instructions are available.
func (p *ToolsPanel) UpdateInstalledStatus(installs []db.InstallInstruction) {
	for i := range p.tools {
		if _, ok := p.installedMap[p.tools[i].ID]; !ok {
			isInstalled := db.IsInstalled(&p.tools[i].Tool, installs)
			p.installedMap[p.tools[i].ID] = isInstalled
			p.tools[i].Installed = isInstalled
		}
	}
}

// UpdateToolInstalledStatus updates the installed status for a specific tool.
func (p *ToolsPanel) UpdateToolInstalledStatus(toolID string, isInstalled bool) {
	p.installedMap[toolID] = isInstalled
	for i := range p.tools {
		if p.tools[i].ID == toolID {
			p.tools[i].Installed = isInstalled

			break
		}
	}
}

// UpdateAllInstalledStatus updates installed status for all tools using a callback.
func (p *ToolsPanel) UpdateAllInstalledStatus(getInstalls func(string) ([]db.InstallInstruction, error)) {
	for i := range p.tools {
		installs, err := getInstalls(p.tools[i].ID)
		if err == nil {
			p.tools[i].Installed = db.IsInstalled(&p.tools[i].Tool, installs)
			p.installedMap[p.tools[i].ID] = p.tools[i].Installed
		}
	}
}

// GetTool returns a tool by index.
func (p *ToolsPanel) GetTool(idx int) *db.SearchResult {
	if idx >= 0 && idx < len(p.tools) {
		return &p.tools[idx]
	}

	return nil
}

// GetMarkedTools returns all marked tools for batch install.
func (p *ToolsPanel) GetMarkedTools() []db.SearchResult {
	var marked []db.SearchResult
	for _, tool := range p.tools {
		if p.markedTools[tool.ID] {
			marked = append(marked, tool)
		}
	}

	return marked
}

// GetMarkedCount returns the number of marked tools.
func (p *ToolsPanel) GetMarkedCount() int {
	return len(p.markedTools)
}

// ClearMarks clears all marked tools.
func (p *ToolsPanel) ClearMarks() {
	p.markedTools = make(map[string]bool)
}

// IsMarked returns true if a tool is marked.
func (p *ToolsPanel) IsMarked(toolID string) bool {
	return p.markedTools[toolID]
}
