package panels

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSearchPanel(t *testing.T) {
	panel := NewSearchPanel()

	if panel == nil {
		t.Fatal("Expected search panel to be created")
	}

	if panel.focused {
		t.Error("Expected panel to start unfocused")
	}

	if panel.GetQuery() != "" {
		t.Error("Expected empty initial query")
	}
}

func TestSearchPanelFocus(t *testing.T) {
	panel := NewSearchPanel()

	if panel.IsFocused() {
		t.Error("Expected panel to start unfocused")
	}

	panel.Focus()

	if !panel.IsFocused() {
		t.Error("Expected panel to be focused after Focus()")
	}

	panel.Blur()

	if panel.IsFocused() {
		t.Error("Expected panel to be unfocused after Blur()")
	}
}

func TestSearchPanelClear(t *testing.T) {
	panel := NewSearchPanel()
	panel.Focus()

	// Set some text
	panel.textInput.SetValue("test query")

	if panel.GetQuery() != "test query" {
		t.Error("Expected query to be set")
	}

	// Press ESC to clear
	msg := tea.KeyMsg{Type: tea.KeyEsc}
	cmd, updatedPanel := panel.Update(msg)

	if cmd == nil {
		t.Error("Expected command to be returned")
	}

	if updatedPanel.GetQuery() != "" {
		t.Errorf("Expected query to be cleared, got %q", updatedPanel.GetQuery())
	}
}

func TestSearchPanelEnterTriggersSearch(t *testing.T) {
	panel := NewSearchPanel()
	panel.Focus()

	panel.textInput.SetValue("test")

	// Press Enter
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	cmd, _ := panel.Update(msg)

	if cmd == nil {
		t.Error("Expected search command to be triggered on Enter")
	}

	// Execute the command to get the message
	result := cmd()

	searchMsg, ok := result.(SearchTriggeredMsg)
	if !ok {
		t.Error("Expected SearchTriggeredMsg")
	}

	if searchMsg.Query != "test" {
		t.Errorf("Expected query 'test', got %q", searchMsg.Query)
	}
}

func TestSearchPanelView(t *testing.T) {
	panel := NewSearchPanel()

	view := panel.View(50, 5)

	if view == "" {
		t.Error("Expected view to render content")
	}
}
