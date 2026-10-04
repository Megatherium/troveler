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

	// installedOverfetchFactor controls how many extra rows we fetch when
	// the installed filter is active. Since we can't filter by "installed"
	// in SQL (it requires a runtime LookPath check), we over-fetch by this
	// multiplier and then trim in Go. A factor of 4 means we fetch 4x the
	// requested limit, which is sufficient for most real-world installed ratios.
	installedOverfetchFactor = 4

	// installedOverfetchMax caps the over-fetch to avoid pulling excessive rows
	// even with large limits. Only applies when the caller specifies a limit > 0.
	installedOverfetchMax = 500

	// installedNoLimitFallback is used when the caller requests limit=0 (no limit)
	// and the installed filter is active. We must fetch all rows since we can't
	// predict how many will be filtered out.
	installedNoLimitFallback = 100000
)

// Search queries tools matching opts, applying filters and sorting.
// Sorting and limiting are always pushed to SQLite via ORDER BY / LIMIT.
// The "installed" filter is resolved in Go (requires exec.LookPath) using
// an over-fetch strategy: we fetch more rows than requested, resolve
// installed status, filter, and trim to the actual limit.
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
	orderByClause := sortField + " COLLATE NOCASE " + sortOrder

	// Determine SQL LIMIT: over-fetch when installed filter is active
	// since we can't filter by installed in SQL.
	sqlLimit := opts.Limit
	hasInstalled := hasInstalledFilter(opts.Filter)
	sqlFilter := opts.Filter
	if hasInstalled {
		// Installed expressions are evaluated by the match flags below.
		sqlFilter = nil
	}
	whereClause, args := BuildWhereClause(sqlFilter, opts.Query)
	installedClause, uninstalledClause := "1=1", "1=1"
	if hasInstalled {
		sqlLimit = overfetchLimit(opts.Limit)
		// Compute a match flag for each possible installed state. The outer
		// query keeps their union, and Go selects the flag for the actual state.
		// All text/tag comparisons remain in SQLite, including LIKE wildcards.
		var installedArgs, uninstalledArgs []interface{}
		installedClause, installedArgs = buildFilterSQLForInstalled(opts.Filter, true)
		uninstalledClause, uninstalledArgs = buildFilterSQLForInstalled(opts.Filter, false)
		args = append(append(installedArgs, uninstalledArgs...), args...)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT * FROM (
			SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, tool_of_the_week,
				CASE WHEN %s THEN 1 ELSE 0 END AS matches_installed,
				CASE WHEN %s THEN 1 ELSE 0 END AS matches_uninstalled
			FROM tools
			WHERE %s
		)
		WHERE matches_installed OR matches_uninstalled
		ORDER BY %s
		LIMIT ?
	`, installedClause, uninstalledClause, whereClause, orderByClause)

	args = append(args, sqlLimit)

	rows, err := s.getDB().QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	type candidate struct {
		tool               Tool
		matchesInstalled   bool
		matchesUninstalled bool
	}
	var tools []candidate
	for rows.Next() {
		var c candidate
		t := &c.tool
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Tagline, &t.Description,
			&t.Language, &t.License, &t.DatePublished, &t.CodeRepository, &t.ToolOfTheWeek,
			&c.matchesInstalled, &c.matchesUninstalled,
		)
		if err != nil {
			return nil, err
		}
		tools = append(tools, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Batch-fetch install instructions (1 query instead of N)
	toolIDs := make([]string, len(tools))
	for i, t := range tools {
		toolIDs[i] = t.tool.ID
	}
	installsByTool, err := s.GetInstallInstructionsBatch(ctx, toolIDs)
	if err != nil {
		return nil, err
	}

	// Build LookPath cache (deduplicated — one stat per unique executable name)
	pathCache := BuildLookPathCache(installsByTool)

	// Select the complete expression's match flag for the actual installed state.
	// Candidates are already sorted by SQLite.
	var results []SearchResult
	for _, c := range tools {
		t := c.tool
		installs := installsByTool[t.ID]
		t.Installed = IsInstalledCached(&t, installs, pathCache)

		match := c.matchesUninstalled
		if t.Installed {
			match = c.matchesInstalled
		}
		if !match {
			continue
		}

		results = append(results, SearchResult{Tool: t})

		// Early exit: we have enough results after filtering
		if hasInstalled && opts.Limit > 0 && len(results) >= opts.Limit {
			break
		}
	}

	return results, nil
}

// overfetchLimit computes the SQL LIMIT when the installed filter is active.
// We fetch more rows than requested to compensate for rows that will be
// filtered out by the Go-side installed check.
// When requested is 0 (no limit), we fall back to a large cap since we
// can't predict how many rows the installed filter will discard.
func overfetchLimit(requested int) int {
	if requested <= 0 {
		return installedNoLimitFallback
	}
	limit := requested * installedOverfetchFactor
	if limit > installedOverfetchMax {
		limit = installedOverfetchMax
	}
	if limit < requested {
		limit = requested
	}
	return limit
}
