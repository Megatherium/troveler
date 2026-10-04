package tui

import (
	"context"
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/install"
	"troveler/tui/panels"
)

func TestQueuedInstallRequestRejectsSupersededSelection(t *testing.T) {
	for _, source := range []string{"global", "panel"} {
		for _, action := range []rune{'i', 'm'} {
			for _, change := range []string{"different tool", "same command", "round trip", "same tool refresh"} {
				t.Run(fmt.Sprintf("%s/Alt+%c/%s", source, action, change), func(t *testing.T) {
					m, previous, current := populatedSelectionModel(t)
					currentCommand := "go install example.invalid/current-fixture@latest"
					if change == "same command" {
						currentCommand = selectionFixtureCommand
					}
					if err := m.db.SaveToolSnapshot(context.Background(), &current, []db.InstallInstruction{{
						ID: "current-install", Platform: "go", Command: currentCommand,
					}}); err != nil {
						t.Fatal(err)
					}
					queued := installRequestFromSource(t, m, source, action)
					if change == "same tool refresh" {
						updateSelectionForTest(m, previous, true)
						current = previous
						currentCommand = selectionFixtureCommand
					} else {
						updateSelectionForTest(m, current, change == "same command")
						if change == "round trip" {
							updateSelectionForTest(m, previous, false)
							current = previous
							currentCommand = selectionFixtureCommand
						}
					}
					if m.selectedTool == nil || m.selectedTool.ID != current.ID ||
						m.installPanel.GetSelectedCommand() != currentCommand {
						t.Fatal("test requires another fully populated selection before delivery")
					}
					// Delay the background command itself until after selection changes:
					// it must have captured its origin when the key was handled.
					stale := queued()
					assertInstallRequestCommand(t, stale, action, selectionFixtureCommand)
					m.executeOutput = "previous output"
					_, execution := m.Update(stale)
					if execution != nil || m.executing || m.modals.IsInstallShown() || m.executeOutput != "previous output" {
						t.Fatal("superseded request started an install or changed execution state")
					}

					fresh := installRequestFromSource(t, m, source, action)()
					assertInstallRequestCommand(t, fresh, action, currentCommand)
					_, execution = m.Update(fresh)
					if execution == nil || !m.executing || !m.modals.IsInstallShown() || m.executeOutput != "" {
						t.Fatal("fresh request for the current selection did not queue an install")
					}
					// Do not invoke execution: fixture install commands must never run.
				})
			}
		}
	}
}

func TestInstallRequestWithoutSelectionOriginIsIgnored(t *testing.T) {
	for _, msg := range []tea.Msg{
		panels.InstallExecuteMsg{Command: selectionFixtureCommand},
		panels.InstallExecuteMiseMsg{Command: "mise use --global go:example.invalid/previous-fixture@latest"},
	} {
		t.Run(fmt.Sprintf("%T", msg), func(t *testing.T) {
			m, _, _ := populatedSelectionModel(t)
			_, execution := m.Update(msg)
			if execution != nil || m.executing || m.modals.IsInstallShown() {
				t.Fatal("request without its originating selection queued an install")
			}
		})
	}
}

func installRequestFromSource(t *testing.T, m *Model, source string, action rune) tea.Cmd {
	t.Helper()
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{action}, Alt: true}
	var cmd tea.Cmd
	if source == "global" {
		_, cmd = m.Update(key)
	} else {
		m.installPanel.Focus()
		_, cmd = m.installPanel.Update(key)
	}
	if cmd == nil {
		t.Fatal("populated selection must offer an install request")
	}
	return cmd
}

func assertInstallRequestCommand(t *testing.T, msg tea.Msg, action rune, command string) {
	t.Helper()
	if action == 'm' {
		request, ok := msg.(panels.InstallExecuteMiseMsg)
		if !ok || request.Command != install.TransformToMise(command) {
			t.Fatalf("wrong mise request: %#v", msg)
		}
		return
	}
	request, ok := msg.(panels.InstallExecuteMsg)
	if !ok || request.Command != command {
		t.Fatalf("wrong install request: %#v", msg)
	}
}
