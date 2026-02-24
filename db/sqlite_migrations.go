package db

import (
	"context"
	"database/sql"
)

func (s *SQLiteDB) runMigrations() error {
	return s.addColumnIfNotExists("tools", "tool_of_the_week", "BOOLEAN DEFAULT false")
}

func (s *SQLiteDB) addColumnIfNotExists(table, column, definition string) error {
	var exists int
	err := s.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
		table, column,
	).Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		return nil
	}

	_, err = s.db.ExecContext(
		context.Background(),
		`ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition,
	)
	return err
}

func (s *SQLiteDB) columnExists(table, column string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(
		context.Background(),
		`SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`,
		table, column,
	).Scan(&count)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return count > 0, err
}
