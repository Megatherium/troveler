package db

import (
	"context"
	"fmt"
)

const (
	sortOrderDesc      = "DESC"
	sortOrderDescLower = "desc"
	sortOrderAsc       = "ASC"
	sortFieldName      = "name"

	// Bound each installed-filter candidate batch to one install-instruction
	// query. This is a batch size, never a cap on the candidates searched.
	installedSearchBatchSize = sqliteVarLimit
)

type searchCandidate struct {
	tool               Tool
	matchesInstalled   bool
	matchesUninstalled bool
	sortValue          string
}

// Search queries tools matching opts, applying filters and sorting.
// SQLite sorts candidates; installed filters are resolved in bounded batches
// until opts.Limit matches are found or all candidates have been examined.
// A nonpositive limit returns all matching results.
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

	// COLLATE NOCASE preserves the case-insensitive sort behavior that the old
	// in-memory compareASC provided via strings.ToLower.
	// IDs break ties so successive batches never skip equal sort keys.
	orderByClause := "search_sort_value COLLATE NOCASE " + sortOrder + ", id ASC"

	sqlLimit := opts.Limit
	if sqlLimit <= 0 {
		sqlLimit = -1 // SQLite's unlimited LIMIT; CLI defaults are applied upstream.
	}
	hasInstalled := hasInstalledFilter(opts.Filter)
	sqlFilter := opts.Filter
	if hasInstalled {
		// Installed expressions are evaluated by the match flags below.
		sqlFilter = nil
	}
	whereClause, args := BuildWhereClause(sqlFilter, opts.Query)
	installedClause, uninstalledClause := "1=1", "1=1"
	if hasInstalled {
		sqlLimit = installedSearchBatchSize
		// Compute a match flag for each possible installed state. The outer
		// query keeps their union, and Go selects the flag for the actual state.
		// All text/tag comparisons remain in SQLite, including LIKE wildcards.
		var installedArgs, uninstalledArgs []interface{}
		installedClause, installedArgs = buildFilterSQLForInstalled(opts.Filter, true)
		uninstalledClause, uninstalledArgs = buildFilterSQLForInstalled(opts.Filter, false)
		args = append(append(installedArgs, uninstalledArgs...), args...)
	}

	queryTemplate := `
		SELECT * FROM (
			SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, tool_of_the_week,
				CASE WHEN %s THEN 1 ELSE 0 END AS matches_installed,
				CASE WHEN %s THEN 1 ELSE 0 END AS matches_uninstalled,
				%s AS search_sort_value
			FROM tools
			WHERE %s
		)
		WHERE (matches_installed OR matches_uninstalled) AND (%s)
		ORDER BY %s
		LIMIT ?
	`

	// Cache PATH checks across batches, not just within each batch.
	pathCache := make(map[string]bool)
	var results []SearchResult
	var last *searchCandidate
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cursorClause := "1=1"
		pageArgs := append([]interface{}{}, args...)
		if last != nil {
			comparison := ">"
			if sortOrder == sortOrderDesc {
				comparison = "<"
			}
			cursorClause = fmt.Sprintf(`search_sort_value COLLATE NOCASE %s ? OR
				(search_sort_value COLLATE NOCASE = ? AND id > ?)`, comparison)
			pageArgs = append(pageArgs, last.sortValue, last.sortValue, last.tool.ID)
		}
		pageArgs = append(pageArgs, sqlLimit)
		sqlQuery := fmt.Sprintf(queryTemplate, installedClause, uninstalledClause,
			sortField, whereClause, cursorClause, orderByClause)
		candidates, err := s.querySearchCandidates(ctx, sqlQuery, pageArgs)
		if err != nil {
			return nil, err
		}
		if len(candidates) == 0 {
			return results, nil
		}

		// Candidate rows are closed before querying installs: SQLiteDB has a
		// single connection, so keeping that cursor open would deadlock.
		toolIDs := make([]string, len(candidates))
		for i, c := range candidates {
			toolIDs[i] = c.tool.ID
		}
		installsByTool, err := s.GetInstallInstructionsBatch(ctx, toolIDs)
		if err != nil {
			return nil, err
		}
		for _, installs := range installsByTool {
			for _, inst := range installs {
				name := resolveExecutableName(inst)
				if _, cached := pathCache[name]; !cached && name != "" {
					pathCache[name] = isCommandAvailable(name)
				}
			}
		}

		// Select the full expression's match flag for the actual installed state.
		for _, c := range candidates {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			t := c.tool
			t.Installed = IsInstalledCached(&t, installsByTool[t.ID], pathCache)
			match := c.matchesUninstalled
			if t.Installed {
				match = c.matchesInstalled
			}
			if !match {
				continue
			}
			results = append(results, SearchResult{Tool: t})
			if opts.Limit > 0 && len(results) >= opts.Limit {
				return results, nil
			}
		}

		if !hasInstalled || len(candidates) < sqlLimit {
			return results, nil
		}
		last = &candidates[len(candidates)-1]
	}
}

// querySearchCandidates releases its rows before the caller queries installs.
func (s *SQLiteDB) querySearchCandidates(ctx context.Context, query string, args []interface{}) ([]searchCandidate, error) {
	rows, err := s.getDB().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var candidates []searchCandidate
	for rows.Next() {
		var c searchCandidate
		t := &c.tool
		if err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Tagline, &t.Description,
			&t.Language, &t.License, &t.DatePublished, &t.CodeRepository, &t.ToolOfTheWeek,
			&c.matchesInstalled, &c.matchesUninstalled, &c.sortValue,
		); err != nil {
			return nil, err
		}
		candidates = append(candidates, c)
	}
	return candidates, rows.Err()
}
