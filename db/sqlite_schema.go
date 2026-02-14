package db

import "context"

func (s *SQLiteDB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tools (
			id TEXT PRIMARY KEY,
			slug TEXT UNIQUE,
			name TEXT,
			tagline TEXT,
			description TEXT,
			language TEXT,
			license TEXT,
			date_published TEXT,
			code_repository TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS install_instructions (
			id TEXT PRIMARY KEY,
			tool_id TEXT REFERENCES tools(id),
			platform TEXT,
			command TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tags (
			name TEXT PRIMARY KEY COLLATE NOCASE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_tags (
			tool_id TEXT REFERENCES tools(id) ON DELETE CASCADE,
			tag_name TEXT REFERENCES tags(name) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tool_id, tag_name)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_language ON tools(language)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_slug ON tools(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_install_tool_id ON install_instructions(tool_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_tags_tag_name ON tool_tags(tag_name)`,
	}

	for _, q := range queries {
		if _, err := s.getDB().ExecContext(context.Background(), q); err != nil {
			return err
		}
	}

	return nil
}
