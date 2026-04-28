// Package db provides SQLite-backed storage for tools and install instructions.
package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // sqlite3 driver
)

// SQLiteDB wraps a *sql.DB for SQLite operations.
type SQLiteDB struct {
	db *sql.DB
}

// New opens (or creates) a SQLite database at dbPath and runs migrations.
func New(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	if _, err := db.ExecContext(context.Background(), "PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	sqlite := &SQLiteDB{db: db}
	if err := sqlite.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	if err := sqlite.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return sqlite, nil
}

// Close releases the underlying database connection.
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

func (s *SQLiteDB) getDB() *sql.DB {
	return s.db
}
