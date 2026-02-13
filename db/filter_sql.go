package db

import (
	"fmt"
	"strings"
)

const (
	filterFieldInstalled = "installed"
	filterValueTrue      = "true"
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
	case "name":
		return "name LIKE ?", []interface{}{"%" + value + "%"}
	case "tagline":
		return "tagline LIKE ?", []interface{}{"%" + value + "%"}
	case "language":
		return "language LIKE ?", []interface{}{"%" + value + "%"}
	case filterFieldInstalled:
		// Special case: handled in Go after query, but needs to return a clause
		// for AND/OR combinations to work properly
		return "1=1", nil
	default:
		// Unknown field - return always true to not filter out results
		return "1=1", nil
	}
}

// filterByInstalled handles the special 'installed' field filter (post-query)
func filterByInstalled(results []SearchResult, filter *Filter) []SearchResult {
	if filter == nil {
		return results
	}

	if !hasInstalledFilter(filter) {
		return results
	}

	value, negated := getInstalledFilterInfo(filter, false)

	var filtered []SearchResult
	for _, r := range results {
		wantInstalled := (value == filterValueTrue || value == "1")
		if negated {
			wantInstalled = !wantInstalled
		}
		if wantInstalled == r.Installed {
			filtered = append(filtered, r)
		}
	}

	return filtered
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
