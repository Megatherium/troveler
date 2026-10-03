package db

import (
	"context"
	"reflect"
	"testing"
)

func TestUpsertToolPreservesIdentityBySlug(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)
	ctx := context.Background()
	original := &Tool{ID: "original-id", Slug: "example", Name: "Original"}
	if err := database.UpsertTool(ctx, original); err != nil {
		t.Fatal(err)
	}
	if err := database.AddTag("example", "favorite"); err != nil {
		t.Fatal(err)
	}
	refreshed := &Tool{ID: "new-crawl-id", Slug: "example", Name: "Refreshed", Description: "New description"}
	if err := database.UpsertTool(ctx, refreshed); err != nil {
		t.Fatal(err)
	}
	if refreshed.ID != original.ID {
		t.Fatalf("stored identity = %q, want %q", refreshed.ID, original.ID)
	}
	tools, err := database.GetAllTools(ctx)
	if err != nil || len(tools) != 1 {
		t.Fatalf("tools = %v, error = %v", tools, err)
	}
	if tools[0].Name != refreshed.Name || tools[0].Description != refreshed.Description {
		t.Fatalf("metadata was not refreshed: %+v", tools[0])
	}
	tags, err := database.GetTags("example")
	if err != nil || !reflect.DeepEqual(tags, []string{"favorite"}) {
		t.Fatalf("tags = %v, error = %v", tags, err)
	}
	// Updating by the stored ID still allows a slug change.
	refreshed.Slug = "renamed"
	if err := database.UpsertTool(ctx, refreshed); err != nil {
		t.Fatal(err)
	}
	tools, err = database.GetToolBySlug("renamed")
	if err != nil || len(tools) != 1 || tools[0].ID != original.ID {
		t.Fatalf("renamed tools = %v, error = %v", tools, err)
	}
}

func TestSaveToolSnapshotReplacesInstallsAndPreservesTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)
	ctx := context.Background()
	tool := &Tool{ID: "original-id", Slug: "example", Name: "Original"}
	oldInstalls := []InstallInstruction{
		{ID: "old-brew", Platform: "brew", Command: "brew install old"},
		{ID: "old-go", Platform: "go", Command: "go install old"},
	}
	if err := database.SaveToolSnapshot(ctx, tool, oldInstalls); err != nil {
		t.Fatal(err)
	}
	if err := database.AddTag("example", "favorite"); err != nil {
		t.Fatal(err)
	}
	refreshed := &Tool{ID: "new-id", Slug: "example", Name: "Refreshed"}
	newInstalls := []InstallInstruction{
		{ID: "new-brew", ToolID: "new-id", Platform: "brew", Command: "brew install new", ExecutableName: "new"},
	}
	for range 2 {
		if err := database.SaveToolSnapshot(ctx, refreshed, newInstalls); err != nil {
			t.Fatal(err)
		}
	}
	installs, err := database.GetInstallInstructions(tool.ID)
	if err != nil || len(installs) != 1 {
		t.Fatalf("installs = %v, error = %v", installs, err)
	}
	if refreshed.ID != tool.ID || installs[0].ToolID != tool.ID || installs[0].Command != "brew install new" ||
		installs[0].ExecutableName != "new" {
		t.Fatalf("incorrect refreshed tool/installs: %+v / %+v", refreshed, installs)
	}
	if newInstalls[0].ToolID != "new-id" {
		t.Fatal("input instructions were mutated")
	}
	tags, err := database.GetTags("example")
	if err != nil || !reflect.DeepEqual(tags, []string{"favorite"}) {
		t.Fatalf("tags = %v, error = %v", tags, err)
	}
	if err := database.SaveToolSnapshot(ctx, refreshed, nil); err != nil {
		t.Fatal(err)
	}
	installs, err = database.GetInstallInstructions(tool.ID)
	if err != nil || len(installs) != 0 {
		t.Fatalf("obsolete installs retained: %v, error = %v", installs, err)
	}
}

func TestSaveToolSnapshotRollsBackPartialRefresh(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)
	ctx := context.Background()
	original := &Tool{ID: "original-id", Slug: "example", Name: "Original"}
	if err := database.SaveToolSnapshot(ctx, original, []InstallInstruction{
		{ID: "old-install", Platform: "brew", Command: "brew install old"},
	}); err != nil {
		t.Fatal(err)
	}
	other := &Tool{ID: "other-id", Slug: "other", Name: "Other"}
	if err := database.SaveToolSnapshot(ctx, other, []InstallInstruction{
		{ID: "collision", Platform: "brew", Command: "brew install other"},
	}); err != nil {
		t.Fatal(err)
	}
	refreshed := &Tool{ID: "new-id", Slug: "example", Name: "Refreshed"}
	before := *refreshed
	err := database.SaveToolSnapshot(ctx, refreshed, []InstallInstruction{
		{ID: "new-install", Platform: "go", Command: "go install new"},
		{ID: "collision", Platform: "brew", Command: "brew install new"},
	})
	if err == nil {
		t.Fatal("expected instruction ID collision to fail")
	}
	if *refreshed != before {
		t.Fatalf("failed refresh changed input tool: %+v", refreshed)
	}
	tools, err := database.GetToolBySlug("example")
	if err != nil || len(tools) != 1 || tools[0].Name != "Original" || tools[0].ID != original.ID {
		t.Fatalf("metadata was not rolled back: %v, error = %v", tools, err)
	}
	installs, err := database.GetInstallInstructions(original.ID)
	if err != nil || len(installs) != 1 || installs[0].ID != "old-install" {
		t.Fatalf("instructions were not rolled back: %v, error = %v", installs, err)
	}
	installs, err = database.GetInstallInstructions(other.ID)
	if err != nil || len(installs) != 1 || installs[0].Command != "brew install other" {
		t.Fatalf("other tool was changed: %v, error = %v", installs, err)
	}
}
