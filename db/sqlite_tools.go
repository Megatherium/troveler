package db

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// UpsertTool inserts or updates a tool record.
func (s *SQLiteDB) UpsertTool(ctx context.Context, tool *Tool) error {
	query := `
		INSERT INTO tools (id, slug, name, tagline, description, language, license,
			date_published, code_repository, tool_of_the_week, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			slug = excluded.slug,
			name = excluded.name,
			tagline = excluded.tagline,
			description = excluded.description,
			language = excluded.language,
			license = excluded.license,
			date_published = excluded.date_published,
			code_repository = excluded.code_repository,
			tool_of_the_week = excluded.tool_of_the_week,
			updated_at = excluded.updated_at
	`

	tool.UpdatedAt = time.Now()
	_, err := s.getDB().ExecContext(ctx, query,
		tool.ID, tool.Slug, tool.Name, tool.Tagline, tool.Description,
		tool.Language, tool.License, tool.DatePublished, tool.CodeRepository, tool.ToolOfTheWeek, tool.UpdatedAt,
	)

	return err
}

// UpsertInstallInstruction inserts or updates an install instruction.
func (s *SQLiteDB) UpsertInstallInstruction(ctx context.Context, inst *InstallInstruction) error {
	query := `
		INSERT INTO install_instructions (id, tool_id, platform, command, executable_name)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			tool_id = excluded.tool_id,
			platform = excluded.platform,
			command = excluded.command,
			executable_name = excluded.executable_name
	`

	_, err := s.getDB().ExecContext(ctx, query, inst.ID, inst.ToolID, inst.Platform, inst.Command, inst.ExecutableName)

	return err
}

// GetAllTools returns every tool in the database.
func (s *SQLiteDB) GetAllTools(ctx context.Context) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license,
		date_published, code_repository, tool_of_the_week, created_at, updated_at FROM tools`
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
			&t.ToolOfTheWeek, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}

	return tools, rows.Err()
}

// GetToolBySlug returns tools matching the given slug.
func (s *SQLiteDB) GetToolBySlug(slug string) ([]Tool, error) {
	query := `SELECT id, slug, name, tagline, description, language, license,
		date_published, code_repository, tool_of_the_week, created_at, updated_at FROM tools WHERE slug = ?`
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
			&t.ToolOfTheWeek, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}

	return tools, rows.Err()
}

// GetInstallInstructions returns all install instructions for a tool.
func (s *SQLiteDB) GetInstallInstructions(toolID string) ([]InstallInstruction, error) {
	query := `SELECT id, tool_id, platform, command, executable_name, created_at FROM install_instructions WHERE tool_id = ?`
	rows, err := s.getDB().QueryContext(context.Background(), query, toolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var insts []InstallInstruction
	for rows.Next() {
		var inst InstallInstruction
		err := rows.Scan(
			&inst.ID, &inst.ToolID, &inst.Platform, &inst.Command, &inst.ExecutableName, &inst.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		insts = append(insts, inst)
	}

	return insts, rows.Err()
}

const sqliteVarLimit = 500

// GetInstallInstructionsBatch returns install instructions for multiple tools
// in a single round-trip (chunked to respect SQLite's variable limit).
// Returns a map keyed by tool_id.
func (s *SQLiteDB) GetInstallInstructionsBatch(ctx context.Context, toolIDs []string) (map[string][]InstallInstruction, error) {
	result := make(map[string][]InstallInstruction, len(toolIDs))
	if len(toolIDs) == 0 {
		return result, nil
	}

	query := `SELECT id, tool_id, platform, command, executable_name, created_at FROM install_instructions WHERE tool_id IN (%s)`

	for i := 0; i < len(toolIDs); i += sqliteVarLimit {
		end := i + sqliteVarLimit
		if end > len(toolIDs) {
			end = len(toolIDs)
		}
		chunk := toolIDs[i:end]

		placeholders := strings.Repeat("?,", len(chunk))
		placeholders = placeholders[:len(placeholders)-1]

		args := make([]interface{}, len(chunk))
		for j, id := range chunk {
			args[j] = id
		}

		rows, err := s.getDB().QueryContext(ctx, fmt.Sprintf(query, placeholders), args...)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var inst InstallInstruction
			if err := rows.Scan(
				&inst.ID, &inst.ToolID, &inst.Platform, &inst.Command, &inst.ExecutableName, &inst.CreatedAt,
			); err != nil {
				_ = rows.Close()
				return nil, err
			}
			result[inst.ToolID] = append(result[inst.ToolID], inst)
		}
		_ = rows.Close()
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ToolCount returns the total number of tools.
func (s *SQLiteDB) ToolCount(ctx context.Context) (int, error) {
	var count int
	err := s.getDB().QueryRowContext(ctx, "SELECT COUNT(*) FROM tools").Scan(&count)

	return count, err
}
