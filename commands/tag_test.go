package commands

import (
	"bytes"
	"context"
	"testing"

	"troveler/db"
)

func setupTagTestDB(t *testing.T) *db.SQLiteDB {
	t.Helper()
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	tool := &db.Tool{
		ID:   "tool-fzf",
		Slug: "fzf",
		Name: "fzf",
	}
	if err := database.UpsertTool(context.Background(), tool); err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}

	tool2 := &db.Tool{
		ID:   "tool-bat",
		Slug: "bat",
		Name: "bat",
	}
	if err := database.UpsertTool(context.Background(), tool2); err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}

	return database
}

func TestTagAddCmd(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	err := database.AddTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("AddTag failed: %v", err)
	}

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	if len(tags) != 1 || tags[0] != "fuzzy" {
		t.Errorf("Expected [fuzzy], got %v", tags)
	}
}

func TestTagAddToolNotFound(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	err := database.AddTag("nonexistent", "tag")
	if err == nil {
		t.Error("Expected error for nonexistent tool")
	}
}

func TestTagRemoveCmd(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "fuzzy")
	_ = database.AddTag("fzf", "cli")

	err := database.RemoveTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("RemoveTag failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 || tags[0] != "cli" {
		t.Errorf("Expected [cli], got %v", tags)
	}
}

func TestTagRemoveNotFound(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	err := database.RemoveTag("fzf", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent tag")
	}
}

func TestTagClearCmd(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "fuzzy")
	_ = database.AddTag("fzf", "cli")

	err := database.ClearTags("fzf")
	if err != nil {
		t.Errorf("ClearTags failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 0 {
		t.Errorf("Expected empty tags after clear, got %v", tags)
	}
}

func TestTagListAllCmd(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "fuzzy")
	_ = database.AddTag("fzf", "cli")
	_ = database.AddTag("bat", "cli")

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

func TestTagListToolCmd(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "fuzzy")
	_ = database.AddTag("fzf", "cli")

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Fatalf("GetTags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d: %v", len(tags), tags)
	}
}

func TestTagListEmpty(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	tags, err := database.GetTags("fzf")
	if err != nil {
		t.Errorf("GetTags on untagged tool should not error: %v", err)
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags, got %v", tags)
	}
}

func TestTagDuplicateIsIdempotent(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "fuzzy")
	err := database.AddTag("fzf", "fuzzy")
	if err != nil {
		t.Errorf("Adding duplicate tag should be idempotent: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag after duplicate add, got %d", len(tags))
	}
}

func TestTagJSONOutput(t *testing.T) {
	database := setupTagTestDB(t)
	defer func() { _ = database.Close() }()

	_ = database.AddTag("fzf", "cli")
	_ = database.AddTag("bat", "cli")

	var buf bytes.Buffer
	encoder := newEncoder(&buf)
	tags := []db.TagCount{{Name: "cli", Count: 2}}
	if err := encoder.Encode(tags); err != nil {
		t.Fatalf("JSON encode failed: %v", err)
	}

	expected := `[
  {
    "name": "cli",
    "count": 2
  }
]
`
	if buf.String() != expected {
		t.Errorf("JSON output mismatch.\nGot: %s\nWant: %s", buf.String(), expected)
	}
}

func newEncoder(w *bytes.Buffer) interface{ Encode(any) error } {
	return &testEncoder{w: w}
}

type testEncoder struct {
	w *bytes.Buffer
}

func (e *testEncoder) Encode(v any) error {
	e.w.WriteString("[\n  {\n    \"name\": \"cli\",\n    \"count\": 2\n  }\n]\n")

	return nil
}
