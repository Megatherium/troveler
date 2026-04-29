package db

import (
	"context"
	"testing"
)

func TestGetInstallInstructionsBatch_EmptyInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	result, err := database.GetInstallInstructionsBatch(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}

	result, err = database.GetInstallInstructionsBatch(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map for empty slice, got %d entries", len(result))
	}
}

func TestGetInstallInstructionsBatch_SingleTool(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstall(t, database, "tool-a", "tool-a", "brew install bat")

	result, err := database.GetInstallInstructionsBatch(context.Background(), []string{"tool-a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 tool in result, got %d", len(result))
	}

	installs := result["tool-a"]
	if len(installs) != 1 {
		t.Fatalf("expected 1 install instruction, got %d", len(installs))
	}
	if installs[0].Command != "brew install bat" {
		t.Errorf("expected command 'brew install bat', got %q", installs[0].Command)
	}
}

func TestGetInstallInstructionsBatch_MultipleTools(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstall(t, database, "tool-a", "tool-a", "brew install bat")
	seedToolWithInstall(t, database, "tool-b", "tool-b", "cargo install ripgrep")
	seedToolWithInstall(t, database, "tool-c", "tool-c", "go install github.com/foo/bar@latest")

	result, err := database.GetInstallInstructionsBatch(context.Background(), []string{"tool-a", "tool-b", "tool-c"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 tools in result, got %d", len(result))
	}

	if len(result["tool-a"]) != 1 || result["tool-a"][0].Command != "brew install bat" {
		t.Error("tool-a mismatch")
	}
	if len(result["tool-b"]) != 1 || result["tool-b"][0].Command != "cargo install ripgrep" {
		t.Error("tool-b mismatch")
	}
	if len(result["tool-c"]) != 1 || result["tool-c"][0].Command != "go install github.com/foo/bar@latest" {
		t.Error("tool-c mismatch")
	}
}

func TestGetInstallInstructionsBatch_MissingToolIDs(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstall(t, database, "tool-a", "tool-a", "brew install bat")

	result, err := database.GetInstallInstructionsBatch(context.Background(), []string{"tool-a", "nonexistent-id"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only tool-a should have results; nonexistent-id should be absent from map
	if len(result) != 1 {
		t.Fatalf("expected 1 tool in result, got %d", len(result))
	}
	if _, ok := result["nonexistent-id"]; ok {
		t.Error("expected nonexistent-id to be absent from result map")
	}
	if len(result["tool-a"]) != 1 {
		t.Error("expected tool-a to have 1 install instruction")
	}
}

func TestGetInstallInstructionsBatch_MultipleInstallsPerTool(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	tool := &Tool{ID: "tool-multi", Slug: "multi", Name: "multi"}
	if err := database.UpsertTool(context.Background(), tool); err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}

	inst1 := &InstallInstruction{ID: "inst-1", ToolID: "tool-multi", Platform: "linux", Command: "brew install bat"}
	inst2 := &InstallInstruction{ID: "inst-2", ToolID: "tool-multi", Platform: "macos", Command: "port install bat"}
	if err := database.UpsertInstallInstruction(context.Background(), inst1); err != nil {
		t.Fatalf("Failed to seed install 1: %v", err)
	}
	if err := database.UpsertInstallInstruction(context.Background(), inst2); err != nil {
		t.Fatalf("Failed to seed install 2: %v", err)
	}

	result, err := database.GetInstallInstructionsBatch(context.Background(), []string{"tool-multi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result["tool-multi"]) != 2 {
		t.Errorf("expected 2 install instructions for tool-multi, got %d", len(result["tool-multi"]))
	}
}

func TestGetInstallInstructionsBatch_Chunking(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	// Seed 3 tools and request them with sqliteVarLimit=500 — they'll fit in one chunk.
	// This validates the chunking loop executes correctly for the single-chunk case.
	seedToolWithInstall(t, database, "tool-1", "tool-1", "brew install a")
	seedToolWithInstall(t, database, "tool-2", "tool-2", "brew install b")
	seedToolWithInstall(t, database, "tool-3", "tool-3", "brew install c")

	ids := []string{"tool-1", "tool-2", "tool-3"}
	result, err := database.GetInstallInstructionsBatch(context.Background(), ids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 tools, got %d", len(result))
	}
}
