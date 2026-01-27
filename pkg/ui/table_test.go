package ui

import (
	"testing"
)

func TestGetGradientColorSimple(t *testing.T) {
	testCases := []struct {
		index    int
		expected string
	}{
		{0, "#90EE90"},
		{1, "#8BE88C"},
		{29, "#004088"},
		{30, "#90EE90"},
		{31, "#8BE88C"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := GetGradientColorSimple(tc.index)
			if result != tc.expected {
				t.Errorf("GetGradientColorSimple(%d) = %s, want %s", tc.index, result, tc.expected)
			}
		})
	}
}

func TestGetGradientColor(t *testing.T) {
	testCases := []struct {
		pos      int
		total    int
		expected string
	}{
		{0, 100, "#90EE90"},
		{50, 100, "#4A9A88"},
		{100, 100, "#004088"},
		{0, 0, "#90EE90"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			result := GetGradientColor(tc.pos, tc.total)
			if result != tc.expected {
				t.Errorf("GetGradientColor(%d, %d) = %s, want %s", tc.pos, tc.total, result, tc.expected)
			}
		})
	}
}

func TestRenderTableGolden(t *testing.T) {
	config := TableConfig{
		Headers:    []string{"#", "Name", "Tagline", "Language"},
		Rows:       [][]string{{"1", "test-tool", "A test tool for testing", "Go"}},
		ShowHeader: true,
	}

	output := RenderTable(config)

	expected := `┌───┬───────────┬─────────────────────────┬──────────┐
│ # │ Name      │ Tagline                 │ Language │
├───┼───────────┼─────────────────────────┼──────────┤
│ 1 │ test-tool │ A test tool for testing │ Go       │
└───┴───────────┴─────────────────────────┴──────────┘`

	if output != expected {
		t.Errorf("RenderTable output does not match golden file")
		t.Logf("Got:\n%s", output)
		t.Logf("Expected:\n%s", expected)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	config := TableConfig{
		Headers:    []string{"#", "Name"},
		Rows:       [][]string{},
		ShowHeader: true,
	}

	output := RenderTable(config)

	if output != "" {
		t.Errorf("Expected empty output for empty rows, got: %s", output)
	}
}

func TestRenderTableNoHeader(t *testing.T) {
	config := TableConfig{
		Rows:       [][]string{{"1", "test-tool"}},
		ShowHeader: false,
	}

	output := RenderTable(config)

	if output == "" {
		t.Error("Expected non-empty output for table without header")
	}

	if len(output) < 10 {
		t.Errorf("Expected at least 10 chars, got: %d", len(output))
	}
}

func TestRenderKeyValueTableGolden(t *testing.T) {
	rows := [][]string{
		{"Name", "Test Tool"},
		{"Tagline", "A test tool"},
		{"Language", "Go"},
		{"License", "MIT"},
	}

	output := RenderKeyValueTable(rows)

	expected := `┌──────────┬─────────────┐
│ Name     │ Test Tool   │
│ Tagline  │ A test tool │
│ Language │ Go          │
│ License  │ MIT         │
└──────────┴─────────────┘`

	if output != expected {
		t.Errorf("RenderKeyValueTable output does not match golden file")
		t.Logf("Got:\n%s", output)
		t.Logf("Expected:\n%s", expected)
	}
}

func TestRenderKeyValueTableEmpty(t *testing.T) {
	rows := [][]string{}

	output := RenderKeyValueTable(rows)

	if output != "" {
		t.Errorf("Expected empty output for empty rows, got: %s", output)
	}
}

func TestRenderTableCustomStyle(t *testing.T) {
	config := TableConfig{
		Headers: []string{"#", "Name"},
		Rows:    [][]string{{"1", "test"}},
		RowFunc: func(s string, rowIdx, colIdx int) string {
			return s
		},
		ShowHeader: true,
	}

	output := RenderTable(config)

	if output == "" {
		t.Error("Expected non-empty output with custom style")
	}

	if len(output) < 5 {
		t.Errorf("Expected at least 5 chars, got: %d", len(output))
	}
}
