package panels

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/tui/styles"
)

// SearchPanel handles the search input
type SearchPanel struct {
	textInput    textinput.Model
	focused      bool
	lastQuery    string
	searchTimer  *time.Timer
	debounceTime time.Duration
}

// SearchTriggeredMsg is sent when search should be executed
type SearchTriggeredMsg struct {
	Query string
}

// NewSearchPanel creates a new search panel
func NewSearchPanel() *SearchPanel {
	ti := textinput.New()
	ti.Placeholder = "Search tools..."
	ti.CharLimit = 100
	ti.Width = 50

	return &SearchPanel{
		textInput:    ti,
		focused:      false,
		debounceTime: 150 * time.Millisecond,
	}
}

// Update handles messages
func (p *SearchPanel) Update(msg tea.Msg) (tea.Cmd, *SearchPanel) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Clear search
			p.textInput.SetValue("")
			p.lastQuery = ""
			return p.triggerSearch(""), p

		case tea.KeyEnter:
			// Immediate search on Enter
			query := p.textInput.Value()
			p.lastQuery = query
			return p.triggerSearch(query), p
		}
	}

	// Update text input
	p.textInput, cmd = p.textInput.Update(msg)

	// Debounced search on value change
	currentQuery := p.textInput.Value()
	if currentQuery != p.lastQuery {
		p.lastQuery = currentQuery

		// Cancel existing timer
		if p.searchTimer != nil {
			p.searchTimer.Stop()
		}

		// Start new debounce timer
		query := currentQuery
		return tea.Batch(
			cmd,
			p.debounceSearchCmd(query),
		), p
	}

	return cmd, p
}

// debounceSearchCmd creates a command that triggers search after debounce
func (p *SearchPanel) debounceSearchCmd(query string) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(p.debounceTime)
		return SearchTriggeredMsg{Query: query}
	}
}

// triggerSearch immediately triggers a search
func (p *SearchPanel) triggerSearch(query string) tea.Cmd {
	return func() tea.Msg {
		return SearchTriggeredMsg{Query: query}
	}
}

// View renders the search panel
func (p *SearchPanel) View(width, height int) string {
	// Adjust input width to fit panel
	p.textInput.Width = width - 4

	return p.textInput.View()
}

// Focus focuses the search panel
func (p *SearchPanel) Focus() {
	p.focused = true
	p.textInput.Focus()
}

// Blur unfocuses the search panel
func (p *SearchPanel) Blur() {
	p.focused = false
	p.textInput.Blur()
}

// IsFocused returns whether the panel is focused
func (p *SearchPanel) IsFocused() bool {
	return p.focused
}

// GetQuery returns the current search query
func (p *SearchPanel) GetQuery() string {
	return p.textInput.Value()
}

// SetStyles updates the input style based on focus
func (p *SearchPanel) SetStyles() {
	if p.focused {
		p.textInput.PromptStyle = styles.CursorStyle
		p.textInput.TextStyle = styles.HighlightStyle
	} else {
		p.textInput.PromptStyle = styles.MutedStyle
		p.textInput.TextStyle = styles.UnselectedStyle
	}
}
