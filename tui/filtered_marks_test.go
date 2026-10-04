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

func TestFilteredBatchMarksDispatch(t *testing.T) {
	for _, action := range []rune{'i', 'm'} {
		for _, scenario := range []string{"hidden only", "single install", "visible and hidden", "restored filter"} {
			t.Run(fmt.Sprintf("Alt+%c/%s", action, scenario), func(t *testing.T) {
				m, previous, current := populatedSelectionModel(t)
				m.width, m.height = 120, 40
				markSelectionForTest(t, m)
				currentCommand := "go install example.invalid/current-fixture@latest"
				if scenario == "single install" {
					if err := m.db.SaveToolSnapshot(context.Background(), &current, []db.InstallInstruction{{
						ID: "current-install", Platform: "go", Command: currentCommand,
					}}); err != nil {
						t.Fatal(err)
					}
				}
				_, _ = m.Update(searchResultMsg{tools: []db.SearchResult{{Tool: current}}})
				if !m.toolsPanel.IsMarked(previous.ID) {
					t.Fatal("nonempty search discarded the previous mark")
				}
				wantBatchID := ""
				if scenario == "visible and hidden" {
					markSelectionForTest(t, m)
					wantBatchID = current.ID
				} else if scenario == "restored filter" {
					_, _ = m.Update(searchResultMsg{tools: []db.SearchResult{{Tool: previous}, {Tool: current}}})
					wantBatchID = previous.ID
				}

				_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{action}, Alt: true})
				if wantBatchID == "" {
					if m.toolsPanel.GetMarkedCount() != 0 || len(m.toolsPanel.GetMarkedTools()) != 0 ||
						strings.Contains(m.renderStatusBar(), "marked") || m.modals.IsBatchConfigShown() ||
						m.batch.IsConfigActive() || m.executing {
						t.Error("hidden marks advertised or opened an empty batch install")
					}
					if scenario == "single install" {
						if cmd == nil {
							t.Fatal("hidden marks blocked the current tool's single install")
						}
						assertInstallRequestCommand(t, cmd(), action, currentCommand)
					} else if cmd != nil {
						t.Error("tool without instructions emitted an install command")
					}
					return
				}

				marked := m.toolsPanel.GetMarkedTools()
				if m.toolsPanel.GetMarkedCount() != 1 || len(marked) != 1 || marked[0].ID != wantBatchID ||
					!strings.Contains(m.renderStatusBar(), "1 marked") || !m.modals.IsBatchConfigShown() || cmd != nil {
					t.Fatal("visible batch marks, status count and shortcut dispatch disagree")
				}
				if !m.batch.IsConfigActive() || m.batch.Config().UseMise != (action == 'm') ||
					!strings.Contains(m.View(), "Batch Install Configuration (1 tools)") {
					t.Fatal("batch wizard mode or displayed target count is wrong")
				}
				// Finish the actual wizard, but never deliver the start message that
				// would queue ProcessTool or execute a fixture install command.
				for step := 0; step < m.batch.Config().ConfigStepCount(); step++ {
					_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
				}
				if cmd == nil {
					t.Fatal("completed wizard did not queue a batch")
				}
				start, ok := cmd().(batchInstallStartMsg)
				if !ok || len(start.tools) != 1 || start.tools[0].ID != wantBatchID ||
					!m.executing || !m.modals.IsInstallShown() || m.modals.IsBatchConfigShown() {
					t.Fatalf("wrong batch install targets or state: %#v", start)
				}
				progress := m.batch.Progress()
				if progress == nil || len(progress.Tools) != 1 || progress.Tools[0].ID != wantBatchID {
					t.Fatal("batch progress targets disagree with displayed count")
				}
			})
		}
	}
}

func TestEmptySearchClearsHiddenBatchMarks(t *testing.T) {
	m, previous, current := populatedSelectionModel(t)
	markSelectionForTest(t, m)
	_, _ = m.Update(searchResultMsg{tools: []db.SearchResult{{Tool: current}}})
	if !m.toolsPanel.IsMarked(previous.ID) {
		t.Fatal("test requires a retained hidden mark before the empty search")
	}
	_, _ = m.Update(searchResultMsg{})
	_, _ = m.Update(searchResultMsg{tools: []db.SearchResult{{Tool: previous}, {Tool: current}}})
	if m.toolsPanel.IsMarked(previous.ID) || m.toolsPanel.GetMarkedCount() != 0 ||
		len(m.toolsPanel.GetMarkedTools()) != 0 {
		t.Fatal("empty search allowed hidden marks to reappear when results returned")
	}
}

func markSelectionForTest(t *testing.T, m *Model) {
	t.Helper()
	m.activePanel = PanelTools
	m.toolsPanel.Focus()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if cmd == nil {
		t.Fatal("mark action did not emit a notification")
	}
	msg, ok := cmd().(panels.ToolMarkedMsg)
	if !ok || !msg.IsMarked || m.selectedTool == nil || msg.Tool.ID != m.selectedTool.ID {
		t.Fatalf("mark action targeted the wrong selection: %#v", msg)
	}
	_, _ = m.Update(msg)
}
