package db

import (
	"fmt"
	"strings"
)

const (
	filterFieldInstalled = "installed"
	filterFieldName      = "name"
)

// BuildWhereClause converts a Filter AST to a SQL candidate predicate. Filters
// containing installed status still require runtime evaluation in Search.
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

// buildFilterSQL keeps candidates that can match in either installed state.
// Substitute the whole expression twice: replacing individual installed leaves
// with true would discard valid candidates under NOT.
func buildFilterSQL(filter *Filter) (string, []interface{}) {
	if !hasInstalledFilter(filter) {
		return buildFilterSQLForInstalled(filter, false)
	}
	installedClause, installedArgs := buildFilterSQLForInstalled(filter, true)
	uninstalledClause, uninstalledArgs := buildFilterSQLForInstalled(filter, false)
	return fmt.Sprintf("(%s OR %s)", installedClause, uninstalledClause), append(installedArgs, uninstalledArgs...)
}

// buildFilterSQLForInstalled evaluates the entire AST for a known installed
// state, leaving SQLite to evaluate text and tag predicates with its own rules.
func buildFilterSQLForInstalled(filter *Filter, installed bool) (string, []interface{}) {
	if filter == nil {
		return "", nil
	}

	switch filter.Type {
	case FilterAnd:
		leftClause, leftArgs := buildFilterSQLForInstalled(filter.Left, installed)
		rightClause, rightArgs := buildFilterSQLForInstalled(filter.Right, installed)
		args := append(leftArgs, rightArgs...)

		return fmt.Sprintf("(%s AND %s)", leftClause, rightClause), args

	case FilterOr:
		leftClause, leftArgs := buildFilterSQLForInstalled(filter.Left, installed)
		rightClause, rightArgs := buildFilterSQLForInstalled(filter.Right, installed)
		args := append(leftArgs, rightArgs...)

		return fmt.Sprintf("(%s OR %s)", leftClause, rightClause), args

	case FilterNot:
		innerClause, innerArgs := buildFilterSQLForInstalled(filter.Left, installed)

		return fmt.Sprintf("NOT (%s)", innerClause), innerArgs

	case FilterField:
		if strings.EqualFold(filter.Field, filterFieldInstalled) {
			wantInstalled := filter.Value == "true" || filter.Value == "1"
			if installed == wantInstalled {
				return "1=1", nil
			}
			return "1=0", nil
		}
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

	if filter.Type == FilterField && strings.EqualFold(filter.Field, filterFieldInstalled) {
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
