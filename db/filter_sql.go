package db

import (
	"fmt"
	"strings"
)

const (
	filterFieldInstalled = "installed"
	filterFieldName      = "name"
)

// BuildWhereClause converts a Filter AST to SQL WHERE clauses
func BuildWhereClause(filter *Filter, searchTerm string) (string, []interface{}) {
	var clauses []string
	var args []interface{}

	// Add search term if provided
	if searchTerm != "" {
		clauses = append(clauses, "(name LIKE ? OR tagline LIKE ? OR description LIKE ?)")
		likeQuery := "%" + searchTerm + "%"
		args = append(args, likeQuery, likeQuery, likeQuery)
	}

	// Add filter clauses
	if filter != nil {
		filterClause, filterArgs := buildFilterSQL(filter)
		if filterClause != "" {
			if searchTerm != "" {
				clauses = append(clauses, "("+filterClause+")")
			} else {
				clauses = append(clauses, filterClause)
			}
			args = append(args, filterArgs...)
		}
	}

	// If no clauses (no search term and no non-empty filter), add default clause
	// This ensures we still query the DB for filtering in Go code (e.g., installed filter)
	if len(clauses) == 0 {
		clauses = append(clauses, "1=1")
	}

	whereClause := strings.Join(clauses, " AND ")

	return whereClause, args
}

// buildFilterSQL recursively builds SQL from Filter AST
func buildFilterSQL(filter *Filter) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	switch filter.Type {
	case FilterAnd:
		leftClause, leftArgs := buildFilterSQL(filter.Left)
		rightClause, rightArgs := buildFilterSQL(filter.Right)
		args := append(leftArgs, rightArgs...)

		return fmt.Sprintf("(%s AND %s)", leftClause, rightClause), args

	case FilterOr:
		leftClause, leftArgs := buildFilterSQL(filter.Left)
		rightClause, rightArgs := buildFilterSQL(filter.Right)
		args := append(leftArgs, rightArgs...)

		return fmt.Sprintf("(%s OR %s)", leftClause, rightClause), args

	case FilterNot:
		innerClause, innerArgs := buildFilterSQL(filter.Left)

		return fmt.Sprintf("NOT (%s)", innerClause), innerArgs

	case FilterField:
		return buildFieldFilter(filter.Field, filter.Value)

	default:
		return "", nil
	}
}

// buildFieldFilter creates SQL for a single field filter
func buildFieldFilter(field, value string) (string, []interface{}) {
	switch strings.ToLower(field) {
	case filterFieldName:
		return "name LIKE ?", []interface{}{"%" + value + "%"}
	case "tagline":
		return "tagline LIKE ?", []interface{}{"%" + value + "%"}
	case "language":
		return "language LIKE ?", []interface{}{"%" + value + "%"}
	case "tag":
		return "EXISTS (SELECT 1 FROM tool_tags WHERE tool_id = tools.id AND tag_name = ?)",
			[]interface{}{strings.ToLower(value)}
	case filterFieldInstalled:
		// Special case: handled in Go after query, but needs to return a clause
		// for AND/OR combinations to work properly
		return "1=1", nil
	default:
		// Unknown field - return always true to not filter out results
		return "1=1", nil
	}
}

// hasInstalledFilter checks if the filter AST contains an installed field filter
func hasInstalledFilter(filter *Filter) bool {
	if filter == nil {
		return false
	}

	if filter.Type == FilterField && filter.Field == filterFieldInstalled {
		return true
	}

	if filter.Left != nil && hasInstalledFilter(filter.Left) {
		return true
	}
	if filter.Right != nil && hasInstalledFilter(filter.Right) {
		return true
	}

	return false
}

// hasInstalledInOrContext checks whether the installed field filter appears
// inside an OR node. When installed is OR-combined with other conditions
// (e.g., OR(installed=true, language=go)), the SQL WHERE already handles the
// non-installed branch (language=go), and a Go-side installed filter would
// incorrectly exclude those results. In that case the Go-side filter must be
// skipped entirely.
func hasInstalledInOrContext(filter *Filter) bool {
	if filter == nil {
		return false
	}

	if filter.Type == FilterOr {
		if hasInstalledFilter(filter.Left) || hasInstalledFilter(filter.Right) {
			return true
		}
	}

	if hasInstalledInOrContext(filter.Left) {
		return true
	}

	return hasInstalledInOrContext(filter.Right)
}

// getInstalledFilterValue extracts the value from an installed filter
func getInstalledFilterInfo(filter *Filter, negated bool) (string, bool) {
	if filter == nil {
		return "", negated
	}

	if filter.Type == FilterNot {
		return getInstalledFilterInfo(filter.Left, !negated)
	}

	if filter.Type == FilterField && filter.Field == filterFieldInstalled {
		return filter.Value, negated
	}

	if value, n := getInstalledFilterInfo(filter.Left, negated); value != "" {
		return value, n
	}

	return getInstalledFilterInfo(filter.Right, negated)
}
