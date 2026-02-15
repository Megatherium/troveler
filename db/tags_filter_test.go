package db

import (
	"context"
	"testing"
)

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
