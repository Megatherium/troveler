package db

import (
	"testing"
)

func TestAddTag(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	err := database.AddTag("fzf", testTagFuzzy)
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	if len(tags) != 1 || tags[0] != testTagFuzzy {
		t.Errorf("Expected [fuzzy], got %v", tags)
	}
}

func TestAddTagDuplicate(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	err := database.AddTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("AddTag duplicate should be idempotent, got: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag after duplicate add, got %d", len(tags))
	}
}

func TestAddTagToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	err := database.AddTag("nonexistent", "fuzzy")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestAddMultipleTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d: %v", len(tags), tags)
	}
	expected := map[string]bool{"fuzzy": true, "cli": true}
	for _, tag := range tags {
		if !expected[tag] {
			t.Errorf("Unexpected tag: %s", tag)
		}
	}
}

func TestGetTagsEmpty(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Errorf("GetTags on untagged tool should not error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags, got %v", tags)
	}
}

func TestRemoveTag(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	err := database.RemoveTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("RemoveTag failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 || tags[0] != testTagCLI {
		t.Errorf("Expected [cli], got %v", tags)
	}
}

func TestRemoveTagNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	err := database.RemoveTag("fzf", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tag")
	}
}

func TestRemoveTagToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	err := database.RemoveTag("nonexistent", "fuzzy")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestClearTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	err := database.ClearTags("fzf")
	if err != nil {
		t.Errorf("ClearTags failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 0 {
		t.Errorf("Expected empty tags after clear, got %v", tags)
	}
}

func TestClearTagsToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	err := database.ClearTags("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}
