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
