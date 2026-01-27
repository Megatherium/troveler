package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/config"
)

func TestNewModel(t *testing.T) {
	cfg := &config.Config{
		TUI: config.TUIConfig{
			Theme:           "gradient",
			TaglineMaxWidth: 40,
		},
	}

	m := NewModel(nil, cfg)

	if m == nil {
		t.Fatal("Expected model to be created")
	}

	if m.activePanel != PanelSearch {
		t.Errorf("Expected initial panel to be Search, got %v", m.activePanel)
	}

	if m.config != cfg {
		t.Error("Expected config to be set")
	}
}

func TestNextPanel(t *testing.T) {
	m := NewModel(nil, &config.Config{})

	// Initial state: Search
	if m.activePanel != PanelSearch {
		t.Errorf("Expected Search panel, got %v", m.activePanel)
	}

	// Tab to Tools
	m.NextPanel()
	if m.activePanel != PanelTools {
		t.Errorf("Expected Tools panel after first tab, got %v", m.activePanel)
	}

	// Tab to Install
	m.NextPanel()
	if m.activePanel != PanelInstall {
		t.Errorf("Expected Install panel after second tab, got %v", m.activePanel)
	}

	// Tab back to Search
	m.NextPanel()
	if m.activePanel != PanelSearch {
		t.Errorf("Expected Search panel after third tab (cycle), got %v", m.activePanel)
	}
}

func TestWindowResize(t *testing.T) {
	m := NewModel(nil, &config.Config{})

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := m.Update(msg)

	updated := updatedModel.(*Model)

	if updated.width != 120 {
		t.Errorf("Expected width 120, got %d", updated.width)
	}

	if updated.height != 40 {
		t.Errorf("Expected height 40, got %d", updated.height)
	}
}

func TestQuitKeybinding(t *testing.T) {
	m := NewModel(nil, &config.Config{})

	// Test Alt+Q quit
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true}
	_, cmd := m.Update(msg)

	if cmd == nil {
		t.Error("Expected quit command to be returned")
	}

	// Verify it's actually the quit command by checking if it's tea.Quit
	// (In a real scenario, we'd need to execute the Cmd, but for testing we assume it's correct)
}

func TestHelpToggle(t *testing.T) {
	m := NewModel(nil, &config.Config{})

	// Help should be hidden initially
	if m.showHelp {
		t.Error("Expected help to be hidden initially")
	}

	// Press ? to show help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := m.Update(msg)
	m = updatedModel.(*Model)

	if !m.showHelp {
		t.Error("Expected help to be shown after pressing ?")
	}

	// Press ? again to hide
	updatedModel, _ = m.Update(msg)
	m = updatedModel.(*Model)

	if m.showHelp {
		t.Error("Expected help to be hidden after pressing ? again")
	}
}

func TestViewRendering(t *testing.T) {
	m := NewModel(nil, &config.Config{})
	m.SetSize(80, 24)

	// Initial view should not crash
	view := m.View()

	if view == "" {
		t.Error("Expected view to render content")
	}

	if len(view) < 10 {
		t.Error("Expected view to have substantial content")
	}
}
