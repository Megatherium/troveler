package db

import (
	"context"
	"testing"
)

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
