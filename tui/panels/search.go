package panels

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/tui/styles"
)

// SearchPanel handles the search input
type SearchPanel struct {
	textInput        textinput.Model
	focused          bool
	lastQuery        string
	searchGeneration uint64
	debounceTime     time.Duration
	width            int
	height           int
}

// SearchTriggeredMsg is sent when search should be executed
type SearchTriggeredMsg struct {
	Query      string
	generation uint64
}

// SearchDebounceMsg returns a timer expiry to the search panel's Update thread.
type SearchDebounceMsg struct {
	query      string
	generation uint64
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

// Init satisfies tea.Model
func (p *SearchPanel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *SearchPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SearchDebounceMsg:
		if msg.generation == 0 || msg.generation != p.searchGeneration {
			return p, nil
		}
		// Consume this timer once, invalidating duplicate deliveries.
		p.searchGeneration++

		return p, p.triggerSearch(msg.query)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Clear search
			p.textInput.SetValue("")
			p.lastQuery = ""
			p.searchGeneration++

			return p, p.triggerSearch("")

		case tea.KeyEnter:
			// Immediate search on Enter
			query := p.textInput.Value()
			p.lastQuery = query
			p.searchGeneration++

			return p, p.triggerSearch(query)
		}
	}

	// Update text input
	p.textInput, cmd = p.textInput.Update(msg)

	// Debounced search on value change
	currentQuery := p.textInput.Value()
	if currentQuery != p.lastQuery {
		p.lastQuery = currentQuery
		p.searchGeneration++

		return p, tea.Batch(
			cmd,
			p.debounceSearchCmd(currentQuery),
		)
	}

	return p, cmd
}

// debounceSearchCmd snapshots a timer's query and generation before it runs.
func (p *SearchPanel) debounceSearchCmd(query string) tea.Cmd {
	generation := p.searchGeneration

	return tea.Tick(p.debounceTime, func(time.Time) tea.Msg {
		return SearchDebounceMsg{query: query, generation: generation}
	})
}

// triggerSearch immediately triggers a search
func (p *SearchPanel) triggerSearch(query string) tea.Cmd {
	msg := SearchTriggeredMsg{Query: query, generation: p.searchGeneration}

	return func() tea.Msg {
		return msg
	}
}

// MatchesSearch rejects queued triggers superseded before database dispatch.
func (p *SearchPanel) MatchesSearch(msg SearchTriggeredMsg) bool {
	return msg.generation != 0 && msg.generation == p.searchGeneration
}

// View renders the search panel
func (p *SearchPanel) View() string {
	// Adjust input width to fit panel, ensuring it's at least 10 columns
	inputWidth := p.width - 4
	if inputWidth < 10 {
		inputWidth = 10
	}
	p.textInput.Width = inputWidth

	return p.textInput.View()
}

// SetSize stores the panel dimensions for use in View()
func (p *SearchPanel) SetSize(width, height int) {
	p.width = width
	p.height = height
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
