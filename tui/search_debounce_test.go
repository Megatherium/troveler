package tui

import (
	"slices"
	"testing"

	"github.com/charmbracelet/bubbles/cursor"
	tea "github.com/charmbracelet/bubbletea"

	"troveler/tui/panels"
)

func TestDebouncedSearchRoutesAfterPanelSwitch(t *testing.T) {
	m := newTestModelWithDB(t)
	_, first := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	_, latest := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.activePanel != PanelTools || m.searchPanel.IsFocused() {
		t.Fatal("test requires changing focus before timer delivery")
	}
	queries := runSearchCommandsForTest(t, m, first)
	queries = append(queries, runSearchCommandsForTest(t, m, latest)...)
	if !slices.Equal(queries, []string{"ab"}) || !m.searching {
		t.Fatalf("database searches=%v, searching=%v; want only ab", queries, m.searching)
	}
}

func TestImmediateSearchInvalidatesDebounce(t *testing.T) {
	for _, key := range []tea.KeyType{tea.KeyEnter, tea.KeyEsc} {
		t.Run(tea.KeyMsg{Type: key}.String(), func(t *testing.T) {
			m := newTestModelWithDB(t)
			_, pending := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			_, immediate := m.Update(tea.KeyMsg{Type: key})
			want := ""
			if key == tea.KeyEnter {
				want = "a"
			}
			if m.searchPanel.GetQuery() != want {
				t.Fatalf("query=%q; want %q", m.searchPanel.GetQuery(), want)
			}
			queries := runSearchCommandsForTest(t, m, immediate)
			queries = append(queries, runSearchCommandsForTest(t, m, pending)...)
			if !slices.Equal(queries, []string{want}) {
				t.Fatalf("immediate key dispatched %v; want only %q", queries, want)
			}
		})
	}
}

func TestQueuedImmediateSearchInvalidatedByEdits(t *testing.T) {
	for _, repeatQuery := range []bool{false, true} {
		t.Run(map[bool]string{false: "different query", true: "same query again"}[repeatQuery], func(t *testing.T) {
			m := newTestModelWithDB(t)
			_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
			_, queued := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			_, latest := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
			want := "ab"
			if repeatQuery {
				_, latest = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				want = "a"
			}
			queries := runSearchCommandsForTest(t, m, queued)
			if len(queries) != 0 || m.searching {
				t.Fatalf("superseded queued Enter launched %v, searching=%v", queries, m.searching)
			}
			queries = runSearchCommandsForTest(t, m, latest)
			if !slices.Equal(queries, []string{want}) || !m.searching {
				t.Fatalf("latest edit launched %v, searching=%v; want %q", queries, m.searching, want)
			}
		})
	}
}

func TestQueuedDebouncedSearchInvalidatedByEdits(t *testing.T) {
	m := newTestModelWithDB(t)
	_, pending := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	// Deliver the expiry, then edit before its returned trigger is delivered.
	_, queued := m.Update(searchExpiryForTest(t, pending))
	if queued == nil {
		t.Fatal("current timer did not produce a queued trigger")
	}
	_, latest := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	queries := runSearchCommandsForTest(t, m, queued)
	if len(queries) != 0 || m.searching {
		t.Fatalf("superseded debounce trigger launched %v, searching=%v", queries, m.searching)
	}
	queries = runSearchCommandsForTest(t, m, latest)
	if !slices.Equal(queries, []string{"ab"}) || !m.searching {
		t.Fatalf("latest edit launched %v, searching=%v; want ab", queries, m.searching)
	}
}

func searchExpiryForTest(t *testing.T, cmd tea.Cmd) panels.SearchDebounceMsg {
	t.Helper()
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		// The debounce command is last; text-input cursor commands come first.
		return searchExpiryForTest(t, batch[len(batch)-1])
	}
	expiry, ok := msg.(panels.SearchDebounceMsg)
	if !ok {
		t.Fatalf("expected timer expiry, got %T", msg)
	}
	return expiry
}

// Drive returned search commands through the real model until a database result
// is produced. Leave result delivery to the separate in-flight-search tests.
func runSearchCommandsForTest(t *testing.T, m *Model, cmd tea.Cmd) []string {
	t.Helper()
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		var queries []string
		for _, child := range msg {
			queries = append(queries, runSearchCommandsForTest(t, m, child)...)
		}
		return queries
	case searchResultMsg:
		return []string{msg.query}
	case searchErrorMsg:
		t.Fatalf("database search failed: %v", msg.err)
	case cursor.BlinkMsg:
		return nil
	default:
		_, next := m.Update(msg)
		return runSearchCommandsForTest(t, m, next)
	}
	return nil
}
