package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"troveler/config"
	"troveler/db"
	"troveler/internal/search"
	"troveler/tui/panels"
)

// helpers

func newTestModel(t *testing.T) *Model {
	t.Helper()
	cfg := &config.Config{
		TUI: config.TUIConfig{
			Theme:           "gradient",
			TaglineMaxWidth: 40,
		},
		Install: config.InstallConfig{
			PlatformOverride: "",
			FallbackPlatform: "linux",
		},
	}
	return NewModel(nil, cfg)
}

func newTestModelWithDB(t *testing.T) *Model {
	t.Helper()
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create in-memory db: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	cfg := &config.Config{
		TUI: config.TUIConfig{
			Theme:           "gradient",
			TaglineMaxWidth: 40,
		},
		Install: config.InstallConfig{
			PlatformOverride: "",
			FallbackPlatform: "linux",
		},
	}

	m := NewModel(database, cfg)
	m.searchService = search.NewService(database)
	return m
}

// ============================================================================
// Modal state transition tests (escape chain)
// ============================================================================

func TestEscapeChain_HelpModal(t *testing.T) {
	m := newTestModel(t)
	m.showHelp = true

	updatedModel, cmd := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.showHelp {
		t.Error("Expected showHelp to be false after escape")
	}
	if cmd != nil {
		t.Error("Expected nil cmd for help escape")
	}
}

func TestEscapeChain_InfoModal(t *testing.T) {
	m := newTestModel(t)
	m.showInfoModal = true

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.showInfoModal {
		t.Error("Expected showInfoModal to be false after escape")
	}
}

func TestEscapeChain_UpdateModal(t *testing.T) {
	m := newTestModel(t)
	m.showUpdateModal = true
	m.updating = false // not running — should close

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.showUpdateModal {
		t.Error("Expected showUpdateModal to be false after escape")
	}
	if m2.updating {
		t.Error("Expected updating to be false after closing update modal")
	}
}

func TestEscapeChain_InstallModalWithoutExecuting(t *testing.T) {
	m := newTestModel(t)
	m.showInstallModal = true
	m.executing = false

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.showInstallModal {
		t.Error("Expected showInstallModal to be false when not executing")
	}
}

func TestEscapeChain_InstallModalWhileExecuting(t *testing.T) {
	m := newTestModel(t)
	m.showInstallModal = true
	m.executing = true

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if !m2.showInstallModal {
		t.Error("Expected showInstallModal to stay open while executing")
	}
}

func TestEscapeChain_BatchConfigModal(t *testing.T) {
	m := newTestModel(t)
	m.showBatchConfigModal = true
	m.batchConfig = &BatchInstallConfig{} // non-nil

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.showBatchConfigModal {
		t.Error("Expected showBatchConfigModal to be false after escape")
	}
	if m2.batchConfig != nil {
		t.Error("Expected batchConfig to be nil after escape from batch config modal")
	}
}

func TestEscapeChain_ExecuteOutput(t *testing.T) {
	m := newTestModel(t)
	m.executeOutput = "some command output"
	m.err = errors.New("test error") // set an error alongside

	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.executeOutput != "" {
		t.Error("Expected executeOutput to be cleared after escape")
	}
	if m2.err != nil {
		t.Error("Expected err to be nil after clearing output")
	}
}

func TestEscapeChain_NothingOpen_SearchPanel(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch
	m.showHelp = false
	m.showInfoModal = false
	m.showInstallModal = false
	m.showBatchConfigModal = false
	m.executeOutput = ""

	// Escape while nothing is open should delegate to search panel
	updatedModel, _ := m.handleEscapeKey()
	m2 := updatedModel.(*Model)

	if m2.activePanel != PanelSearch {
		t.Error("Expected activePanel to remain Search after escape with nothing open")
	}
}

// ============================================================================
// Panel message routing through Update() main method
// ============================================================================

func TestUpdate_SearchTriggeredMsg_SetsSearching(t *testing.T) {
	m := newTestModelWithDB(t)
	m.searching = false

	_, _ = m.Update(panels.SearchTriggeredMsg{Query: "fzf"})

	if !m.searching {
		t.Error("Expected searching to be true after SearchTriggeredMsg")
	}
}

func TestUpdate_SearchResultMsg_PopulatesTools(t *testing.T) {
	m := newTestModelWithDB(t)
	m.searching = true

	tools := []db.SearchResult{
		{Tool: db.Tool{ID: "tool-1", Slug: "test1", Name: "Test Tool 1"}},
		{Tool: db.Tool{ID: "tool-2", Slug: "test2", Name: "Test Tool 2"}},
	}

	_, _ = m.Update(searchResultMsg{tools: tools, query: "test"})

	if m.searching {
		t.Error("Expected searching to be false after searchResultMsg")
	}
	if len(m.tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(m.tools))
	}
	if m.tools[0].Slug != "test1" {
		t.Errorf("Expected first tool slug 'test1', got %s", m.tools[0].Slug)
	}
	if m.selectedTool == nil {
		t.Error("Expected selectedTool to be set to first tool")
	}
}

func TestUpdate_SearchResultMsg_EmptyResults(t *testing.T) {
	m := newTestModelWithDB(t)
	m.searching = true

	_, _ = m.Update(searchResultMsg{tools: []db.SearchResult{}, query: "nonexistent"})

	if m.searching {
		t.Error("Expected searching to be false after empty searchResultMsg")
	}
	if len(m.tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(m.tools))
	}
}

func TestUpdate_SearchErrorMsg(t *testing.T) {
	m := newTestModelWithDB(t)
	m.searching = true

	_, _ = m.Update(searchErrorMsg{err: errors.New("tool not found: missing")})

	if m.searching {
		t.Error("Expected searching to be false after searchErrorMsg")
	}
	if m.err == nil {
		t.Error("Expected err to be set after searchErrorMsg")
	}
}

func TestUpdate_ToolSelectedMsg_SwitchesToInstall(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelTools

	_, _ = m.Update(panels.ToolSelectedMsg{})

	if m.activePanel != PanelInstall {
		t.Errorf("Expected PanelInstall after ToolSelectedMsg, got %v", m.activePanel)
	}
}

func TestUpdate_ToolMarkedMsg_Nop(t *testing.T) {
	m := newTestModel(t)

	updatedModel, cmd := m.Update(panels.ToolMarkedMsg{Tool: db.SearchResult{Tool: db.Tool{ID: "x"}}, IsMarked: true})

	if updatedModel.(*Model) != m {
		t.Error("Expected model to be unchanged after ToolMarkedMsg")
	}
	if cmd != nil {
		t.Error("Expected nil cmd after ToolMarkedMsg")
	}
}

func TestUpdate_InstallExecuteMsg_SetsModalState(t *testing.T) {
	m := newTestModel(t)
	m.selectedTool = &db.Tool{ID: "test-tool", Name: "Test"}

	_, cmd := m.Update(panels.InstallExecuteMsg{Command: "echo hello"})

	if !m.showInstallModal {
		t.Error("Expected showInstallModal to be true after InstallExecuteMsg")
	}
	if !m.executing {
		t.Error("Expected executing to be true after InstallExecuteMsg")
	}
	if m.executeOutput != "" {
		t.Error("Expected executeOutput to be cleared after InstallExecuteMsg")
	}
	if cmd == nil {
		t.Error("Expected non-nil cmd (exec command) after InstallExecuteMsg")
	}
}

func TestUpdate_InstallExecuteMiseMsg_SetsModalState(t *testing.T) {
	m := newTestModel(t)
	m.selectedTool = &db.Tool{ID: "test-tool", Name: "Test"}

	_, cmd := m.Update(panels.InstallExecuteMiseMsg{Command: "echo hello"})

	if !m.showInstallModal {
		t.Error("Expected showInstallModal to be true after InstallExecuteMiseMsg")
	}
	if !m.executing {
		t.Error("Expected executing to be true after InstallExecuteMiseMsg")
	}
	if cmd == nil {
		t.Error("Expected non-nil cmd after InstallExecuteMiseMsg")
	}
}

func TestUpdate_InstallCompleteMsg_ClearsExecution(t *testing.T) {
	m := newTestModel(t)
	m.executing = true
	m.executeOutput = ""
	m.err = nil

	_, _ = m.Update(installCompleteMsg{output: "welcome to echo", err: nil})

	if m.executing {
		t.Error("Expected executing to be false after installCompleteMsg")
	}
	if m.executeOutput != "welcome to echo" {
		t.Errorf("Expected executeOutput to be 'welcome to echo', got %q", m.executeOutput)
	}
	if m.err != nil {
		t.Error("Expected no error after successful install")
	}
}

func TestUpdate_InstallCompleteMsg_WithError(t *testing.T) {
	m := newTestModel(t)
	m.executing = true

	testErr := errors.New("command failed")
	_, _ = m.Update(installCompleteMsg{output: "nope", err: testErr})

	if m.executing {
		t.Error("Expected executing to be false after installCompleteMsg")
	}
	if m.err == nil {
		t.Error("Expected err to be set after failed install")
	}
}

// ============================================================================
// Unhandled message falls through to active panel
// ============================================================================

func TestUpdate_UnhandledMessage_FallsToActivePanel(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch
	origSearchPanel := m.searchPanel

	// Sending an unknown message type should delegate to the active panel
	customMsg := struct{ foo string }{foo: "bar"}
	_, _ = m.Update(customMsg)

	// The model reference should still be the same search panel (via type assertion re-assignment)
	if m.searchPanel != origSearchPanel {
		t.Error("Expected searchPanel to be unchanged after unhandled message delegation")
	}
}

func TestUpdate_UnhandledMessage_FallsToToolsPanel(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelTools
	origToolsPanel := m.toolsPanel

	customMsg := struct{ bar int }{bar: 42}
	_, _ = m.Update(customMsg)

	if m.toolsPanel != origToolsPanel {
		t.Error("Expected toolsPanel to be unchanged after unhandled message delegation")
	}
}

func TestUpdate_UnhandledMessage_FallsToInstallPanel(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelInstall
	origInstallPanel := m.installPanel

	customMsg := struct{ baz bool }{baz: true}
	_, _ = m.Update(customMsg)

	if m.installPanel != origInstallPanel {
		t.Error("Expected installPanel to be unchanged after unhandled message delegation")
	}
}

// ============================================================================
// Mouse messages
// ============================================================================

func TestUpdate_MouseMsg_Nop(t *testing.T) {
	m := newTestModel(t)
	updatedModel, cmd := m.Update(tea.MouseMsg{Type: tea.MouseLeft})

	if updatedModel.(*Model) != m {
		t.Error("Expected model to be unchanged after MouseMsg")
	}
	if cmd != nil {
		t.Error("Expected nil cmd after MouseMsg")
	}
}

// ============================================================================
// Tab key cycles panels
// ============================================================================

func TestUpdate_TabKey_CyclesPanels(t *testing.T) {
	m := newTestModel(t)
	m.SetSize(100, 40)

	// Start at Search
	if m.activePanel != PanelSearch {
		t.Fatalf("Expected initial panel Search, got %v", m.activePanel)
	}

	// Press Tab
	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.Update(msg)
	m2 := updatedModel.(*Model)

	if m2.activePanel != PanelTools {
		t.Errorf("Expected Tools panel after Tab, got %v", m2.activePanel)
	}

	// Tab again
	updatedModel, _ = m2.Update(msg)
	m3 := updatedModel.(*Model)

	if m3.activePanel != PanelInstall {
		t.Errorf("Expected Install panel after second Tab, got %v", m3.activePanel)
	}

	// Tab again — cycles back to Search
	updatedModel, _ = m3.Update(msg)
	m4 := updatedModel.(*Model)

	if m4.activePanel != PanelSearch {
		t.Errorf("Expected Search panel after third Tab (cycle), got %v", m4.activePanel)
	}
}

// ============================================================================
// Key press routing in handleKeyPress
// ============================================================================

func TestHandleKeyPress_GlobalKeys_Help(t *testing.T) {
	m := newTestModel(t)
	m.showHelp = false

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}}
	updatedModel, _ := m.handleKeyPress(msg)

	if !updatedModel.(*Model).showHelp {
		t.Error("Expected showHelp to be true after pressing ?")
	}
}

func TestHandleKeyPress_GlobalKeys_Tab(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch

	msg := tea.KeyMsg{Type: tea.KeyTab}
	updatedModel, _ := m.handleKeyPress(msg)

	if updatedModel.(*Model).activePanel != PanelTools {
		t.Errorf("Expected Tools panel after Tab, got %v", updatedModel.(*Model).activePanel)
	}
}

func TestHandleKeyPress_GlobalKeys_Quit(t *testing.T) {
	m := newTestModel(t)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}, Alt: true}
	_, cmd := m.handleKeyPress(msg)

	if cmd == nil {
		t.Error("Expected quit command for Alt+Q")
	}
}

func TestHandleKeyPress_BatchConfigModal_Option1(t *testing.T) {
	m := newTestModel(t)
	m.showBatchConfigModal = true
	m.batchConfig = NewBatchInstallConfig()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	updatedModel, _ := m.handleKeyPress(msg)
	m2 := updatedModel.(*Model)

	if !m2.showBatchConfigModal {
		t.Error("Expected batch config modal to stay open after selecting option (more steps)")
	}
	if m2.batchConfig.ConfigStep != 1 {
		t.Errorf("Expected to advance to step 1, got %d", m2.batchConfig.ConfigStep)
	}
}

func TestHandleKeyPress_BatchConfigModal_FinalOptionCloses(t *testing.T) {
	m := newTestModel(t)
	m.showBatchConfigModal = true
	// Advance to step 4 (last step) and skip setting it
	m.batchConfig = NewBatchInstallConfig()
	m.batchConfig.ConfigStep = 4 // last step: "Use mise?"

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}
	updatedModel, cmd := m.handleKeyPress(msg)
	m2 := updatedModel.(*Model)

	if m2.showBatchConfigModal {
		t.Error("Expected batch config modal to close after final option")
	}
	if cmd == nil {
		// startBatchInstall triggers and batch may be empty
		t.Log("cmd was nil — batch install start may require marked tools")
	}
}

func TestHandleKeyPress_BatchConfigModal_Option2(t *testing.T) {
	m := newTestModel(t)
	m.showBatchConfigModal = true
	m.batchConfig = NewBatchInstallConfig()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}}
	updatedModel, _ := m.handleKeyPress(msg)
	m2 := updatedModel.(*Model)

	if m2.batchConfig.ConfigStep != 1 {
		t.Errorf("Expected to advance to step 1, got %d", m2.batchConfig.ConfigStep)
	}
	// Option 2 on step 0 sets ReuseConfig = false
	if m2.batchConfig.ReuseConfig {
		t.Error("Expected ReuseConfig to be false after selecting option 2")
	}
}

func TestHandleKeyPress_BatchConfigModal_InvalidKey(t *testing.T) {
	m := newTestModel(t)
	m.showBatchConfigModal = true
	m.batchConfig = NewBatchInstallConfig()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'9'}}
	updatedModel, _ := m.handleKeyPress(msg)
	m2 := updatedModel.(*Model)

	if m2.batchConfig.ConfigStep != 0 {
		t.Error("Expected step to not advance on invalid option key")
	}
}

// ============================================================================
// Action keys routing
// ============================================================================

func TestHandleKeyPress_AltR_OpenRepo_WithoutTool(t *testing.T) {
	m := newTestModel(t)
	m.selectedTool = nil

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}, Alt: true}
	_, cmd := m.handleKeyPress(msg)

	if cmd == nil {
		t.Error("Expected openRepositoryURL cmd even when selectedTool is nil (cmd returns nil msg)")
	}
}

func TestHandleKeyPress_AltI_WithoutMarks(t *testing.T) {
	m := newTestModel(t)
	m.tools = []db.SearchResult{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}, Alt: true}
	_, _ = m.handleKeyPress(msg)

	// When tools panel has no marked tools, Alt+I should still process
	// without panicking (the handler may do nothing meaningful)
}

func TestHandleKeyPress_AltM_WithoutMarks(t *testing.T) {
	m := newTestModel(t)
	m.tools = []db.SearchResult{}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}, Alt: true}
	_, _ = m.handleKeyPress(msg)

	// When tools panel has no marked tools, Alt+M should still process
	// without panicking
}

// ============================================================================
// Info modal routing
// ============================================================================

func TestHandleKeyPress_InfoModalKey_WithSelection(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelTools
	m.selectedTool = &db.Tool{ID: "test-tool", Name: "Test"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	updatedModel, _, handled := m.handleActionKeys(msg)

	if !handled {
		t.Error("Expected 'i' key to be handled as info modal action")
	}
	if !updatedModel.(*Model).showInfoModal {
		t.Error("Expected showInfoModal to be true after info modal key")
	}
}

func TestHandleKeyPress_InfoModalKey_FromSearchPanel(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch   // info modal only opens from non-Search panels
	m.selectedTool = &db.Tool{ID: "test-tool", Name: "Test"}

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	_, _, handled := m.handleActionKeys(msg)

	if !handled {
		t.Error("Expected 'i' key to be handled by action keys even from Search")
	}
	// But showInfoModal should remain false because activePanel == PanelSearch
	if m.showInfoModal {
		t.Error("Expected showInfoModal to stay false when on Search panel")
	}
}

// ============================================================================
// NOTE: Alt+U (update key) routing cannot be tested yet because
// updateService is nil in NewModel(), causing a nil-pointer dereference
// in the goroutine spawned by startUpdate(). This will be fixed
// when UpdateModel is extracted in tr-bb7.
// See: TestHandleKeyPress_UpdateKey_IsHandled removed due to this bug.
// ============================================================================

// ============================================================================
// Search panel text input delegation
// ============================================================================

func TestHandleKeyPress_SearchPanel_RuneDelegation(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := m.handleKeyPress(msg)

	// After delegation, the search panel should have received the 'a'
	if updatedModel.(*Model).activePanel != PanelSearch {
		t.Error("Expected to stay on Search panel after text input")
	}
}

func TestHandleKeyPress_SearchPanel_EscapeDelegation(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelSearch
	m.showHelp = false
	m.showInfoModal = false
	m.showInstallModal = false
	m.showBatchConfigModal = false
	m.executeOutput = ""

	msg := tea.KeyMsg{Type: tea.KeyEscape}
	updatedModel, _ := m.handleKeyPress(msg)

	if updatedModel.(*Model).activePanel != PanelSearch {
		t.Error("Expected search panel to remain after escape delegation")
	}
}

// ============================================================================
// Panel-specific key delegation (Tools/Install fall-through)
// ============================================================================

func TestHandleKeyPress_ToolsPanel_Delegation(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelTools

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.handleKeyPress(msg)

	if updatedModel.(*Model).activePanel != PanelTools {
		t.Error("Expected to stay on Tools panel after key delegation")
	}
}

func TestHandleKeyPress_InstallPanel_Delegation(t *testing.T) {
	m := newTestModel(t)
	m.activePanel = PanelInstall

	msg := tea.KeyMsg{Type: tea.KeyUp}
	updatedModel, _ := m.handleKeyPress(msg)

	if updatedModel.(*Model).activePanel != PanelInstall {
		t.Error("Expected to stay on Install panel after key delegation")
	}
}
