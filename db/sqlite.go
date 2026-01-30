package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
}

func New(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	sqlite := &SQLiteDB{db: db}
	if err := sqlite.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return sqlite, nil
}

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
		`CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_language ON tools(language)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_slug ON tools(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_install_tool_id ON install_instructions(tool_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteDB) UpsertTool(ctx context.Context, tool *Tool) error {
	query := `
		INSERT INTO tools (id, slug, name, tagline, description, language, license, date_published, code_repository, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			slug = excluded.slug,
			name = excluded.name,
			tagline = excluded.tagline,
			description = excluded.description,
			language = excluded.language,
			license = excluded.license,
			date_published = excluded.date_published,
			code_repository = excluded.code_repository,
			updated_at = excluded.updated_at
	`

	tool.UpdatedAt = time.Now()
	_, err := s.db.ExecContext(ctx, query,
		tool.ID, tool.Slug, tool.Name, tool.Tagline, tool.Description,
		tool.Language, tool.License, tool.DatePublished, tool.CodeRepository, tool.UpdatedAt,
	)
	return err
}

func (s *SQLiteDB) UpsertInstallInstruction(ctx context.Context, inst *InstallInstruction) error {
	query := `
		INSERT INTO install_instructions (id, tool_id, platform, command)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			tool_id = excluded.tool_id,
			platform = excluded.platform,
			command = excluded.command
	`

	_, err := s.db.ExecContext(ctx, query, inst.ID, inst.ToolID, inst.Platform, inst.Command)
	return err
}

func (s *SQLiteDB) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	// Validate sort field to prevent SQL injection
	allowedFields := map[string]string{
		"name":     "name",
		"tagline":  "tagline",
		"language":  "language",
	}

	sortField, ok := allowedFields[opts.SortField]
	if !ok {
		sortField = "name"
	}

	sortOrder := "ASC"
	if opts.SortOrder == "DESC" || opts.SortOrder == "desc" {
		sortOrder = "DESC"
	}

	// Build WHERE clause from filters
	whereClause, args := BuildWhereClause(opts.Filter, opts.Query)

	// If no where clause, use default search
	if whereClause == "" {
		whereClause = "name LIKE ? OR tagline LIKE ? OR description LIKE ?"
		likeQuery := "%" + opts.Query + "%"
		args = []interface{}{likeQuery, likeQuery, likeQuery}
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, slug, name, tagline, description, language, license, date_published, code_repository
		FROM tools
		WHERE %s
		ORDER BY %s %s
		LIMIT ?
	`, whereClause, sortField, sortOrder)

	// Add limit to args
	args = append(args, opts.Limit)

	rows, err := s.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		err := rows.Scan(
			&r.ID, &r.Slug, &r.Name, &r.Tagline, &r.Description,
			&r.Language, &r.License, &r.DatePublished, &r.CodeRepository,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	// Handle installed filter (post-query filtering since it's not in DB)
	if opts.Filter != nil {
		results = filterByInstalled(results, opts.Filter)
	}

	return results, rows.Err()
}

func (s *SQLiteDB) GetAllTools(ctx context.Context) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, created_at, updated_at FROM tools`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Tagline, &t.Description,
			&t.Language, &t.License, &t.DatePublished, &t.CodeRepository,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}

	return tools, rows.Err()
}

func (s *SQLiteDB) GetToolBySlug(slug string) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, created_at, updated_at FROM tools WHERE slug = ?`
	rows, err := s.db.QueryContext(context.Background(), query, slug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var t Tool
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Tagline, &t.Description,
			&t.Language, &t.License, &t.DatePublished, &t.CodeRepository,
			&t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}

	return tools, rows.Err()
}

func (s *SQLiteDB) GetInstallInstructions(toolID string) ([]InstallInstruction, error) {
	query := `SELECT id, tool_id, platform, command, created_at FROM install_instructions WHERE tool_id = ?`
	rows, err := s.db.QueryContext(context.Background(), query, toolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insts []InstallInstruction
	for rows.Next() {
		var inst InstallInstruction
		err := rows.Scan(
			&inst.ID, &inst.ToolID, &inst.Platform, &inst.Command, &inst.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}

	return insts, rows.Err()
}

func (s *SQLiteDB) ToolCount(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tools").Scan(&count)
	return count, err
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}
