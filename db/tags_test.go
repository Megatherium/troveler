package db

import (
	"context"
	"testing"
)

const (
	testTagCLI    = "cli"
	testTagFuzzy  = "fuzzy"
	testTagSearch = "search"
)

func setupTestDB(t *testing.T) *SQLiteDB {
	t.Helper()
	database, err := New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	return database
}

func checkClose(t *testing.T, db *SQLiteDB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func seedTool(t *testing.T, database *SQLiteDB, slug, name string) {
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
}

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

func TestNormalizeTagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"cli", "cli", false},
		{"  cli  ", "cli", false},
		{"CLI", "cli", false},
		{"  CLI  ", "cli", false},
		{"", "", true},
		{"   ", "", true},
		{"\t\n", "", true},
		{"FuZzY", "fuzzy", false},
	}

	for _, tt := range tests {
		result, err := normalizeTagName(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("normalizeTagName(%q) expected error, got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("normalizeTagName(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("normalizeTagName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	}
}

func TestAddTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	err := database.AddTag("fzf", "")
	if err == nil {
		t.Error("AddTag with empty name should fail")
	}

	err = database.AddTag("fzf", "   ")
	if err == nil {
		t.Error("AddTag with whitespace-only name should fail")
	}
}

func TestAddTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	if err := database.AddTag("fzf", "  CLI  "); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 {
		t.Errorf("Expected 1 normalized tag, got %d: %v", len(tags), tags)
	}
	if len(tags) > 0 && tags[0] != "cli" {
		t.Errorf("Expected normalized tag 'cli', got %q", tags[0])
	}
}

func TestRemoveTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	err := database.RemoveTag("fzf", "  CLI  ")
	if err != nil {
		t.Errorf("RemoveTag with different case/whitespace should work: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after removal, got %d", len(tags))
	}
}

func TestRemoveTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	err := database.RemoveTag("fzf", "")
	if err == nil {
		t.Error("RemoveTag with empty name should fail")
	}
}

func TestGetToolsByTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tools, err := database.GetToolsByTag("  CLI  ")
	if err != nil {
		t.Errorf("GetToolsByTag with different case/whitespace should work: %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestGetToolsByTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	_, err := database.GetToolsByTag("")
	if err == nil {
		t.Error("GetToolsByTag with empty name should fail")
	}
}

func TestPruneOrphanedTagsAfterRemove(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "orphan"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	if err := database.RemoveTag("fzf", "orphan"); err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	var count int
	err := database.db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM tags").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT tags failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tags after pruning, got %d", count)
	}
}

func TestPruneOrphanedTagsAfterClear(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "orphan"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	if err := database.ClearTags("fzf"); err != nil {
		t.Fatalf("ClearTags failed: %v", err)
	}

	var count int
	err := database.db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM tags").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT tags failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tags after pruning, got %d", count)
	}
}

func TestPruneOrphanedTagsPreservesUsedTags(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag fzf/cli failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag bat/cli failed: %v", err)
	}

	if err := database.RemoveTag("fzf", "cli"); err != nil {
		t.Fatalf("RemoveTag failed: %v", err)
	}

	var count int
	err := database.db.QueryRowContext(context.Background(),
		"SELECT COUNT(*) FROM tags").Scan(&count)
	if err != nil {
		t.Fatalf("COUNT tags failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 tag still in use, got %d", count)
	}
}

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

func TestSearchWithTagFilter(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag fzf/cli failed: %v", err)
	}
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag fzf/fuzzy failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag bat/cli failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "search"); err != nil {
		t.Fatalf("AddTag ripgrep/search failed: %v", err)
	}

	filter := &Filter{
		Type:  FilterField,
		Field: "tag",
		Value: "cli",
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for tag=cli, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["fzf"] || !names["bat"] {
		t.Errorf("Expected fzf and bat, got %v", names)
	}
}

func TestSearchWithTagNotFilter(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "experimental"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	filter := &Filter{
		Type: FilterNot,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "experimental",
		},
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	for _, r := range results {
		if r.Name == "fzf" {
			t.Errorf("fzf should have been excluded by !tag=experimental")
		}
	}
}

func TestSearchWithMultipleTagFilters(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "search"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "cli",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "fuzzy",
		},
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for tag=cli&tag=fuzzy, got %d", len(results))
	}
	if len(results) > 0 && results[0].Name != "fzf" {
		t.Errorf("Expected fzf, got %s", results[0].Name)
	}
}

func TestSearchWithTagOrFilter(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	seedTool(t, database, "bat", "bat")
	seedTool(t, database, "ripgrep", "ripgrep")

	if err := database.AddTag("fzf", "fuzzy"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "search"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	filter := &Filter{
		Type: FilterOr,
		Left: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "fuzzy",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "search",
		},
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for tag=fuzzy|tag=search, got %d", len(results))
	}

	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["fzf"] || !names["ripgrep"] {
		t.Errorf("Expected fzf and ripgrep, got %v", names)
	}
}

func TestSearchWithTagAndLanguageFilter(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	tool1 := &Tool{ID: "tool-fzf", Slug: "fzf", Name: "fzf", Language: "Go"}
	tool2 := &Tool{ID: "tool-bat", Slug: "bat", Name: "bat", Language: "Rust"}
	tool3 := &Tool{ID: "tool-rg", Slug: "ripgrep", Name: "ripgrep", Language: "Rust"}

	if err := database.UpsertTool(context.Background(), tool1); err != nil {
		t.Fatalf("UpsertTool failed: %v", err)
	}
	if err := database.UpsertTool(context.Background(), tool2); err != nil {
		t.Fatalf("UpsertTool failed: %v", err)
	}
	if err := database.UpsertTool(context.Background(), tool3); err != nil {
		t.Fatalf("UpsertTool failed: %v", err)
	}

	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("bat", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("ripgrep", "search"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "Rust",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "tag",
			Value: "cli",
		},
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Filter: filter,
		Limit:  10,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for language=Rust&tag=cli, got %d", len(results))
	}
	if len(results) > 0 && results[0].Name != "bat" {
		t.Errorf("Expected bat, got %s", results[0].Name)
	}
}
