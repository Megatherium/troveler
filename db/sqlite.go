package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

	// SQLite only enforces foreign keys on the connection that enables them.
	// Limit to one connection so the PRAGMA applies to all operations.
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

func (s *SQLiteDB) GetTags(slug string) ([]string, error) {
	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return nil, err
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	query := `
		SELECT tt.tag_name
		FROM tool_tags tt
		WHERE tt.tool_id = ?
		ORDER BY tt.tag_name
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
		SELECT tt.tag_name, COUNT(tt.tool_id) as count
		FROM tool_tags tt
		GROUP BY tt.tag_name
		ORDER BY tt.tag_name
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

func normalizeTagName(tagName string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(tagName))
	if normalized == "" {
		return "", fmt.Errorf("tag name cannot be empty")
	}
	return normalized, nil
}

func (s *SQLiteDB) pruneOrphanedTags() error {
	_, err := s.db.ExecContext(context.Background(),
		"DELETE FROM tags WHERE name NOT IN (SELECT DISTINCT tag_name FROM tool_tags)")
	return err
}

func (s *SQLiteDB) AddTag(slug, tagName string) error {
	normalized, err := normalizeTagName(tagName)
	if err != nil {
		return err
	}

	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	_, err = s.db.ExecContext(context.Background(),
		"INSERT OR IGNORE INTO tags (name) VALUES (?)", normalized)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	_, err = s.db.ExecContext(context.Background(),
		"INSERT OR IGNORE INTO tool_tags (tool_id, tag_name) VALUES (?, ?)", toolID, normalized)
	return err
}

func (s *SQLiteDB) RemoveTag(slug, tagName string) error {
	normalized, err := normalizeTagName(tagName)
	if err != nil {
		return err
	}

	tools, err := s.GetToolBySlug(slug)
	if err != nil {
		return err
	}
	if len(tools) == 0 {
		return fmt.Errorf("tool not found: %s", slug)
	}
	toolID := tools[0].ID

	result, err := s.db.ExecContext(context.Background(),
		"DELETE FROM tool_tags WHERE tool_id = ? AND tag_name = ?", toolID, normalized)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("tag not found on tool: %s", normalized)
	}

	return s.pruneOrphanedTags()
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
	if err != nil {
		return err
	}

	return s.pruneOrphanedTags()
}

func (s *SQLiteDB) GetToolsByTag(tagName string) ([]Tool, error) {
	normalized, err := normalizeTagName(tagName)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT t.id, t.slug, t.name, t.tagline, t.description, t.language, t.license,
			t.date_published, t.code_repository, t.created_at, t.updated_at
		FROM tools t
		JOIN tool_tags tt ON t.id = tt.tool_id
		WHERE tt.tag_name = ?
		ORDER BY t.name
	`
	rows, err := s.db.QueryContext(context.Background(), query, normalized)
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

// GetAllTagsBySlug returns all user tags keyed by tool slug.
// This is used to preserve tags during database updates.
func (s *SQLiteDB) GetAllTagsBySlug() (map[string][]string, error) {
	query := `
		SELECT t.slug, tt.tag_name
		FROM tool_tags tt
		JOIN tools t ON tt.tool_id = t.id
		ORDER BY t.slug, tt.tag_name
	`
	rows, err := s.db.QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string][]string)
	for rows.Next() {
		var slug, tagName string
		if err := rows.Scan(&slug, &tagName); err != nil {
			return nil, err
		}
		if slug == "" {
			continue
		}
		result[slug] = append(result[slug], tagName)
	}

	return result, rows.Err()
}

// ReapplyTags restores user tags to tools after a database update.
// It takes a map of slug -> tag names that was captured before the update.
func (s *SQLiteDB) ReapplyTags(slugToTags map[string][]string) error {
	for slug, tagNames := range slugToTags {
		tools, err := s.GetToolBySlug(slug)
		if err != nil {
			return fmt.Errorf("lookup tool by slug %q: %w", slug, err)
		}
		if len(tools) == 0 {
			continue
		}
		toolID := tools[0].ID

		for _, tagName := range tagNames {
			_, err = s.db.ExecContext(context.Background(),
				"INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
			if err != nil {
				return fmt.Errorf("create tag %q: %w", tagName, err)
			}

			_, err = s.db.ExecContext(context.Background(),
				"INSERT OR IGNORE INTO tool_tags (tool_id, tag_name) VALUES (?, ?)", toolID, tagName)
			if err != nil {
				return fmt.Errorf("link tag %q to tool %q: %w", tagName, slug, err)
			}
		}
	}

	return nil
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}
