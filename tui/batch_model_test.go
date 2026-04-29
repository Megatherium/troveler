package tui

import (
	"context"
	"strings"
	"testing"

	"troveler/db"

	tea "github.com/charmbracelet/bubbletea"
)

// ============================================================================
// BatchInstallModel — config lifecycle
// ============================================================================

func TestBatchModel_StartBatchConfig_SetStep_ClearConfig(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	if batch.IsConfigActive() {
		t.Error("expected IsConfigActive false for fresh model")
	}
	if batch.Config() != nil {
		t.Error("expected nil Config for fresh model")
	}

	batch.StartBatchConfig(false)
	if !batch.IsConfigActive() {
		t.Error("expected IsConfigActive true after StartBatchConfig")
	}
	if batch.Config() == nil {
		t.Fatal("expected non-nil Config after StartBatchConfig")
	}
	if batch.Config().UseMise {
		t.Error("expected UseMise false when StartBatchConfig(false)")
	}
	if batch.Config().ConfigStep != 0 {
		t.Errorf("expected ConfigStep 0, got %d", batch.Config().ConfigStep)
	}

	batch.SetStep(3)
	if batch.Config().ConfigStep != 3 {
		t.Errorf("expected ConfigStep 3 after SetStep(3), got %d", batch.Config().ConfigStep)
	}

	batch.ClearConfig()
	if batch.IsConfigActive() {
		t.Error("expected IsConfigActive false after ClearConfig")
	}
	if batch.Config() != nil {
		t.Error("expected nil Config after ClearConfig")
	}
}

func TestBatchModel_StartBatchConfig_UseMise(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.StartBatchConfig(true)
	if !batch.Config().UseMise {
		t.Error("expected UseMise true when StartBatchConfig(true)")
	}
}

func TestBatchModel_SetStep_NilConfig(_ *testing.T) {
	batch := NewBatchInstallModel(nil)
	// Should not panic when config is nil
	batch.SetStep(3)
}

func TestBatchModel_ClearConfig_NilConfig(_ *testing.T) {
	batch := NewBatchInstallModel(nil)
	// Should not panic when config is already nil
	batch.ClearConfig()
}

func TestBatchModel_SetStepValue_AdvanceStep(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.StartBatchConfig(false)

	// Step 0: "Reuse configuration?" → option 0 = Yes → ReuseConfig stays true (default)
	batch.SetStepValue(0)
	if !batch.Config().ReuseConfig {
		t.Error("expected ReuseConfig true after option 0 on step 0")
	}
	if !batch.AdvanceStep() {
		t.Error("expected AdvanceStep to return true (not done)")
	}

	// Step 1: "Use sudo?" → option 1 = No → UseSudo false
	batch.SetStepValue(1)
	if batch.Config().UseSudo {
		t.Error("expected UseSudo false after option 1 on step 1")
	}
	batch.AdvanceStep()

	// Step 2: "Sudo only for system?" → option 0 = Yes → SudoOnlySystem true
	batch.SetStepValue(0)
	if !batch.Config().SudoOnlySystem {
		t.Error("expected SudoOnlySystem true after option 0 on step 2")
	}
	batch.AdvanceStep()

	// Step 3: "Skip if blind?" → option 1 = No → SkipIfBlind false
	batch.SetStepValue(1)
	if batch.Config().SkipIfBlind {
		t.Error("expected SkipIfBlind false after option 1 on step 3")
	}
	batch.AdvanceStep()

	// Step 4: "Use mise?" → final step
	batch.SetStepValue(0)
	if !batch.Config().UseMise {
		t.Error("expected UseMise true after option 0 on step 4")
	}
	if batch.AdvanceStep() {
		t.Error("expected AdvanceStep to return false after last step (wizard done)")
	}
}

func TestBatchModel_SetStepValue_NilConfig(_ *testing.T) {
	batch := NewBatchInstallModel(nil)
	// Should not panic when config is nil
	batch.SetStepValue(0)
}

func TestBatchModel_AdvanceStep_NilConfig(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	if batch.AdvanceStep() {
		t.Error("expected AdvanceStep to return false when config is nil")
	}
}

// ============================================================================
// BatchInstallModel — progress lifecycle
// ============================================================================

func TestBatchModel_Progress_IsRunning(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	if batch.IsRunning() {
		t.Error("expected IsRunning false for fresh model")
	}
	if batch.Progress() != nil {
		t.Error("expected nil Progress for fresh model")
	}
}

func TestBatchModel_HandleProgress_Completed(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools:        []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
		CurrentIndex: 0,
	}

	// Process the single tool as completed
	finished, nextIdx := batch.HandleProgress(batchInstallProgressMsg{
		toolID: "t1",
		output: "done",
	})

	if !finished {
		t.Error("expected finished true after processing the only tool")
	}
	if nextIdx != 0 {
		t.Errorf("expected nextIndex 0 when finished, got %d", nextIdx)
	}
	if len(batch.Progress().Completed) != 1 {
		t.Errorf("expected 1 completed, got %d", len(batch.Progress().Completed))
	}
}

func TestBatchModel_HandleProgress_Failed(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools:        []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
		CurrentIndex: 0,
	}

	finished, _ := batch.HandleProgress(batchInstallProgressMsg{
		toolID: "t1",
		err:    &testError{msg: "install failed"},
	})

	if !finished {
		t.Error("expected finished true after processing the only tool (even on failure)")
	}
	if len(batch.Progress().Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(batch.Progress().Failed))
	}
}

func TestBatchModel_HandleProgress_Skipped(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools:        []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
		CurrentIndex: 0,
	}

	finished, _ := batch.HandleProgress(batchInstallProgressMsg{
		toolID:  "t1",
		skipped: true,
	})

	if !finished {
		t.Error("expected finished true after skipping the only tool")
	}
	if len(batch.Progress().Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(batch.Progress().Skipped))
	}
}

func TestBatchModel_HandleProgress_MultiTool(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools: []db.SearchResult{
			{Tool: db.Tool{ID: "t1", Name: "Tool1"}},
			{Tool: db.Tool{ID: "t2", Name: "Tool2"}},
			{Tool: db.Tool{ID: "t3", Name: "Tool3"}},
		},
		CurrentIndex: 0,
	}

	// First tool completes
	finished, nextIdx := batch.HandleProgress(batchInstallProgressMsg{
		toolID: "t1",
		output: "ok",
	})
	if finished {
		t.Error("expected not finished after first tool of 3")
	}
	if nextIdx != 1 {
		t.Errorf("expected nextIndex 1, got %d", nextIdx)
	}

	// Second tool fails
	finished, nextIdx = batch.HandleProgress(batchInstallProgressMsg{
		toolID: "t2",
		err:    &testError{msg: "fail"},
	})
	if finished {
		t.Error("expected not finished after second tool of 3")
	}
	if nextIdx != 2 {
		t.Errorf("expected nextIndex 2, got %d", nextIdx)
	}

	// Third tool skipped
	finished, nextIdx = batch.HandleProgress(batchInstallProgressMsg{
		toolID:  "t3",
		skipped: true,
	})
	if !finished {
		t.Error("expected finished after last tool")
	}
	if nextIdx != 0 {
		t.Errorf("expected nextIndex 0 when done, got %d", nextIdx)
	}

	// Final state
	prog := batch.Progress()
	if len(prog.Completed) != 1 {
		t.Errorf("expected 1 completed, got %d", len(prog.Completed))
	}
	if len(prog.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(prog.Failed))
	}
	if len(prog.Skipped) != 1 {
		t.Errorf("expected 1 skipped, got %d", len(prog.Skipped))
	}
	if !prog.IsComplete {
		t.Error("expected IsComplete true")
	}
}

func TestBatchModel_HandleProgress_NilProgress(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	finished, nextIdx := batch.HandleProgress(batchInstallProgressMsg{
		toolID: "t1",
		output: "ok",
	})
	if !finished {
		t.Error("expected finished true when progress is nil (nothing to track)")
	}
	if nextIdx != 0 {
		t.Errorf("expected nextIndex 0, got %d", nextIdx)
	}
}

func TestBatchModel_HandleComplete(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools: []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
	}

	batch.HandleComplete()
	if !batch.Progress().IsComplete {
		t.Error("expected IsComplete true after HandleComplete")
	}
}

func TestBatchModel_HandleComplete_NilProgress(_ *testing.T) {
	batch := NewBatchInstallModel(nil)
	// Should not panic
	batch.HandleComplete()
}

// ============================================================================
// BatchInstallModel — StartInstall
// ============================================================================

func TestBatchModel_StartInstall_SetsProgress(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	marked := []db.SearchResult{
		{Tool: db.Tool{ID: "t1", Name: "Tool1"}},
		{Tool: db.Tool{ID: "t2", Name: "Tool2"}},
	}

	cmd := batch.StartInstall(marked)
	if cmd == nil {
		t.Error("expected non-nil Cmd from StartInstall")
	}
	if !batch.IsRunning() {
		t.Error("expected IsRunning true after StartInstall")
	}
	prog := batch.Progress()
	if prog == nil {
		t.Fatal("expected non-nil Progress after StartInstall")
	}
	if len(prog.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(prog.Tools))
	}
	if prog.CurrentIndex != 0 {
		t.Errorf("expected CurrentIndex 0, got %d", prog.CurrentIndex)
	}
}

func TestBatchModel_StartInstall_ReturnsStartMsg(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	marked := []db.SearchResult{
		{Tool: db.Tool{ID: "t1", Name: "Tool1"}},
	}

	cmd := batch.StartInstall(marked)
	msg := cmd()

	startMsg, ok := msg.(batchInstallStartMsg)
	if !ok {
		t.Fatalf("expected batchInstallStartMsg, got %T", msg)
	}
	if len(startMsg.tools) != 1 {
		t.Errorf("expected 1 tool in start msg, got %d", len(startMsg.tools))
	}
}

// ============================================================================
// BatchInstallModel — ProcessTool edge cases
// ============================================================================

func TestBatchModel_ProcessTool_NilProgress(t *testing.T) {
	batch := NewBatchInstallModel(nil)

	cmd := batch.ProcessTool(0)
	msg := cmd()

	_, ok := msg.(batchInstallCompleteMsg)
	if !ok {
		t.Errorf("expected batchInstallCompleteMsg when progress is nil, got %T", msg)
	}
}

func TestBatchModel_ProcessTool_IndexOutOfBounds(t *testing.T) {
	batch := NewBatchInstallModel(nil)
	batch.progress = &BatchInstallProgress{
		Tools:        []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
		CurrentIndex: 0,
	}

	cmd := batch.ProcessTool(5) // out of bounds
	msg := cmd()

	_, ok := msg.(batchInstallCompleteMsg)
	if !ok {
		t.Errorf("expected batchInstallCompleteMsg for out-of-bounds index, got %T", msg)
	}
}

func TestBatchModel_ProcessTool_HappyPath(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	tool := &db.Tool{ID: "t1", Slug: "tool1", Name: "Tool1"}
	if err := database.UpsertTool(ctx, tool); err != nil {
		t.Fatalf("failed to seed tool: %v", err)
	}
	inst := &db.InstallInstruction{ID: "inst-t1", ToolID: "t1", Platform: "linux", Command: "echo ok"}
	if err := database.UpsertInstallInstruction(ctx, inst); err != nil {
		t.Fatalf("failed to seed install instruction: %v", err)
	}

	batch := NewBatchInstallModel(database)
	batch.progress = &BatchInstallProgress{
		Tools:        []db.SearchResult{{Tool: db.Tool{ID: "t1", Name: "Tool1"}}},
		CurrentIndex: 0,
	}

	cmd := batch.ProcessTool(0)
	msg := cmd()

	progMsg, ok := msg.(batchInstallProgressMsg)
	if !ok {
		t.Fatalf("expected batchInstallProgressMsg, got %T", msg)
	}
	if progMsg.err != nil {
		t.Errorf("expected no error from echo command, got: %v", progMsg.err)
	}
	if !strings.Contains(progMsg.output, "ok") {
		t.Errorf("expected output to contain 'ok', got: %q", progMsg.output)
	}
}

func TestBatchModel_ProcessTool_SkipIfBlind_NoInstalls(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("failed to create in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	ctx := context.Background()
	// Seed a tool but NO install instructions, so GetInstallInstructions returns empty
	tool := &db.Tool{ID: "t1", Slug: "tool1", Name: "Tool1"}
	if err := database.UpsertTool(ctx, tool); err != nil {
		t.Fatalf("failed to seed tool: %v", err)
	}

	batch := NewBatchInstallModel(database)
	batch.StartBatchConfig(false)
	batch.Config().SkipIfBlind = true

	batch.progress = &BatchInstallProgress{
		Tools: []db.SearchResult{
			{Tool: db.Tool{ID: "t1", Name: "Tool1"}},
		},
		CurrentIndex: 0,
	}

	cmd := batch.ProcessTool(0)
	msg := cmd()

	progMsg, ok := msg.(batchInstallProgressMsg)
	if !ok {
		t.Fatalf("expected batchInstallProgressMsg, got %T", msg)
	}
	if !progMsg.skipped {
		t.Error("expected skipped=true when SkipIfBlind is set and no install instructions exist")
	}
}

// ============================================================================
// tea.Msg type tests
// ============================================================================

func TestBatchInstallStartMsg_Fields(t *testing.T) {
	msg := batchInstallStartMsg{
		tools: []db.SearchResult{
			{Tool: db.Tool{ID: "a", Name: "Alpha"}},
		},
	}
	if len(msg.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(msg.tools))
	}
	if msg.tools[0].Name != "Alpha" {
		t.Errorf("expected Alpha, got %s", msg.tools[0].Name)
	}
}

func TestBatchInstallProgressMsg_Fields(t *testing.T) {
	msg := batchInstallProgressMsg{
		toolID:  "t1",
		output:  "hello",
		err:     &testError{msg: "boom"},
		skipped: false,
	}
	if msg.toolID != "t1" {
		t.Errorf("expected t1, got %s", msg.toolID)
	}
	if msg.output != "hello" {
		t.Errorf("expected hello, got %s", msg.output)
	}
	if msg.err == nil {
		t.Error("expected non-nil error")
	}
	if msg.skipped {
		t.Error("expected skipped=false")
	}
}

func TestBatchInstallCompleteMsg(_ *testing.T) {
	msg := batchInstallCompleteMsg{}
	// Verify it satisfies tea.Msg interface
	var _ tea.Msg = msg
}

// testError is a simple error type for tests
type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }
