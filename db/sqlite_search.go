package db

import (
	"context"
	"fmt"
)

func (s *SQLiteDB) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
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

	whereClause, args := BuildWhereClause(opts.Filter, opts.Query)

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

	args = append(args, opts.Limit)

	rows, err := s.getDB().QueryContext(ctx, sqlQuery, args...)
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
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, t := range tools {
		installs, _ := s.GetInstallInstructions(t.ID)
		t.Installed = IsInstalled(&t, installs)
		results = append(results, SearchResult{Tool: t})
	}

	if opts.Filter != nil {
		results = filterByInstalled(results, opts.Filter)
	}

	return results, nil
}
