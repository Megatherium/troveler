package search

import (
	"context"
	"fmt"

	"troveler/db"
)

// Service handles tool search operations
type Service struct {
	db *db.SQLiteDB
}

// NewService creates a new search service
func NewService(database *db.SQLiteDB) *Service {
	return &Service{db: database}
}

// Options contains search configuration
type Options struct {
	Query     string
	Limit     int
	SortField string
	SortOrder string // ASC or DESC
}

// Result represents a search result with metadata
type Result struct {
	Tools      []db.SearchResult
	TotalCount int
	Query      string
	SortField  string
	SortOrder  string
}

// ValidSortFields defines allowed sort fields
var ValidSortFields = map[string]bool{
	"name":     true,
	"tagline":  true,
	"language": true,
}

// Search performs a tool search with:: given options
func (s *Service) Search(ctx context.Context, opts Options) (*Result, error) {
	// Parse filters from query
	filter, searchTerm, err := ParseFilters(opts.Query)
	if err != nil {
		return nil, fmt.Errorf("invalid filter syntax: %w", err)
	}

	// Apply defaults
	if opts.Limit <= 0 {
		opts.Limit = 50
	}

	// Validate and default sort field
	if !ValidSortFields[opts.SortField] {
		opts.SortField = "name"
	}

	// Normalize sort order
	if opts.SortOrder != "DESC" && opts.SortOrder != "desc" {
		opts.SortOrder = "ASC"
	}

	// Perform search
	dbOpts := db.SearchOptions{
		Query:     searchTerm,
		Limit:     opts.Limit,
		SortField: opts.SortField,
		SortOrder: opts.SortOrder,
		Filter:    filter,
	}

	tools, err := s.db.Search(ctx, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return &Result{
		Tools:      tools,
		TotalCount: len(tools),
		Query:      searchTerm,
		SortField:  opts.SortField,
		SortOrder:  opts.SortOrder,
	}, nil
}

// SearchAll returns all tools (for initial TUI load)
func (s *Service) SearchAll(ctx context.Context, limit int) (*Result, error) {
	return s.Search(ctx, Options{
		Query:     "",
		Limit:     limit,
		SortField: "name",
		SortOrder: "ASC",
	})
}
