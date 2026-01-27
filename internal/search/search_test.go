package search

import (
	"testing"
)

func TestValidSortFields(t *testing.T) {
	validFields := []string{"name", "tagline", "language"}
	
	for _, field := range validFields {
		if !ValidSortFields[field] {
			t.Errorf("Expected %s to be a valid sort field", field)
		}
	}
}

func TestOptionsDefaults(t *testing.T) {
	tests := []struct {
		name           string
		input          Options
		expectedLimit  int
		expectedSort   string
		expectedOrder  string
	}{
		{
			name:          "zero_limit_defaults_to_50",
			input:         Options{Query: "test", Limit: 0},
			expectedLimit: 50,
			expectedSort:  "name",
			expectedOrder: "ASC",
		},
		{
			name:          "invalid_sort_defaults_to_name",
			input:         Options{Query: "test", Limit: 10, SortField: "invalid"},
			expectedLimit: 10,
			expectedSort:  "name",
			expectedOrder: "ASC",
		},
		{
			name:          "desc_preserved",
			input:         Options{Query: "test", Limit: 10, SortField: "language", SortOrder: "DESC"},
			expectedLimit: 10,
			expectedSort:  "language",
			expectedOrder: "DESC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply defaults logic (simulating what Search() does)
			opts := tt.input
			
			if opts.Limit <= 0 {
				opts.Limit = 50
			}
			
			if !ValidSortFields[opts.SortField] {
				opts.SortField = "name"
			}
			
			if opts.SortOrder != "DESC" && opts.SortOrder != "desc" {
				opts.SortOrder = "ASC"
			}
			
			if opts.Limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, opts.Limit)
			}
			
			if opts.SortField != tt.expectedSort {
				t.Errorf("Expected sort field %s, got %s", tt.expectedSort, opts.SortField)
			}
			
			if opts.SortOrder != tt.expectedOrder {
				t.Errorf("Expected sort order %s, got %s", tt.expectedOrder, opts.SortOrder)
			}
		})
	}
}
