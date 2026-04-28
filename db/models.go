// Package db provides SQLite-backed storage for tools and install instructions.
package db

import "time"

// Tool represents a CLI tool entry in the database.
type Tool struct {
	ID             string    `json:"id" db:"id"`
	Slug           string    `json:"slug" db:"slug"`
	Name           string    `json:"name" db:"name"`
	Tagline        string    `json:"tagline" db:"tagline"`
	Description    string    `json:"description" db:"description"`
	Language       string    `json:"language" db:"language"`
	License        string    `json:"license" db:"license"`
	DatePublished  string    `json:"date_published" db:"date_published"`
	CodeRepository string    `json:"code_repository" db:"code_repository"`
	ToolOfTheWeek  bool      `json:"tool_of_the_week" db:"tool_of_the_week"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	Installed      bool      `json:"installed" db:"-"`
}

// InstallInstruction represents a single install command for a platform.
type InstallInstruction struct {
	ID        string    `json:"id" db:"id"`
	ToolID    string    `json:"tool_id" db:"tool_id"`
	Platform  string    `json:"platform" db:"platform"`
	Command   string    `json:"command" db:"command"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SearchResult wraps a Tool with its installation map.
type SearchResult struct {
	Tool
	Installations map[string]string `json:"-"`
}

// SearchOptions controls search query parameters.
type SearchOptions struct {
	Query     string
	Limit     int
	SortField string
	SortOrder string
	Filter    *Filter
}

// FilterType enumerates the kinds of filter tree nodes.
type FilterType int

const (
	// FilterAnd is a logical AND node.
	FilterAnd FilterType = iota
	// FilterOr is a logical OR node.
	FilterOr
	// FilterNot is a logical NOT node.
	FilterNot
	// FilterField is a field=value leaf node.
	FilterField
)

// Filter is a node in a filter expression tree.
type Filter struct {
	Type  FilterType
	Field string
	Value string
	Left  *Filter
	Right *Filter
}

// Tag represents a named tag.
type Tag struct {
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// TagCount pairs a tag name with its usage count.
type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
