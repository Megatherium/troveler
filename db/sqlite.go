package db

import (
	"context"
	"database/sql"
	"errors"
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

	if err := db.PingContext(context.Background()); err != nil {
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
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_tags (
			tool_id TEXT REFERENCES tools(id) ON DELETE CASCADE,
			tag_id TEXT REFERENCES tags(id) ON DELETE CASCADE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tool_id, tag_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_name ON tools(name)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_language ON tools(language)`,
		`CREATE INDEX IF NOT EXISTS idx_tools_slug ON tools(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_install_tool_id ON install_instructions(tool_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tool_tags_tag_id ON tool_tags(tag_id)`,
	}

	for _, q := range queries {
		if _, err := s.db.ExecContext(context.Background(), q); err != nil {
			return err
		}
	}

	return nil
}

func (s *SQLiteDB) UpsertTool(ctx context.Context, tool *Tool) error {
	query := `
		INSERT INTO tools (id, slug, name, tagline, description, language, license,
			date_published, code_repository, updated_at)
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
		"language": "language",
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
		// Populate installed status by checking install instructions
		installs, _ := s.GetInstallInstructions(r.ID)
		r.Installed = IsInstalled(&r.Tool, installs)
		results = append(results, r)
	}

	// Handle installed filter (post-query filtering since it's not in DB)
	if opts.Filter != nil {
		results = filterByInstalled(results, opts.Filter)
	}

	return results, rows.Err()
}

func (s *SQLiteDB) GetAllTools(ctx context.Context) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license,
		date_published, code_repository, created_at, updated_at FROM tools`
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
	query := `SELECT id, slug, name, tagline, description, language, license,
		date_published, code_repository, created_at, updated_at FROM tools WHERE slug = ?`
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

func (s *SQLiteDB) GetTags(toolID string) ([]string, error) {
	query := `
		SELECT t.name
		FROM tags t
		JOIN tool_tags tt ON t.id = tt.tag_id
		WHERE tt.tool_id = ?
		ORDER BY t.name
	`
	rows, err := s.db.QueryContext(context.Background(), query, toolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

func (s *SQLiteDB) GetAllTags() ([]TagCount, error) {
	query := `
		SELECT t.name, COUNT(tt.tool_id) as count
		FROM tags t
		LEFT JOIN tool_tags tt ON t.id = tt.tag_id
		GROUP BY t.id, t.name
		ORDER BY t.name
	`
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []TagCount
	for rows.Next() {
		var tc TagCount
		if err := rows.Scan(&tc.Name, &tc.Count); err != nil {
			return nil, err
		}
		tags = append(tags, tc)
	}
	return tags, rows.Err()
}

func (s *SQLiteDB) AddTag(slug, tagName string) error {
	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	var tagID string
	err = s.db.QueryRowContext(context.Background(),
		"SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if errors.Is(err, sql.ErrNoRows) {
		tagID = fmt.Sprintf("tag-%s", tagName)
		_, err = s.db.ExecContext(context.Background(),
			"INSERT INTO tags (id, name) VALUES (?, ?)", tagID, tagName)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to lookup tag: %w", err)
	}

	_, err = s.db.ExecContext(context.Background(),
		"INSERT OR IGNORE INTO tool_tags (tool_id, tag_id) VALUES (?, ?)", toolID, tagID)
	return err
}

func (s *SQLiteDB) RemoveTag(slug, tagName string) error {
	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	result, err := s.db.ExecContext(context.Background(),
		`DELETE FROM tool_tags WHERE tool_id = ? AND tag_id = (
			SELECT id FROM tags WHERE name = ?
		)`, toolID, tagName)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("tag not found on tool: %s", tagName)
	}
	return nil
}

func (s *SQLiteDB) ClearTags(slug string) error {
	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	_, err = s.db.ExecContext(context.Background(),
		"DELETE FROM tool_tags WHERE tool_id = ?", toolID)
	return err
}

func (s *SQLiteDB) GetToolsByTag(tagName string) ([]Tool, error) {
	query := `
		SELECT t.id, t.slug, t.name, t.tagline, t.description, t.language, t.license,
			t.date_published, t.code_repository, t.created_at, t.updated_at
		FROM tools t
		JOIN tool_tags tt ON t.id = tt.tool_id
		JOIN tags tag ON tt.tag_id = tag.id
		WHERE tag.name = ?
		ORDER BY t.name
	`
	rows, err := s.db.QueryContext(context.Background(), query, tagName)
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

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}
