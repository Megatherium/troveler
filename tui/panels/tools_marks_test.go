package panels

import (
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
)

func TestToolsPanelFilteredMarks(t *testing.T) {
	p := NewToolsPanel()
	p.Focus()
	all := []db.SearchResult{
		{Tool: db.Tool{ID: "a"}},
		{Tool: db.Tool{ID: "b"}},
		{Tool: db.Tool{ID: "c"}},
	}
	p.SetTools(all)
	toggle := func(index int) {
		t.Helper()
		p.cursor = index
		_, cmd := p.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		if cmd == nil {
			t.Fatal("mark key did not produce a mark notification")
		}
		msg, ok := cmd().(ToolMarkedMsg)
		if !ok || msg.Tool.ID != p.tools[index].ID || msg.IsMarked != p.IsMarked(msg.Tool.ID) {
			t.Fatalf("wrong mark notification: %#v", msg)
		}
	}
	assertMarks := func(want ...string) {
		t.Helper()
		var ids []string
		for _, tool := range p.GetMarkedTools() {
			ids = append(ids, tool.ID)
		}
		if !slices.Equal(ids, want) || p.GetMarkedCount() != len(want) {
			t.Errorf("visible marks = %v, count = %d; want %v", ids, p.GetMarkedCount(), want)
		}
	}
	toggle(0)
	toggle(2)
	assertMarks("a", "c")

	p.SetTools(all[1:])
	assertMarks("c")
	if !p.IsMarked("a") {
		t.Error("nonempty filtering discarded the hidden mark")
	}
	toggle(1) // Unmark visible C while hidden A stays marked.
	assertMarks()
	toggle(0) // Mark visible B.
	assertMarks("b")

	p.SetTools(all)
	assertMarks("a", "b")
	p.SetTools(all[2:])
	assertMarks()
	p.ClearMarks() // Clear must discard hidden marks too.
	p.SetTools(all)
	assertMarks()
	if p.IsMarked("a") || p.IsMarked("b") {
		t.Error("clearing marks retained hidden IDs")
	}
}
