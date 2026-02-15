package db

import (
	"testing"
)

func TestGetAllTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

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
	defer checkClose(t, database)

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
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "search"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

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
	defer checkClose(t, database)

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
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	_, err := database.db.ExecContext(nil,
		"DELETE FROM tools WHERE slug = ?", "fzf")
	if err != nil {
		t.Fatalf("DELETE tool failed: %v", err)
	}

	var count int
	err = database.db.QueryRowContext(nil,
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
		Name: "fuzzy",
	}
	if tag.Name != "fuzzy" {
		t.Errorf("Tag.Name = %v, want %v", tag.Name, "fuzzy")
	}
}

func TestTagCountModel(t *testing.T) {
	tc := TagCount{
		Name:  testTagCLI,
		Count: 5,
	}
	if tc.Name != testTagCLI {
		t.Errorf("TagCount.Name = %v, want %v", tc.Name, "cli")
	}
	if tc.Count != 5 {
		t.Errorf("TagCount.Count = %v, want %v", tc.Count, 5)
	}
}
