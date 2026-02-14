package db

import (
	"context"
	"time"
)

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
	_, err := s.getDB().ExecContext(ctx, query,
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

	_, err := s.getDB().ExecContext(ctx, query, inst.ID, inst.ToolID, inst.Platform, inst.Command)

	return err
}

func (s *SQLiteDB) GetAllTools(ctx context.Context) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license,
		date_published, code_repository, created_at, updated_at FROM tools`
	rows, err := s.getDB().QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	rows, err := s.getDB().QueryContext(context.Background(), query, slug)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	rows, err := s.getDB().QueryContext(context.Background(), query, toolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	err := s.getDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM tools").Scan(&count)

	return count, err
}
