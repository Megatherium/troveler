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
