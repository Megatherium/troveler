package tui

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/db"
	"troveler/internal/search"
	"troveler/tui/panels"
)

func TestSearchCompletionsIgnoreOlderRequests(t *testing.T) {
	for _, staleError := range []bool{false, true} {
		for _, newestFirst := range []bool{false, true} {
			t.Run(fmt.Sprintf("error=%v/newestFirst=%v", staleError, newestFirst), func(t *testing.T) {
				m, _, current := populatedSearchResultsModel(t)
				var older tea.Msg
				if staleError {
					older = failedSearchForTest(t, m, "name=Previous")
				} else {
					older = m.performSearch("name=Previous")()
				}
				newest := m.performSearch("name=Current")()
				if newestFirst {
					_, _ = m.Update(newest)
					assertCurrentSearchSelection(t, m, current)
					markSelectionForTest(t, m)
				}
				before := visibleSearchState(m)
				_, cmd := m.Update(older)
				if cmd != nil || !reflect.DeepEqual(visibleSearchState(m), before) {
					t.Error("obsolete completion changed results, panels, marks, error or searching state")
				}
				if !newestFirst {
					_, _ = m.Update(newest)
					assertCurrentSearchSelection(t, m, current)
				}
			})
		}
	}
}

func TestRepeatedSearchQueryKeepsNewestSnapshot(t *testing.T) {
	m, previous, _ := populatedSearchResultsModel(t)
	older := m.performSearch("name=Previous")()
	previous.Tagline = "refreshed snapshot"
	newCommand := "go install example.invalid/refreshed-fixture@latest"
	if err := m.db.SaveToolSnapshot(context.Background(), &previous, []db.InstallInstruction{{
		ID: "refreshed-install", Platform: "go", Command: newCommand,
	}}); err != nil {
		t.Fatal(err)
	}
	newest := m.performSearch("name=Previous")()
	_, _ = m.Update(newest)
	if m.selectedTool == nil || m.selectedTool.Tagline != previous.Tagline ||
		m.installPanel.GetSelectedCommand() != newCommand || m.searching {
		t.Fatal("latest repeated query did not apply refreshed metadata and instructions")
	}
	before := visibleSearchState(m)
	_, _ = m.Update(older)
	if !reflect.DeepEqual(visibleSearchState(m), before) {
		t.Fatal("identical query string allowed older tool metadata to replace the newest snapshot")
	}
}

func TestSearchCompletionInvalidatedBeforeNextDispatch(t *testing.T) {
	for _, edit := range []string{"type", "escape", "backspace", "round trip"} {
		for _, staleError := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/error=%v", edit, staleError), func(t *testing.T) {
				m, _, _ := populatedSearchResultsModel(t)
				m.activePanel = PanelSearch
				m.searchPanel.Focus()
				_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
				var older tea.Msg
				if staleError {
					older = failedSearchForTest(t, m, "name=Previous")
				} else {
					older = m.performSearch("name=Previous")()
				}
				switch edit {
				case "type":
					_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
				case "escape":
					_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
				case "backspace":
					_, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				case "round trip":
					_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
					_, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
				}
				// Do not execute any new timer/trigger: the old response must become
				// obsolete as soon as the input changes, before another DB request.
				m.err = errors.New("unrelated current error")
				before := visibleSearchState(m)
				_, _ = m.Update(older)
				if !reflect.DeepEqual(visibleSearchState(m), before) {
					t.Fatal("completion applied during the debounce window or after clearing input")
				}
			})
		}
	}
}

func TestLatestSearchErrorSurvivesOlderSuccess(t *testing.T) {
	m, _, _ := populatedSearchResultsModel(t)
	older := m.performSearch("name=Current")()
	failure := failedSearchForTest(t, m, "name=Previous")
	_, _ = m.Update(failure)
	if m.searching || !errors.Is(m.err, failure.err) {
		t.Fatal("latest search failure did not finish loading and report its error")
	}
	before := visibleSearchState(m)
	_, _ = m.Update(older)
	if !reflect.DeepEqual(visibleSearchState(m), before) {
		t.Fatal("older success replaced state controlled by the latest failed search")
	}
}

func TestSuccessfulSearchRetryClearsOnlySearchError(t *testing.T) {
	for _, unrelated := range []bool{false, true} {
		t.Run(fmt.Sprintf("unrelated=%v", unrelated), func(t *testing.T) {
			m, _, current := populatedSearchResultsModel(t)
			failure := failedSearchForTest(t, m, "name=Previous")
			_, _ = m.Update(failure)
			newest := m.performSearch("name=Current")()
			var want error
			if unrelated {
				want = errors.New("install error after search started")
				m.SetError(want)
			}
			_, _ = m.Update(newest)
			if m.err != want || m.searching || m.selectedTool == nil || m.selectedTool.ID != current.ID {
				t.Fatalf("successful retry: err=%v, searching=%v, selected=%v", m.err, m.searching, m.selectedTool)
			}
		})
	}
}

func TestInitialSearchCannotReplaceLaterInput(t *testing.T) {
	m, _, current := populatedSearchResultsModel(t)
	initial := m.Init()()
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	newest := m.performSearch("name=Current")()
	_, _ = m.Update(newest)
	assertCurrentSearchSelection(t, m, current)
	before := visibleSearchState(m)
	_, _ = m.Update(initial)
	if !reflect.DeepEqual(visibleSearchState(m), before) {
		t.Fatal("startup search replaced the later input's results")
	}
}

func TestSearchRequestCapturedBeforeCommandRuns(t *testing.T) {
	for _, staleError := range []bool{false, true} {
		t.Run(fmt.Sprintf("error=%v", staleError), func(t *testing.T) {
			m, _, current := populatedSearchResultsModel(t)
			older := m.performSearch("name=Previous")
			if staleError {
				older = failedSearchCommandForTest(t, m, "name=Previous")
			}
			newest := m.performSearch("name=Current")
			// Run the old background work after the latest request is registered,
			// while that latest request is still pending.
			completion := older()
			if staleError {
				if _, ok := completion.(searchErrorMsg); !ok {
					t.Fatal("queued command read a replacement search service")
				}
			}
			before := visibleSearchState(m)
			_, _ = m.Update(completion)
			if !reflect.DeepEqual(visibleSearchState(m), before) || !m.searching {
				t.Fatal("background command acquired the latest request's identity")
			}
			_, _ = m.Update(newest())
			assertCurrentSearchSelection(t, m, current)
		})
	}
}

func TestSearchCompletionWithoutOriginIsIgnored(t *testing.T) {
	for _, msg := range []tea.Msg{
		searchResultMsg{query: "name=Previous"},
		searchErrorMsg{err: errors.New("unversioned error")},
	} {
		t.Run(fmt.Sprintf("%T", msg), func(t *testing.T) {
			m, _, current := populatedSearchResultsModel(t)
			newest := m.performSearch("name=Current")
			before := visibleSearchState(m)
			_, _ = m.Update(msg)
			if !reflect.DeepEqual(visibleSearchState(m), before) {
				t.Fatal("completion without an origin changed the pending request's state")
			}
			_, _ = m.Update(newest())
			assertCurrentSearchSelection(t, m, current)
		})
	}
}

func TestSearchCompletionConsumedOnce(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(fmt.Sprintf("error=%v", failure), func(t *testing.T) {
			m, _, _ := populatedSearchResultsModel(t)
			var completion tea.Msg
			if failure {
				completion = failedSearchForTest(t, m, "name=Current")
			} else {
				completion = m.performSearch("name=Current")()
			}
			_, _ = m.Update(completion)
			if m.searching {
				t.Fatal("completion did not finish loading")
			}
			m.SetError(errors.New("subsequent action error"))
			before := visibleSearchState(m)
			_, _ = m.Update(completion)
			if !reflect.DeepEqual(visibleSearchState(m), before) {
				t.Fatal("duplicate completion changed selection or subsequent errors")
			}
		})
	}
}

func TestInitialSearchAppliesCurrentCompletion(t *testing.T) {
	m, _, _ := populatedSearchResultsModel(t)
	queued := m.Init()
	if !m.searching || m.searchPanel.Generation() != 0 {
		t.Fatal("startup request should be loading with an untouched input generation")
	}
	_, _ = m.Update(queued())
	if m.searching || m.err != nil || len(m.tools) != 2 || m.selectedTool == nil ||
		m.selectedTool.ID != m.tools[0].ID {
		t.Fatal("current startup completion did not populate sorted results and selection")
	}
}

func populatedSearchResultsModel(t *testing.T) (*Model, db.Tool, db.Tool) {
	t.Helper()
	m, previous, current := populatedSelectionModel(t)
	if err := m.db.SaveToolSnapshot(context.Background(), &current, []db.InstallInstruction{{
		ID: "current-install", Platform: "go", Command: "go install example.invalid/current-fixture@latest",
	}}); err != nil {
		t.Fatal(err)
	}
	return m, previous, current
}

func failedSearchForTest(t *testing.T, m *Model, query string) searchErrorMsg {
	t.Helper()
	msg := failedSearchCommandForTest(t, m, query)()
	failure, ok := msg.(searchErrorMsg)
	if !ok || failure.err == nil {
		t.Fatalf("closed database did not return an actual search failure: %#v", msg)
	}
	return failure
}

func failedSearchCommandForTest(t *testing.T, m *Model, query string) tea.Cmd {
	t.Helper()
	failingDB, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := failingDB.Close(); err != nil {
		t.Fatal(err)
	}
	workingService := m.searchService
	m.searchService = search.NewService(failingDB)
	cmd := m.performSearch(query)
	m.searchService = workingService
	return cmd
}

func assertCurrentSearchSelection(t *testing.T, m *Model, current db.Tool) {
	t.Helper()
	if m.searching || m.err != nil || len(m.tools) != 1 || m.tools[0].ID != current.ID ||
		m.selectedTool == nil || m.selectedTool.ID != current.ID ||
		m.installPanel.GetSelectedCommand() != "go install example.invalid/current-fixture@latest" {
		t.Fatalf("latest results did not populate current selection: tools=%v, selected=%v, err=%v", m.tools, m.selectedTool, m.err)
	}
}

type searchStateSnapshot struct {
	tools             []db.SearchResult
	selected          db.Tool
	cursorID          string
	installs          []db.InstallInstruction
	command           string
	installGeneration uint64
	info              string
	marked            []db.SearchResult
	searching         bool
	err               error
}

func visibleSearchState(m *Model) searchStateSnapshot {
	snapshot := searchStateSnapshot{
		tools: slices.Clone(m.tools), installs: slices.Clone(m.installs),
		command: m.installPanel.GetSelectedCommand(), info: m.infoPanel.View(),
		marked: m.toolsPanel.GetMarkedTools(), searching: m.searching, err: m.err,
	}
	if m.selectedTool != nil {
		snapshot.selected = *m.selectedTool
	}
	if tool := m.toolsPanel.GetSelectedTool(); tool != nil {
		snapshot.cursorID = tool.ID
	}
	if cmd := m.installPanel.InstallRequest(false); cmd != nil {
		snapshot.installGeneration = cmd().(panels.InstallExecuteMsg).SelectionGeneration
	}
	return snapshot
}

// Existing selection/mark fixtures replace the tool payload but obtain its
// request identity from a real command. Closed-DB fixtures reuse the error's
// origin to exercise instruction lookup failure after a result is delivered.
func searchResultWithOriginForTest(m *Model, fixture searchResultMsg) searchResultMsg {
	switch completion := m.performSearch(fixture.query)().(type) {
	case searchResultMsg:
		fixture.request = completion.request
	case searchErrorMsg:
		fixture.request = completion.request
	default:
		panic("search command returned an unexpected message")
	}
	return fixture
}
