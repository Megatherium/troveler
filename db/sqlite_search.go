package db

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

const (
	defaultUnfilteredLimit = 100000
	sortOrderDesc          = "DESC"
	sortOrderDescLower     = "desc"
	sortOrderAsc           = "ASC"
	sortFieldName          = "name"
)

// Search queries tools matching opts, applying filters and sorting.
func (s *SQLiteDB) Search(ctx context.Context, opts SearchOptions) ([]SearchResult, error) {
	allowedFields := map[string]string{
		"name":           "name",
		"tagline":        "tagline",
		"language":       "language",
		"date_published": "date_published",
	}

	sortField, ok := allowedFields[opts.SortField]
	if !ok {
		sortField = sortFieldName
	}

	sortOrder := sortOrderAsc
	if opts.SortOrder == sortOrderDesc || opts.SortOrder == sortOrderDescLower {
		sortOrder = sortOrderDesc
	}

	hasInstalled := hasInstalledFilter(opts.Filter)

	whereClause, args := BuildWhereClause(opts.Filter, opts.Query)

	if whereClause == "" {
		whereClause = "name LIKE ? OR tagline LIKE ? OR description LIKE ?"
		likeQuery := "%" + opts.Query + "%"
		args = []interface{}{likeQuery, likeQuery, likeQuery}
	}

	limit := opts.Limit
	if hasInstalled {
		limit = defaultUnfilteredLimit
	}

	var sqlQuery string
	if hasInstalled {
		sqlQuery = fmt.Sprintf(`
			SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, tool_of_the_week
			FROM tools
			WHERE %s
			LIMIT ?
		`, whereClause)
	} else {
		sqlQuery = fmt.Sprintf(`
			SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, tool_of_the_week
			FROM tools
			WHERE %s
			ORDER BY %s %s
			LIMIT ?
		`, whereClause, sortField, sortOrder)
	}

	args = append(args, limit)

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
			&t.Language, &t.License, &t.DatePublished, &t.CodeRepository, &t.ToolOfTheWeek,
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

	if hasInstalled {
		results = filterByInstalled(results, opts.Filter)
		results = sortAndLimitResults(results, opts.SortField, sortOrder, opts.Limit)
	}

	return results, nil
}

func sortAndLimitResults(results []SearchResult, sortField, sortOrder string, limit int) []SearchResult {
	if len(results) == 0 {
		return results
	}

	sortField = strings.ToLower(sortField)

	switch sortField {
	case sortFieldName:
		sort.Slice(results, func(i, j int) bool {
			return compareASC(results[i].Name, results[j].Name, sortOrder == sortOrderDesc)
		})
	case "tagline":
		sort.Slice(results, func(i, j int) bool {
			return compareASC(results[i].Tagline, results[j].Tagline, sortOrder == sortOrderDesc)
		})
	case "language":
		sort.Slice(results, func(i, j int) bool {
			return compareASC(results[i].Language, results[j].Language, sortOrder == sortOrderDesc)
		})
	case "date_published":
		sort.Slice(results, func(i, j int) bool {
			return compareASC(results[i].DatePublished, results[j].DatePublished, sortOrder == sortOrderDesc)
		})
	default:
		sort.Slice(results, func(i, j int) bool {
			return compareASC(results[i].Name, results[j].Name, sortOrder == sortOrderDesc)
		})
	}

	if limit > 0 && limit < len(results) {
		return results[:limit]
	}

	return results
}

func compareASC(a, b string, reverse bool) bool {
	lowA := strings.ToLower(a)
	lowB := strings.ToLower(b)
	if lowA < lowB {
		return !reverse
	}
	if lowA > lowB {
		return reverse
	}

	return false
}
