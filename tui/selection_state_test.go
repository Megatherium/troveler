package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/tui/panels"
)

func TestEmptySearchClearsSelectionAndInstallActions(t *testing.T) {
	for _, allocated := range []bool{false, true} {
		for _, marked := range []bool{false, true} {
			t.Run(fmt.Sprintf("allocated=%v/marked=%v", allocated, marked), func(t *testing.T) {
				m, previous, _ := populatedSelectionModel(t)
				queued := queuedSelectionInstalls(m)
				if marked {
					m.activePanel = PanelTools
					m.toolsPanel.Focus()
					_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
					if cmd == nil {
						t.Fatal("expected mark action for the selected tool")
					}
					_, _ = m.Update(cmd())
					if m.toolsPanel.GetMarkedCount() != 1 {
						t.Fatal("expected populated batch selection before empty search")
					}
				}
				var empty []db.SearchResult
				if allocated {
					empty = []db.SearchResult{}
				}
				m.searching = true
				_, _ = m.Update(searchResultWithOriginForTest(m, searchResultMsg{tools: empty, query: "no-matching-fixture"}))

				if m.searching || len(m.tools) != 0 || m.toolsPanel.GetSelectedTool() != nil {
					t.Error("empty search should finish with an empty tools list")
				}
				if m.selectedTool != nil || len(m.installs) != 0 {
					t.Errorf("retained selection or installs: tool=%v, installs=%v", m.selectedTool, m.installs)
				}
				if m.toolsPanel.GetMarkedCount() != 0 {
					t.Error("empty search retained batch selection")
				}
				infoView := m.infoPanel.View()
				if !strings.Contains(infoView, "Select a tool to view details") || strings.Contains(infoView, previous.Name) {
					t.Errorf("stale info panel: %q", infoView)
				}
				if !strings.Contains(m.installPanel.View(), "Select a tool to see install options") {
					t.Errorf("stale install panel: %q", m.installPanel.View())
				}
				assertNoSelectionInstallActions(t, m, queued...)
				m.activePanel = PanelTools
				_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
				if m.modals.IsInfoShown() {
					t.Error("empty search still allows opening the previous tool's info modal")
				}

				// A later populated search must rebuild the cleared selection.
				_, _ = m.Update(searchResultWithOriginForTest(m, searchResultMsg{tools: []db.SearchResult{{Tool: previous}}}))
				if m.selectedTool == nil || m.selectedTool.ID != previous.ID || len(m.installs) != 1 ||
					m.installPanel.GetSelectedCommand() != selectionFixtureCommand {
					t.Fatal("selection and commands did not recover after an empty search")
				}
			})
		}
	}
}

func TestSelectionLookupFailureClearsPreviousCommands(t *testing.T) {
	for _, fromSearch := range []bool{false, true} {
		t.Run(fmt.Sprintf("fromSearch=%v", fromSearch), func(t *testing.T) {
			m, previous, current := populatedSelectionModel(t)
			queued := queuedSelectionInstalls(m)
			if err := m.db.Close(); err != nil {
				t.Fatal(err)
			}
			updateSelectionForTest(m, current, fromSearch)
			assertCurrentSelectionWithoutCommands(t, m, previous, current, queued...)
			if m.err == nil || !strings.Contains(m.err.Error(), "install instructions") {
				t.Errorf("expected instruction lookup error to be reported, got %v", m.err)
			}
		})
	}
}

func TestSelectionWithoutInstructionsClearsPreviousCommands(t *testing.T) {
	for _, fromSearch := range []bool{false, true} {
		t.Run(fmt.Sprintf("fromSearch=%v", fromSearch), func(t *testing.T) {
			m, previous, current := populatedSelectionModel(t)
			queued := queuedSelectionInstalls(m)
			updateSelectionForTest(m, current, fromSearch)
			assertCurrentSelectionWithoutCommands(t, m, previous, current, queued...)
			if m.err != nil {
				t.Errorf("empty install instructions should not report a lookup error: %v", m.err)
			}
		})
	}
}

const selectionFixtureCommand = "go install example.invalid/previous-fixture@latest"

func populatedSelectionModel(t *testing.T) (*Model, db.Tool, db.Tool) {
	t.Helper()
	m := newTestModelWithDB(t)
	previous := db.Tool{ID: "previous-fixture", Slug: "previous-fixture", Name: "Previous fixture", Language: "Go"}
	current := db.Tool{ID: "current-fixture", Slug: "current-fixture", Name: "Current fixture", Language: "Rust"}
	if err := m.db.SaveToolSnapshot(context.Background(), &previous, []db.InstallInstruction{{
		ID: "previous-install", Platform: "go", Command: selectionFixtureCommand,
	}}); err != nil {
		t.Fatal(err)
	}
	if err := m.db.SaveToolSnapshot(context.Background(), &current, nil); err != nil {
		t.Fatal(err)
	}
	// Force a deterministic fallback selection, independent of the host OS.
	m.installPanel = panels.NewInstallPanel("unmatched-fixture-platform", "", "")
	m.infoPanel.SetSize(80, 20)
	_, _ = m.Update(searchResultWithOriginForTest(m, searchResultMsg{tools: []db.SearchResult{{Tool: previous}}}))
	if m.selectedTool == nil || m.selectedTool.ID != previous.ID || len(m.installs) != 1 ||
		m.installPanel.GetSelectedCommand() != selectionFixtureCommand || !m.installPanel.IsFallbackMode() ||
		!strings.Contains(m.infoPanel.View(), previous.Name) {
		t.Fatal("fixture must populate a real tool, its details, instructions and fallback install command")
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true})
	if cmd == nil {
		t.Fatal("fixture install action must be available before clearing selection")
	}
	msg, ok := cmd().(panels.InstallExecuteMsg)
	if !ok || msg.Command != selectionFixtureCommand {
		t.Fatalf("expected fixture install action, got %#v", msg)
	}
	return m, previous, current
}

func updateSelectionForTest(m *Model, current db.Tool, fromSearch bool) {
	if fromSearch {
		_, _ = m.Update(searchResultWithOriginForTest(m, searchResultMsg{tools: []db.SearchResult{{Tool: current}}}))
	} else {
		_, _ = m.Update(panels.ToolCursorChangedMsg{Tool: db.SearchResult{Tool: current}})
	}
}

func queuedSelectionInstalls(m *Model) []tea.Msg {
	return []tea.Msg{m.installPanel.InstallRequest(false)(), m.installPanel.InstallRequest(true)()}
}

func assertCurrentSelectionWithoutCommands(t *testing.T, m *Model, previous, current db.Tool, queued ...tea.Msg) {
	t.Helper()
	if m.selectedTool == nil || m.selectedTool.ID != current.ID || len(m.installs) != 0 {
		t.Errorf("current selection retained old instructions: tool=%v, installs=%v", m.selectedTool, m.installs)
	}
	infoView := m.infoPanel.View()
	if !strings.Contains(infoView, current.Name) || strings.Contains(infoView, previous.Name) {
		t.Errorf("wrong info panel after selection changed: %q", infoView)
	}
	assertNoSelectionInstallActions(t, m, queued...)
}

func assertNoSelectionInstallActions(t *testing.T, m *Model, queued ...tea.Msg) {
	t.Helper()
	if m.installPanel.HasCommands() || m.installPanel.GetSelectedCommand() != "" || m.installPanel.IsFallbackMode() {
		t.Error("install panel retained previous commands or fallback state")
	}
	for _, panel := range []PanelID{PanelSearch, PanelTools, PanelInstall} {
		m.activePanel = panel
		for _, action := range []rune{'i', 'm'} {
			_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{action}, Alt: true})
			if cmd != nil || m.executing || m.modals.IsInstallShown() || m.modals.IsBatchConfigShown() {
				t.Errorf("panel %v: Alt+%c still offers an install action", panel, action)
			}
		}
	}
	m.installPanel.Focus()
	for _, action := range []rune{'i', 'm'} {
		_, cmd := m.installPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{action}, Alt: true})
		if cmd != nil {
			t.Errorf("install panel still emits an Alt+%c command", action)
		}
	}
	// A command message queued before the selection changed must also be inert.
	for _, msg := range queued {
		_, cmd := m.Update(msg)
		if cmd != nil || m.executing || m.modals.IsInstallShown() {
			t.Errorf("queued %T still starts an install after commands were cleared", msg)
		}
	}
}
