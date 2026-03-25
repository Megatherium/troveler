package db

import (
	"testing"
)

func TestSortAndLimitResultsNameASC(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "zebra"}},
		{Tool: Tool{Name: "apple"}},
		{Tool: Tool{Name: "mango"}},
	}

	sorted := sortAndLimitResults(results, "name", "ASC", 0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Name != "apple" {
		t.Errorf("expected 'apple' first, got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "mango" {
		t.Errorf("expected 'mango' second, got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "zebra" {
		t.Errorf("expected 'zebra' third, got '%s'", sorted[2].Name)
	}
}

func TestSortAndLimitResultsNameDESC(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "zebra"}},
		{Tool: Tool{Name: "apple"}},
		{Tool: Tool{Name: "mango"}},
	}

	sorted := sortAndLimitResults(results, "name", "DESC", 0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Name != "zebra" {
		t.Errorf("expected 'zebra' first, got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "mango" {
		t.Errorf("expected 'mango' second, got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "apple" {
		t.Errorf("expected 'apple' third, got '%s'", sorted[2].Name)
	}
}

func TestSortAndLimitResultsWithLimit(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "zebra"}},
		{Tool: Tool{Name: "apple"}},
		{Tool: Tool{Name: "mango"}},
		{Tool: Tool{Name: "banana"}},
		{Tool: Tool{Name: "cherry"}},
	}

	sorted := sortAndLimitResults(results, "name", "ASC", 3)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Name != "apple" {
		t.Errorf("expected 'apple' first, got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "banana" {
		t.Errorf("expected 'banana' second, got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "cherry" {
		t.Errorf("expected 'cherry' third, got '%s'", sorted[2].Name)
	}
}

func TestSortAndLimitResultsLimitExceedsLength(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "zebra"}},
		{Tool: Tool{Name: "apple"}},
	}

	sorted := sortAndLimitResults(results, "name", "ASC", 10)

	if len(sorted) != 2 {
		t.Fatalf("expected 2 results, got %d", len(sorted))
	}
}

func TestSortAndLimitResultsLanguageASC(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "a", Language: "rust"}},
		{Tool: Tool{Name: "b", Language: "go"}},
		{Tool: Tool{Name: "c", Language: "python"}},
	}

	sorted := sortAndLimitResults(results, "language", "ASC", 0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Language != "go" {
		t.Errorf("expected 'go' first, got '%s'", sorted[0].Language)
	}
	if sorted[1].Language != "python" {
		t.Errorf("expected 'python' second, got '%s'", sorted[1].Language)
	}
	if sorted[2].Language != "rust" {
		t.Errorf("expected 'rust' third, got '%s'", sorted[2].Language)
	}
}

func TestSortAndLimitResultsEmpty(t *testing.T) {
	results := []SearchResult{}

	sorted := sortAndLimitResults(results, "name", "ASC", 10)

	if len(sorted) != 0 {
		t.Fatalf("expected 0 results, got %d", len(sorted))
	}
}

func TestSortAndLimitResultsCaseInsensitive(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "Zebra"}},
		{Tool: Tool{Name: "apple"}},
		{Tool: Tool{Name: "Mango"}},
	}

	sorted := sortAndLimitResults(results, "name", "ASC", 0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Name != "apple" {
		t.Errorf("expected 'apple' first (case-insensitive), got '%s'", sorted[0].Name)
	}
	if sorted[1].Name != "Mango" {
		t.Errorf("expected 'Mango' second (case-insensitive), got '%s'", sorted[1].Name)
	}
	if sorted[2].Name != "Zebra" {
		t.Errorf("expected 'Zebra' third (case-insensitive), got '%s'", sorted[2].Name)
	}
}

func TestSortAndLimitResultsUnknownSortField(t *testing.T) {
	results := []SearchResult{
		{Tool: Tool{Name: "zebra"}},
		{Tool: Tool{Name: "apple"}},
		{Tool: Tool{Name: "mango"}},
	}

	sorted := sortAndLimitResults(results, "unknown_field", "ASC", 0)

	if len(sorted) != 3 {
		t.Fatalf("expected 3 results, got %d", len(sorted))
	}
	if sorted[0].Name != "apple" {
		t.Errorf("expected 'apple' first (default to name), got '%s'", sorted[0].Name)
	}
}
