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

	whereClause, args := BuildWhereClause(opts.Filter, opts.Query)

	if whereClause == "" {
		whereClause = "name LIKE ? OR tagline LIKE ? OR description LIKE ?"
		likeQuery := "%" + opts.Query + "%"
		args = []interface{}{likeQuery, likeQuery, likeQuery}
	}

	// Determine SQL LIMIT: over-fetch when installed filter is active
	// since we can't filter by installed in SQL.
	sqlLimit := opts.Limit
	hasInstalled := hasInstalledFilter(opts.Filter)
	if hasInstalled {
		sqlLimit = overfetchLimit(opts.Limit)
	}

	sqlQuery := fmt.Sprintf(`
		SELECT id, slug, name, tagline, description, language, license, date_published, code_repository, tool_of_the_week
		FROM tools
		WHERE %s
		ORDER BY %s
		LIMIT ?
	`, whereClause, orderByClause)

	args = append(args, sqlLimit)

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

	// Batch-fetch install instructions (1 query instead of N)
	toolIDs := make([]string, len(tools))
	for i, t := range tools {
		toolIDs[i] = t.ID
	}
	installsByTool, err := s.GetInstallInstructionsBatch(ctx, toolIDs)
	if err != nil {
		return nil, err
	}

	// Build LookPath cache (deduplicated — one stat per unique executable name)
	pathCache := BuildLookPathCache(installsByTool)

	// Resolve installed status and apply installed filter in one pass.
	// Results are already sorted by SQLite, so we just need to filter
	// by installed status and trim to the requested limit.
	//
	// When installed appears in an OR context (e.g., OR(installed=true, language=go)),
	// the SQL WHERE already handles the non-installed branch. Applying a Go-side
	// filter would incorrectly exclude those results, so we skip it.
	skipGoFilter := hasInstalled && hasInstalledInOrContext(opts.Filter)
	wantInstalled, negated := false, false
	if hasInstalled && !skipGoFilter {
		wantInstalled, negated = installedFilterValue(opts.Filter)
	}

	var results []SearchResult
	for _, t := range tools {
		installs := installsByTool[t.ID]
		t.Installed = IsInstalledCached(&t, installs, pathCache)

		if hasInstalled && !skipGoFilter {
			match := t.Installed == wantInstalled
			if negated {
				match = !match
			}
			if !match {
				continue
			}
		}

		results = append(results, SearchResult{Tool: t})

		// Early exit: we have enough results after filtering
		if hasInstalled && !skipGoFilter && opts.Limit > 0 && len(results) >= opts.Limit {
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

// installedFilterValue extracts the desired installed state from the filter AST.
// Returns (value, negated) where value is true if the filter wants installed=true.
func installedFilterValue(filter *Filter) (bool, bool) {
	if filter == nil {
		return false, false
	}
	val, negated := getInstalledFilterInfo(filter, false)
	wantInstalled := (val == "true" || val == "1")
	return wantInstalled, negated
}
