package db

import (
	"testing"
)

func TestGetAllTagsBySlug(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tagsBySlug, err := database.GetAllTagsBySlug()
	if err != nil {
		t.Fatalf("GetAllTagsBySlug failed: %v", err)
	}

	if len(tagsBySlug) != 2 {
		t.Errorf("Expected 2 slugs with tags, got %d", len(tagsBySlug))
	}

	fzfTags := tagsBySlug["fzf"]
	if len(fzfTags) != 2 {
		t.Errorf("Expected fzf to have 2 tags, got %d: %v", len(fzfTags), fzfTags)
	}

	batTags := tagsBySlug["bat"]
	if len(batTags) != 1 {
		t.Errorf("Expected bat to have 1 tag, got %d: %v", len(batTags), batTags)
	}
}

func TestGetAllTagsBySlugEmpty(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	tagsBySlug, err := database.GetAllTagsBySlug()
	if err != nil {
		t.Fatalf("GetAllTagsBySlug failed: %v", err)
	}

	if len(tagsBySlug) != 0 {
		t.Errorf("Expected 0 slugs with tags, got %d", len(tagsBySlug))
	}
}

func TestReapplyTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	snapshot, err := database.GetAllTagsBySlug()
	if err != nil {
		t.Fatalf("GetAllTagsBySlug failed: %v", err)
	}

	if err := database.ClearTags("fzf"); err != nil {
		t.Fatalf("ClearTags failed: %v", err)
	}

	if err := database.ReapplyTags(snapshot); err != nil {
		t.Fatalf("ReapplyTags failed: %v", err)
	}

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	if len(tags) != 1 || tags[0] != "fuzzy" {
		t.Errorf("Expected tags to be restored, got %v", tags)
	}
}

func TestReapplyTagsNewTool(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	snapshot := map[string][]string{
		"fzf": {"cli", "fuzzy"},
	}

	if err := database.ReapplyTags(snapshot); err != nil {
		t.Fatalf("ReapplyTags failed: %v", err)
	}

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d: %v", len(tags), tags)
	}
}

func TestReapplyTagsMissingTool(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	snapshot := map[string][]string{
		"fzf":         {"cli"},
		"nonexistent": {"orphan"},
	}

	if err := database.ReapplyTags(snapshot); err != nil {
		t.Fatalf("ReapplyTags failed: %v", err)
	}

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d: %v", len(tags), tags)
	}
}
