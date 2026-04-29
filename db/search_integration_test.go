package db

import (
	"context"
	"testing"
)

func TestSearchWithInstalledFilterAndLimit(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstall(t, database, "tool-go", "go", "go")
	seedToolWithInstall(t, database, "tool-bat", "bat", "bat")
	seedToolWithInstall(t, database, "tool-fzf", "fzf", "fzf")
	seedToolWithInstall(t, database, "tool-notinstalled", "notinstalled", "this-command-does-not-exist-12345")

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     10,
		SortField: "name",
		SortOrder: "ASC",
		Filter: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Errorf("expected 3 installed tools (go, bat, fzf are installed), got %d", len(results))
	}

	for _, r := range results {
		if !r.Installed {
			t.Errorf("expected tool %s to be installed", r.Name)
		}
	}
}

func TestSearchWithInstalledFilterAndSmallLimit(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstall(t, database, "tool-go", "go", "go")
	seedToolWithInstall(t, database, "tool-bat", "bat", "bat")
	seedToolWithInstall(t, database, "tool-fzf", "fzf", "fzf")
	seedToolWithInstall(t, database, "tool-notinstalled", "notinstalled", "this-command-does-not-exist-12345")

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     2,
		SortField: "name",
		SortOrder: "ASC",
		Filter: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("expected 2 results (limit applied after filter), got %d", len(results))
	}

	if results[0].Name != "bat" {
		t.Errorf("expected 'bat' first (sorted by name ASC), got '%s'", results[0].Name)
	}
	if results[1].Name != "fzf" {
		t.Errorf("expected 'fzf' second (sorted by name ASC), got '%s'", results[1].Name)
	}
}

func TestSearchWithInstalledFilterAndLanguageFilter(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedToolWithInstallAndLanguage(t, database, "tool-go-cli", "go-cli", "go", "go")
	seedToolWithInstallAndLanguage(t, database, "tool-bat-clj", "bat-clj", "bat", "clojure")
	seedToolWithInstallAndLanguage(t, database, "tool-fzf-rs", "fzf-rs", "fzf", "rust")
	seedToolWithInstallAndLanguage(t, database, "tool-notinstalled", "notinstalled",
		"this-command-does-not-exist-12345", "go")

	filter := &Filter{
		Type: FilterAnd,
		Left: &Filter{
			Type:  FilterField,
			Field: "installed",
			Value: "true",
		},
		Right: &Filter{
			Type:  FilterField,
			Field: "language",
			Value: "go",
		},
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     10,
		SortField: "name",
		SortOrder: "ASC",
		Filter:    filter,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result (go language + installed), got %d", len(results))
	}

	if len(results) > 0 && results[0].Language != "go" {
		t.Errorf("expected language 'go', got '%s'", results[0].Language)
	}
}

func TestSearchSortOrderCaseInsensitive(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	// Seed tools with mixed-case names to verify COLLATE NOCASE
	for _, name := range []string{"Zebra", "apple", "Mango", "banana", "Cherry"} {
		tool := &Tool{ID: "tool-" + name, Slug: name, Name: name}
		if err := database.UpsertTool(context.Background(), tool); err != nil {
			t.Fatalf("Failed to seed tool %s: %v", name, err)
		}
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     10,
		SortField: "name",
		SortOrder: "ASC",
		Filter:    nil,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 5 {
		t.Fatalf("expected 5 results, got %d", len(results))
	}

	// Case-insensitive ASC: apple, banana, Cherry, Mango, Zebra
	expected := []string{"apple", "banana", "Cherry", "Mango", "Zebra"}
	for i, exp := range expected {
		if results[i].Name != exp {
			t.Errorf("position %d: expected '%s', got '%s'", i, exp, results[i].Name)
		}
	}
}

func TestSearchSortOrderDesc(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	for _, name := range []string{"alpha", "beta", "gamma"} {
		tool := &Tool{ID: "tool-" + name, Slug: name, Name: name}
		if err := database.UpsertTool(context.Background(), tool); err != nil {
			t.Fatalf("Failed to seed tool %s: %v", name, err)
		}
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     10,
		SortField: "name",
		SortOrder: "DESC",
		Filter:    nil,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	if results[0].Name != "gamma" {
		t.Errorf("expected 'gamma' first (DESC), got '%s'", results[0].Name)
	}
	if results[1].Name != "beta" {
		t.Errorf("expected 'beta' second (DESC), got '%s'", results[1].Name)
	}
	if results[2].Name != "alpha" {
		t.Errorf("expected 'alpha' third (DESC), got '%s'", results[2].Name)
	}
}

func TestSearchSortByLanguage(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	for _, tc := range []struct {
		name     string
		language string
	}{
		{"tool-rust", "rust"},
		{"tool-go", "go"},
		{"tool-python", "python"},
	} {
		tool := &Tool{ID: tc.name, Slug: tc.name, Name: tc.name, Language: tc.language}
		if err := database.UpsertTool(context.Background(), tool); err != nil {
			t.Fatalf("Failed to seed tool %s: %v", tc.name, err)
		}
	}

	results, err := database.Search(context.Background(), SearchOptions{
		Query:     "",
		Limit:     10,
		SortField: "language",
		SortOrder: "ASC",
		Filter:    nil,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	// Case-insensitive ASC: go, python, rust
	if results[0].Language != "go" {
		t.Errorf("expected 'go' first, got '%s'", results[0].Language)
	}
	if results[1].Language != "python" {
		t.Errorf("expected 'python' second, got '%s'", results[1].Language)
	}
	if results[2].Language != "rust" {
		t.Errorf("expected 'rust' third, got '%s'", results[2].Language)
	}
}

func seedToolWithInstall(t *testing.T, database *SQLiteDB, id, name, command string) {
	t.Helper()
	tool := &Tool{
		ID:   id,
		Slug: name,
		Name: name,
	}
	err := database.UpsertTool(context.Background(), tool)
	if err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}

	inst := &InstallInstruction{
		ID:       "inst-" + id,
		ToolID:   id,
		Platform: "linux",
		Command:  command,
	}
	err = database.UpsertInstallInstruction(context.Background(), inst)
	if err != nil {
		t.Fatalf("Failed to seed install instruction: %v", err)
	}
}

func seedToolWithInstallAndLanguage(t *testing.T, database *SQLiteDB, id, name, command, language string) {
	t.Helper()
	tool := &Tool{
		ID:       id,
		Slug:     name,
		Name:     name,
		Language: language,
	}
	err := database.UpsertTool(context.Background(), tool)
	if err != nil {
		t.Fatalf("Failed to seed tool: %v", err)
	}

	inst := &InstallInstruction{
		ID:       "inst-" + id,
		ToolID:   id,
		Platform: "linux",
		Command:  command,
	}
	err = database.UpsertInstallInstruction(context.Background(), inst)
	if err != nil {
		t.Fatalf("Failed to seed install instruction: %v", err)
	}
}
