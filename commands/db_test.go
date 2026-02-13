package commands

import (
	"context"
	"testing"

	"troveler/config"
	"troveler/db"
)

func TestDatabaseInitialization(t *testing.T) {
	cfg := &config.Config{
		DSN: ":memory:",
	}

	database, err := db.New(cfg.DSN)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	count, err := database.ToolCount(context.Background())
	if err != nil {
		t.Fatalf("Failed to count tools: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 tools in new database, got %d", count)
	}
}

func TestDatabaseInitializationWithBadDSN(t *testing.T) {
	cfg := &config.Config{
		DSN: "file:///invalid/path/to/database.db",
	}

	_, err := db.New(cfg.DSN)
	if err == nil {
		t.Error("Expected error for invalid DSN path, got nil")
	}
}

func TestWithDBHelper(t *testing.T) {
	cfg := &config.Config{
		DSN: ":memory:",
	}

	called := false
	var dbPtr *db.SQLiteDB

	err := testWithDBHelper(cfg, func(ctx context.Context, database *db.SQLiteDB) error {
		called = true
		dbPtr = database

		return nil
	})

	if !called {
		t.Error("Handler function was not called")
	}

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if dbPtr == nil {
		t.Error("Database pointer was nil in handler")
	}
}

func testWithDBHelper(cfg *config.Config, fn func(context.Context, *db.SQLiteDB) error) error {
	database, err := db.New(cfg.DSN)
	if err != nil {
		return err
	}
	defer func() { _ = database.Close() }()

	return fn(context.Background(), database)
}
