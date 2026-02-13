package db

import (
	"context"
	"testing"
)

func setupTestDB(t *testing.T) *SQLiteDB {
	t.Helper()
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	return database
}

func seedTool(t *testing.T, database *SQLiteDB, slug, name string) *Tool {
	t.Helper()
	tool := &Tool{
		ID:   "tool-" + slug,
		Slug: slug,
		Name: name,
	}
	err := database.UpsertTool(context.Background(), tool)
	if err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}
	return tool
}

func TestAddTag(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")

	err := database.AddTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	tags, err := database.GetTags("tool-fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	if len(tags) != 1 || tags[0] != "fuzzy" {
		t.Errorf("Expected [fuzzy], got %v", tags)
	}
}

func TestAddTagDuplicate(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")

	database.AddTag("fzf", "fuzzy")
	err := database.AddTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("AddTag duplicate should be idempotent, got: %v", err)
	}

	tags, _ := database.GetTags("tool-fzf")
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag after duplicate add, got %d", len(tags))
	}
}

func TestAddTagToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	err := database.AddTag("nonexistent", "fuzzy")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestAddMultipleTags(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")

	database.AddTag("fzf", "fuzzy")
	database.AddTag("fzf", "cli")

	tags, _ := database.GetTags("tool-fzf")
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
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")

	tags, err := database.GetTags("tool-fzf")
	if err != nil {
		t.Errorf("GetTags on untagged tool should not error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags, got %v", tags)
	}
}

func TestRemoveTag(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	database.AddTag("fzf", "fuzzy")
	database.AddTag("fzf", "cli")

	err := database.RemoveTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("RemoveTag failed: %v", err)
	}

	tags, _ := database.GetTags("tool-fzf")
	if len(tags) != 1 || tags[0] != "cli" {
		t.Errorf("Expected [cli], got %v", tags)
	}
}

func TestRemoveTagNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	database.AddTag("fzf", "fuzzy")

	err := database.RemoveTag("fzf", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tag")
	}
}

func TestRemoveTagToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	err := database.RemoveTag("nonexistent", "fuzzy")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestClearTags(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	database.AddTag("fzf", "fuzzy")
	database.AddTag("fzf", "cli")

	err := database.ClearTags("fzf")
	if err != nil {
		t.Errorf("ClearTags failed: %v", err)
	}

	tags, _ := database.GetTags("tool-fzf")
	if len(tags) != 0 {
		t.Errorf("Expected empty tags after clear, got %v", tags)
	}
}

func TestClearTagsToolNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	err := database.ClearTags("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestGetAllTags(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")

	database.AddTag("fzf", "fuzzy")
	database.AddTag("fzf", "cli")
	database.AddTag("bat", "cli")

	tags, err := database.GetAllTags()
	if err != nil {
		t.Fatalf("GetAllTags failed: %v", err)
	}

	expected := map[string]int{"cli": 2, "fuzzy": 1}
	for _, tc := range tags {
		if expected[tc.Name] != tc.Count {
			t.Errorf("Tag %s: expected count %d, got %d", tc.Name, expected[tc.Name], tc.Count)
		}
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}

func TestGetAllTagsEmpty(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	tags, err := database.GetAllTags()
	if err != nil {
		t.Errorf("GetAllTags on empty db should not error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags, got %v", tags)
	}
}

func TestGetToolsByTag(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	database.AddTag("fzf", "cli")
	database.AddTag("bat", "cli")
	database.AddTag("ripgrep", "search")

	tools, err := database.GetToolsByTag("cli")
	if err != nil {
		t.Fatalf("GetToolsByTag failed: %v", err)
	}

	if len(tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(tools))
	}

	names := make(map[string]bool)
	for _, tool := range tools {
		names[tool.Name] = true
	}
	if !names["fzf"] || !names["bat"] {
		t.Errorf("Expected fzf and bat, got %v", tools)
	}
}

func TestGetToolsByTagNotFound(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	tools, err := database.GetToolsByTag("nonexistent")
	if err != nil {
		t.Errorf("GetToolsByTag should not error for nonexistent tag: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("Expected empty tools, got %v", tools)
	}
}

func TestForeignKeyCascade(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	// Deleting the tool should cascade-delete tool_tags rows.
	_, err := database.db.ExecContext(context.Background(),
		"DELETE FROM tools WHERE slug = ?", "fzf")
	if err != nil {
		t.Fatalf("DELETE tool failed: %v", err)
	}

	// tool_tags should be empty — cascade fired.
	var count int
	err = database.db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM tool_tags").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT tool_tags failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tool_tags after cascade delete, got %d", count)
	}
}

func TestTagModel(t *testing.T) {
	tag := Tag{
		ID:   "tag-fuzzy",
		Name: "fuzzy",
	}
	if tag.ID != "tag-fuzzy" {
		t.Errorf("Tag.ID = %v, want %v", tag.ID, "tag-fuzzy")
	}
	if tag.Name != "fuzzy" {
		t.Errorf("Tag.Name = %v, want %v", tag.Name, "fuzzy")
	}
}

func TestTagCountModel(t *testing.T) {
	tc := TagCount{
		Name:  "cli",
		Count: 5,
	}
	if tc.Name != "cli" {
		t.Errorf("TagCount.Name = %v, want %v", tc.Name, "cli")
	}
	if tc.Count != 5 {
		t.Errorf("TagCount.Count = %v, want %v", tc.Count, 5)
	}
}
