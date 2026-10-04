package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mattn/go-sqlite3"
)

// Use real SQLite behind a test-local connector so initialization runs its actual
// schema and migrations, with deterministic failures and observable ownership.
type initializationConnector struct {
	failure  string
	cause    error
	closeErr error
	closes   atomic.Int32
}

func (c *initializationConnector) Connect(context.Context) (driver.Conn, error) {
	if c.failure == "connect" {
		return nil, c.cause
	}

	conn, err := c.Driver().Open(":memory:")
	if err != nil {
		return nil, err
	}

	sqlite, ok := conn.(*sqlite3.SQLiteConn)
	if !ok {
		return nil, errors.Join(fmt.Errorf("unexpected SQLite connection type %T", conn), conn.Close())
	}

	return &initializationConn{SQLiteConn: sqlite, fixture: c}, nil
}

func (*initializationConnector) Driver() driver.Driver {
	return &sqlite3.SQLiteDriver{}
}

type initializationConn struct {
	*sqlite3.SQLiteConn
	fixture *initializationConnector
}

func (c *initializationConn) Ping(ctx context.Context) error {
	if c.fixture.failure == "ping" {
		return c.fixture.cause
	}

	return c.SQLiteConn.Ping(ctx)
}

func (c *initializationConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.fixture.failure != "" && strings.Contains(query, c.fixture.failure) {
		return nil, c.fixture.cause
	}

	return c.SQLiteConn.ExecContext(ctx, query, args)
}

func (c *initializationConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if strings.Contains(query, "pragma_table_info") && len(args) > 0 &&
		c.fixture.failure == "query:"+fmt.Sprint(args[0].Value) {
		return nil, c.fixture.cause
	}

	return c.SQLiteConn.QueryContext(ctx, query, args)
}

func (c *initializationConn) Close() error {
	c.fixture.closes.Add(1)

	return errors.Join(c.SQLiteConn.Close(), c.fixture.closeErr)
}

func TestSQLiteInitializationClosesFailures(t *testing.T) {
	cases := []struct {
		name    string
		failure string
		prefix  string
	}{
		{"connection_open", "connect", "failed to ping db"},
		{"ping", "ping", "failed to ping db"},
		{"foreign_keys", "PRAGMA foreign_keys", "failed to enable foreign keys"},
		{"first_table", "CREATE TABLE IF NOT EXISTS tools", "failed to create tables"},
		{"last_index", "CREATE INDEX IF NOT EXISTS idx_tool_tags_tag_name", "failed to create tables"},
		{"first_migration_query", "query:tools", "failed to run migrations"},
		{"second_migration_query", "query:install_instructions", "failed to run migrations"},
		{"migration_alter", "ALTER TABLE", "failed to run migrations"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cause := errors.New("injected initialization failure")
			fixture := &initializationConnector{failure: tc.failure, cause: cause}
			handle := sql.OpenDB(fixture)
			t.Cleanup(func() {
				if err := handle.Close(); err != nil {
					t.Errorf("cleanup database: %v", err)
				}
			})

			database, err := initializeSQLite(handle)
			if database != nil || !errors.Is(err, cause) || !strings.HasPrefix(err.Error(), tc.prefix) {
				t.Fatalf("initialization = %v, %v; want nil database and wrapped %q", database, err, tc.prefix)
			}
			wantCloses := int32(1)
			if tc.failure == "connect" {
				wantCloses = 0
			}
			if got := fixture.closes.Load(); got != wantCloses {
				t.Errorf("connection closes = %d, want %d", got, wantCloses)
			}
			if got := handle.Stats().OpenConnections; got != 0 {
				t.Errorf("open connections after failure = %d, want 0", got)
			}
			// Clear the fault: a leaked handle could reconnect or reuse its existing
			// connection. A closed handle must remain unusable without another close.
			fixture.failure = ""
			if err := handle.PingContext(context.Background()); err == nil {
				t.Error("failed initialization left the database handle usable")
			}
		})
	}
}

func TestSQLiteInitializationPreservesCloseError(t *testing.T) {
	cause := errors.New("injected ping failure")
	closeErr := errors.New("injected close failure")
	fixture := &initializationConnector{failure: "ping", cause: cause, closeErr: closeErr}
	handle := sql.OpenDB(fixture)
	t.Cleanup(func() { _ = handle.Close() })

	database, err := initializeSQLite(handle)
	if database != nil || !errors.Is(err, cause) || !errors.Is(err, closeErr) {
		t.Fatalf("initialization = %v, %v; want both initialization and cleanup causes", database, err)
	}
	if got := fixture.closes.Load(); got != 1 {
		t.Fatalf("connection closes = %d, want 1", got)
	}
}

func TestSQLiteInitializationTransfersSuccessfulOwnership(t *testing.T) {
	fixture := &initializationConnector{}
	handle := sql.OpenDB(fixture)
	t.Cleanup(func() { _ = handle.Close() })

	database, err := initializeSQLite(handle)
	if err != nil {
		t.Fatalf("initialize database: %v", err)
	}
	if got := fixture.closes.Load(); got != 0 {
		t.Fatalf("successful initialization closed connection %d times", got)
	}
	var foreignKeys int
	if err := handle.QueryRowContext(context.Background(), "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("query foreign keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign keys = %d, want 1", foreignKeys)
	}
	if _, err := handle.ExecContext(context.Background(),
		"INSERT INTO tools (id, slug, name) VALUES ('fixture', 'fixture', 'Fixture')"); err != nil {
		t.Fatalf("write initialized schema: %v", err)
	}
	if _, err := handle.ExecContext(context.Background(),
		"INSERT INTO install_instructions (id, tool_id, executable_name) VALUES ('install', 'fixture', 'fixture')"); err != nil {
		t.Fatalf("write migrated column: %v", err)
	}
	if err := database.Close(); err != nil {
		t.Fatalf("close initialized database: %v", err)
	}
	if got := fixture.closes.Load(); got != 1 {
		t.Fatalf("connection closes after caller close = %d, want 1", got)
	}
}
