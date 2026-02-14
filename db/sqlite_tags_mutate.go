package db

import (
	"context"
	"fmt"
	"strings"
)

func normalizeTagName(tagName string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(tagName))
	if normalized == "" {
		return "", fmt.Errorf("tag name cannot be empty")
	}

	return normalized, nil
}

func (s *SQLiteDB) pruneOrphanedTags() error {
	_, err := s.getDB().ExecContext(context.Background(),
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

	_, err = s.getDB().ExecContext(context.Background(),
		"INSERT OR IGNORE INTO tags (name) VALUES (?)", normalized)
	if err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	_, err = s.getDB().ExecContext(context.Background(),
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

	result, err := s.getDB().ExecContext(context.Background(),
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

	_, err = s.getDB().ExecContext(context.Background(),
		"DELETE FROM tool_tags WHERE tool_id = ?", toolID)
	if err != nil {
		return err
	}

	return s.pruneOrphanedTags()
}

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
			_, err = s.getDB().ExecContext(context.Background(),
				"INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
			if err != nil {
				return fmt.Errorf("create tag %q: %w", tagName, err)
			}

			_, err = s.getDB().ExecContext(context.Background(),
				"INSERT OR IGNORE INTO tool_tags (tool_id, tag_name) VALUES (?, ?)", toolID, tagName)
			if err != nil {
				return fmt.Errorf("link tag %q to tool %q: %w", tagName, slug, err)
			}
		}
	}

	return nil
}
