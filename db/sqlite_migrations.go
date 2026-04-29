package db

import (
	"context"
)

func (s *SQLiteDB) runMigrations() error {
	if err := s.addColumnIfNotExists("tools", "tool_of_the_week", "BOOLEAN DEFAULT false"); err != nil {
		return err
	}

	return s.addColumnIfNotExists("install_instructions", "executable_name", "TEXT DEFAULT ''")
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
