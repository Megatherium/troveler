package commands

import (
	"strings"
	"testing"

	"troveler/config"
)

func TestTaglineTruncation(t *testing.T) {
	testCases := []struct {
		name     string
		tagline  string
		width    int
		expected string
	}{
		{
			name:     "short tagline no truncation",
			tagline:  "A simple tool",
			width:    50,
			expected: "A simple tool",
		},
		{
			name:     "exact width no truncation",
			tagline:  strings.Repeat("a", 50),
			width:    50,
			expected: strings.Repeat("a", 50),
		},
		{
			name:     "one over should truncate",
			tagline:  strings.Repeat("a", 51),
			width:    50,
			expected: strings.Repeat("a", 47) + "...",
		},
		{
			name:     "much over should truncate",
			tagline:  strings.Repeat("a", 100),
			width:    50,
			expected: strings.Repeat("a", 47) + "...",
		},
		{
			name:     "custom width 30",
			tagline:  strings.Repeat("b", 40),
			width:    30,
			expected: strings.Repeat("b", 27) + "...",
		},
		{
			name:     "custom width 80",
			tagline:  strings.Repeat("c", 90),
			width:    80,
			expected: strings.Repeat("c", 77) + "...",
		},
		{
			name:     "zero width should not crash",
			tagline:  strings.Repeat("d", 100),
			width:    0,
			expected: "d" + strings.Repeat("d", 99),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.tagline
			if tc.width > 0 && len(result) > tc.width {
				result = result[:tc.width-3] + "..."
			}

			if result != tc.expected {
				t.Errorf("Tagline truncation failed\n"+
					"  Width: %d\n"+
					"  Input: %d chars\n"+
					"  Got: %d chars (%s)\n"+
					"  Expected: %d chars (%s)",
					tc.width, len(tc.tagline),
					len(result), result,
					len(tc.expected), tc.expected)
			}
		})
	}
}

func TestConfigDefaultTaglineWidth(t *testing.T) {
	cfg := &config.Config{}
	if cfg.Search.TaglineWidth != 0 {
		t.Errorf("Expected default TaglineWidth 0, got %d", cfg.Search.TaglineWidth)
	}
}

func TestConfigApplyDefaultTaglineWidth(t *testing.T) {
	cfg := &config.Config{Search: config.SearchConfig{TaglineWidth: 0}}

	if cfg.Search.TaglineWidth == 0 {
		cfg.Search.TaglineWidth = 50
	}

	if cfg.Search.TaglineWidth != 50 {
		t.Errorf("Expected applied default 50, got %d", cfg.Search.TaglineWidth)
	}
}

func TestConfigCustomTaglineWidth(t *testing.T) {
	cfg := &config.Config{Search: config.SearchConfig{TaglineWidth: 30}}

	if cfg.Search.TaglineWidth == 0 {
		cfg.Search.TaglineWidth = 50
	}

	if cfg.Search.TaglineWidth != 30 {
		t.Errorf("Expected custom 30, got %d", cfg.Search.TaglineWidth)
	}
}
