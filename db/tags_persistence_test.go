package db

import (
	"context"
	"testing"
)

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
