package commands

import (
	"context"
	"testing"

	"troveler/crawler"
	"troveler/db"
)

func TestHandleDatabaseWritesRefreshesExistingTool(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	ctx := context.Background()
	for i, name := range []string{"Original", "Refreshed"} {
		details := make(chan crawler.DetailPage, 1)
		details <- crawler.DetailPage{
			Tool:          db.Tool{ID: name, Slug: "example", Name: name},
			Installations: map[string]string{"brew": "brew install " + name},
		}
		close(details)
		if err := handleDatabaseWrites(ctx, database, details); err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			if err := database.AddTag("example", "favorite"); err != nil {
				t.Fatal(err)
			}
		}
	}
	tools, err := database.GetAllTools(ctx)
	if err != nil || len(tools) != 1 || tools[0].ID != "Original" || tools[0].Name != "Refreshed" {
		t.Fatalf("incorrect tools after refresh: %v, error = %v", tools, err)
	}
	installs, err := database.GetInstallInstructions("Original")
	if err != nil || len(installs) != 1 || installs[0].Command != "brew install Refreshed" {
		t.Fatalf("incorrect installs after refresh: %v, error = %v", installs, err)
	}
}

func TestHandleDatabaseWritesReturnsWriteFailure(t *testing.T) {
	database, err := db.New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	details := make(chan crawler.DetailPage, 1)
	details <- crawler.DetailPage{Tool: db.Tool{ID: "example", Slug: "example"}}
	close(details)
	if err := handleDatabaseWrites(context.Background(), database, details); err == nil {
		t.Fatal("expected a database write failure")
	}
}
