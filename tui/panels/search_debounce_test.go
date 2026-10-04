package panels

import (
	"slices"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchDebounceRapidEdits(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		t.Run(map[bool]string{false: "in order", true: "reversed"}[reverse], func(t *testing.T) {
			p := debounceTestPanel()
			var queued []tea.Cmd
			for _, key := range []tea.KeyMsg{
				{Type: tea.KeyRunes, Runes: []rune{'a'}},
				{Type: tea.KeyRunes, Runes: []rune{'b'}},
				{Type: tea.KeyBackspace}, // Return to A: query equality cannot identify old timers.
			} {
				_, cmd := p.Update(key)
				queued = append(queued, cmd)
			}
			if reverse {
				slices.Reverse(queued)
			}
			var queries []string
			for _, cmd := range queued {
				queries = append(queries, resolvePanelSearch(t, p, cmd)...)
			}
			if !slices.Equal(queries, []string{"a"}) {
				t.Fatalf("rapid edits dispatched %v; want only the latest A", queries)
			}
		})
	}
}

func TestSearchDebounceImmediateAndClearInvalidatePending(t *testing.T) {
	for _, key := range []tea.KeyType{tea.KeyEnter, tea.KeyEsc, tea.KeyBackspace} {
		t.Run(tea.KeyMsg{Type: key}.String(), func(t *testing.T) {
			p := debounceTestPanel()
			_, pending := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			_, immediate := p.Update(tea.KeyMsg{Type: key})
			want := ""
			if key == tea.KeyEnter {
				want = "a"
			}
			queries := resolvePanelSearch(t, p, immediate)
			queries = append(queries, resolvePanelSearch(t, p, pending)...)
			if p.GetQuery() != want || !slices.Equal(queries, []string{want}) {
				t.Fatalf("query=%q, dispatched=%v; want only %q", p.GetQuery(), queries, want)
			}
		})
	}
}

func TestSearchDebounceWaitsForInterval(t *testing.T) {
	p := debounceTestPanel()
	p.debounceTime = 20 * time.Millisecond
	started := time.Now()
	_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	queries := resolvePanelSearch(t, p, cmd)
	if elapsed := time.Since(started); elapsed < p.debounceTime {
		t.Fatalf("search dispatched after %v, before debounce interval %v", elapsed, p.debounceTime)
	}
	if !slices.Equal(queries, []string{"a"}) {
		t.Fatalf("query never dispatched after debounce interval: %v", queries)
	}
}

func TestSearchDebounceTimerConsumedOnce(t *testing.T) {
	p := debounceTestPanel()
	_, pending := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_, _ = p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	// Evaluate A's background callback only after the query has changed.
	old, ok := pending().(SearchDebounceMsg)
	if !ok || old.query != "a" {
		t.Fatal("timer did not snapshot its query before running")
	}
	_, stale := p.Update(old)
	if stale != nil {
		t.Fatal("old callback read the current generation and dispatched A")
	}
	_, latest := p.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	expiry := latest()
	_, trigger := p.Update(expiry)
	if trigger == nil || !p.MatchesSearch(trigger().(SearchTriggeredMsg)) {
		t.Fatal("current timer failed to produce a valid trigger")
	}
	_, duplicate := p.Update(expiry)
	if duplicate != nil {
		t.Fatal("delivering one timer twice produced a duplicate search")
	}
}

func debounceTestPanel() *SearchPanel {
	p := NewSearchPanel()
	p.Focus()
	p.textInput.Cursor.SetMode(cursor.CursorStatic)
	p.debounceTime = time.Millisecond
	return p
}

func resolvePanelSearch(t *testing.T, p *SearchPanel, cmd tea.Cmd) []string {
	t.Helper()
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		var queries []string
		for _, child := range msg {
			queries = append(queries, resolvePanelSearch(t, p, child)...)
		}
		return queries
	case SearchTriggeredMsg:
		return []string{msg.Query}
	default:
		_, next := p.Update(msg)
		return resolvePanelSearch(t, p, next)
	}
}
