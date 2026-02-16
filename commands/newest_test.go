package commands

import (
	"testing"

	"troveler/db"
)

func TestFilterNewestResults(t *testing.T) {
	tests := []struct {
		name     string
		results  []db.SearchResult
		onlyTotw bool
		langs    []string
		expected int
	}{
		{
			name: "No filters returns all",
			results: []db.SearchResult{
				{Tool: db.Tool{Name: "tool1"}},
				{Tool: db.Tool{Name: "tool2"}},
			},
			onlyTotw: false,
			langs:    []string{},
			expected: 2,
		},
		{
			name: "Only TOTW filters correctly",
			results: []db.SearchResult{
				{Tool: db.Tool{Name: "tool1", ToolOfTheWeek: true}},
				{Tool: db.Tool{Name: "tool2", ToolOfTheWeek: false}},
			},
			onlyTotw: true,
			langs:    []string{},
			expected: 1,
		},
		{
			name: "Language filter (single)",
			results: []db.SearchResult{
				{Tool: db.Tool{Name: "tool1", Language: "go"}},
				{Tool: db.Tool{Name: "tool2", Language: "rust"}},
			},
			onlyTotw: false,
			langs:    []string{"go"},
			expected: 1,
		},
		{
			name: "Language filter (multiple, case-insensitive)",
			results: []db.SearchResult{
				{Tool: db.Tool{Name: "tool1", Language: "Go"}},
				{Tool: db.Tool{Name: "tool2", Language: "RUST"}},
				{Tool: db.Tool{Name: "tool3", Language: "python"}},
			},
			onlyTotw: false,
			langs:    []string{"go", "rust"},
			expected: 2,
		},
		{
			name: "Combined filters",
			results: []db.SearchResult{
				{Tool: db.Tool{Name: "tool1", Language: "go", ToolOfTheWeek: true}},
				{Tool: db.Tool{Name: "tool2", Language: "go", ToolOfTheWeek: false}},
				{Tool: db.Tool{Name: "tool3", Language: "rust", ToolOfTheWeek: true}},
			},
			onlyTotw: true,
			langs:    []string{"go"},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterNewestResults(tt.results, tt.onlyTotw, tt.langs)
			if len(filtered) != tt.expected {
				t.Errorf("filterNewestResults() got %d results, expected %d", len(filtered), tt.expected)
			}
		})
	}
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Empty string returns unknown",
			input:    "",
			expected: "unknown",
		},
		{
			name:     "Invalid date returns input",
			input:    "not a date",
			expected: "not a date",
		},
		{
			name:     "Valid RFC3339 date formats correctly",
			input:    "2024-03-12T00:00:00.000Z",
			expected: "2024-03-12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatDate(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
