package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
)

func TestNewToolsPanel(t *testing.T) {
	panel := NewToolsPanel()

	if panel == nil {
		t.Fatal("Expected tools panel to be created")
	}

	if panel.cursor != 0 {
		t.Error("Expected cursor to start at 0")
	}

	if panel.sortAscending != true {
		t.Error("Expected sort to start ascending")
	}
}

func TestToolsPanelSetTools(t *testing.T) {
	panel := NewToolsPanel()
	panel.cursor = 5

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "tool1"}},
		{Tool: db.Tool{Name: "tool2"}},
	}

	panel.SetTools(tools)

	if len(panel.tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(panel.tools))
	}

	if panel.cursor != 0 {
		t.Error("Expected cursor to reset to 0")
	}
}

func TestToolsPanelNavigation(t *testing.T) {
	panel := NewToolsPanel()
	panel.Focus()
	panel.height = 20

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "tool1"}},
		{Tool: db.Tool{Name: "tool2"}},
		{Tool: db.Tool{Name: "tool3"}},
	}
	panel.SetTools(tools)

	// Move down
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if panel.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", panel.cursor)
	}

	// Move down again
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if panel.cursor != 2 {
		t.Errorf("Expected cursor at 2, got %d", panel.cursor)
	}

	// Try to move past end
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if panel.cursor != 2 {
		t.Errorf("Expected cursor to stay at 2, got %d", panel.cursor)
	}

	// Move up
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if panel.cursor != 1 {
		t.Errorf("Expected cursor at 1, got %d", panel.cursor)
	}
}

func TestToolsPanelColumnSelection(t *testing.T) {
	panel := NewToolsPanel()
	panel.Focus()

	// Start at column 0
	if panel.selectedCol != 0 {
		t.Error("Expected to start at column 0")
	}

	// Move right
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if panel.selectedCol != 1 {
		t.Errorf("Expected column 1, got %d", panel.selectedCol)
	}

	// Move right again
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if panel.selectedCol != 2 {
		t.Errorf("Expected column 2, got %d", panel.selectedCol)
	}

	// Move right to installed column
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if panel.selectedCol != 3 {
		t.Errorf("Expected column 3, got %d", panel.selectedCol)
	}

	// Try to move past end
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if panel.selectedCol != 3 {
		t.Errorf("Expected to stay at column 3, got %d", panel.selectedCol)
	}

	// Move left
	panel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
	if panel.selectedCol != 2 {
		t.Errorf("Expected column 2, got %d", panel.selectedCol)
	}
}

func TestToolsPanelSorting(t *testing.T) {
	panel := NewToolsPanel()
	panel.Focus()

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "zebra"}},
		{Tool: db.Tool{Name: "alpha"}},
		{Tool: db.Tool{Name: "beta"}},
	}
	panel.SetTools(tools)

	// Initially unsorted (zebra, alpha, beta)
	if panel.tools[0].Name != "zebra" {
		t.Errorf("Expected first tool before sort to be 'zebra', got %s", panel.tools[0].Name)
	}

	// Trigger sort directly (test the sort logic itself)
	panel.sortTools()

	if panel.tools[0].Name != "alpha" {
		t.Errorf("Expected first tool to be 'alpha' after ascending sort, got %s", panel.tools[0].Name)
	}

	// Toggle to descending
	panel.sortAscending = false
	panel.sortTools()

	if panel.tools[0].Name != "zebra" {
		t.Errorf("Expected first tool to be 'zebra' after descending sort, got %s", panel.tools[0].Name)
	}
}

func TestToolsPanelGetSelectedTool(t *testing.T) {
	panel := NewToolsPanel()

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "tool1"}},
		{Tool: db.Tool{Name: "tool2"}},
	}
	panel.SetTools(tools)

	selected := panel.GetSelectedTool()
	if selected == nil {
		t.Fatal("Expected selected tool")
	}

	if selected.Name != "tool1" {
		t.Errorf("Expected 'tool1', got %s", selected.Name)
	}

	// Move cursor
	panel.cursor = 1
	selected = panel.GetSelectedTool()
	if selected.Name != "tool2" {
		t.Errorf("Expected 'tool2', got %s", selected.Name)
	}
}

func TestToolsPanelView(t *testing.T) {
	panel := NewToolsPanel()

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "tool1", Tagline: "A test tool", Language: "go"}},
	}
	panel.SetTools(tools)

	view := panel.View(80, 20)

	if view == "" {
		t.Error("Expected view to render content")
	}

	if !contains(view, "tool1") {
		t.Error("Expected view to contain tool name")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func TestToolsPanelNarrowTerminal(t *testing.T) {
	panel := NewToolsPanel()

	tools := []db.SearchResult{
		{Tool: db.Tool{Name: "tool1", Tagline: "A test tool", Language: "go"}},
	}
	panel.SetTools(tools)

	// Test with very narrow width (less than minimum)
	view := panel.View(40, 20)

	if view == "" {
		t.Error("Expected view to render even with narrow width")
	}

	// Should show error message about narrow terminal
	if !contains(view, "too narrow") {
		t.Error("Expected view to contain error message about narrow terminal")
	}

	// Test with adequate width
	view = panel.View(80, 20)

	if view == "" {
		t.Error("Expected view to render with adequate width")
	}

	if !contains(view, "tool1") {
		t.Error("Expected view to contain tool name with adequate width")
	}
}
