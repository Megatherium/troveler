package db

import "time"

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
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	Installed      bool      `json:"installed" db:"-"`
}

type InstallInstruction struct {
	ID        string    `json:"id" db:"id"`
	ToolID    string    `json:"tool_id" db:"tool_id"`
	Platform  string    `json:"platform" db:"platform"`
	Command   string    `json:"command" db:"command"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type SearchResult struct {
	Tool
	Installations map[string]string `json:"-"`
}

type SearchOptions struct {
	Query     string
	Limit     int
	SortField string
	SortOrder string  // ASC or DESC
	Filter    *Filter // Optional filter AST
}

// FilterType represents the type of filter
type FilterType int

const (
	FilterAnd FilterType = iota
	FilterOr
	FilterNot
	FilterField
)

// Filter represents a parsed filter expression
type Filter struct {
	Type  FilterType
	Field string
	Value string
	Left  *Filter
	Right *Filter
}

type Tag struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type TagCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
