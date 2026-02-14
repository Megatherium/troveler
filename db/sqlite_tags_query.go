package db

import (
	"context"
	"fmt"
)

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
	rows, err := s.getDB().QueryContext(context.Background(), query, toolID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	rows, err := s.getDB().QueryContext(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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
	rows, err := s.getDB().QueryContext(context.Background(), query, normalized)
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

func (s *SQLiteDB) GetAllTagsBySlug() (map[string][]string, error) {
	query := `
		SELECT t.slug, tt.tag_name
		FROM tool_tags tt
		JOIN tools t ON tt.tool_id = t.id
		ORDER BY t.slug, tt.tag_name
	`
	rows, err := s.getDB().QueryContext(context.Background(), query)
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
