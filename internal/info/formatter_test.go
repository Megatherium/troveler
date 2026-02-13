package info

import (
	"strings"
	"testing"

	"troveler/db"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected int // expected number of lines
	}{
		{
			name:     "short_text_no_wrap",
			input:    "Hello world",
			width:    50,
			expected: 1,
		},
		{
			name:     "long_text_wraps",
			input:    "This is a very long text that should wrap to multiple lines when rendered",
			width:    20,
			expected: 4,
		},
		{
			name:     "exact_width_no_wrap",
			input:    "Exactly twenty chars",
			width:    20,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapText(tt.input, tt.width)
			lines := strings.Split(result, "\n")

			if len(lines) != tt.expected {
				t.Errorf("Expected %d lines, got %d\nResult: %q", tt.expected, len(lines), result)
			}
		})
	}
}

func TestFormatTool(t *testing.T) {
	tool := &db.Tool{
		Name:           "Test Tool",
		Slug:           "test-tool",
		Tagline:        "A test tool",
		Description:    "This is a test description",
		Language:       "go",
		License:        "MIT",
		CodeRepository: "https://github.com/test/tool",
		DatePublished:  "2024-01-01",
	}

	installs := []db.InstallInstruction{
		{Platform: "brew", Command: "brew install test-tool"},
		{Platform: "cargo", Command: "cargo install test-tool"},
	}

	info := FormatTool(tool, installs)

	if info.Name != "Test Tool" {
		t.Errorf("Expected name 'Test Tool', got %s", info.Name)
	}

	if len(info.InstallOptions) != 2 {
		t.Errorf("Expected 2 install options, got %d", len(info.InstallOptions))
	}

	if info.InstallOptions[0].Platform != "brew" {
		t.Errorf("Expected first platform 'brew', got %s", info.InstallOptions[0].Platform)
	}
}

func TestGetKeyValuePairs(t *testing.T) {
	info := &ToolInfo{
		Name:          "Test",
		Language:      "go",
		License:       "MIT",
		Repository:    "https://github.com/test",
		DatePublished: "2024-01-01",
		Slug:          "test",
	}

	pairs := info.GetKeyValuePairs()

	if len(pairs) < 5 {
		t.Errorf("Expected at least 5 pairs, got %d", len(pairs))
	}

	// Check that slug is always included
	found := false
	for _, pair := range pairs {
		if pair[0] == "Slug" && pair[1] == "test" {
			found = true

			break
		}
	}

	if !found {
		t.Error("Slug should always be in key-value pairs")
	}
}

func TestRenderPlainText(t *testing.T) {
	info := &ToolInfo{
		Name:        "Test Tool",
		Tagline:     "A test",
		Description: "Test description",
		Language:    "go",
		Slug:        "test",
	}

	output := info.RenderPlainText()

	if !strings.Contains(output, "Test Tool") {
		t.Error("Output should contain tool name")
	}

	if !strings.Contains(output, "A test") {
		t.Error("Output should contain tagline")
	}

	if !strings.Contains(output, "Language:") {
		t.Error("Output should contain language field")
	}
}
