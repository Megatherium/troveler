package db

import (
	"testing"
)

func TestNormalizeTagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		hasError bool
	}{
		{"cli", "cli", false},
		{"  cli  ", "cli", false},
		{"CLI", "cli", false},
		{"  CLI  ", "cli", false},
		{"", "", true},
		{"   ", "", true},
		{"\t\n", "", true},
		{"FuZzY", "fuzzy", false},
	}

	for _, tt := range tests {
		result, err := normalizeTagName(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("normalizeTagName(%q) expected error, got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("normalizeTagName(%q) unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("normalizeTagName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		}
	}
}

func TestAddTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	err := database.AddTag("fzf", "")
	if err == nil {
		t.Error("AddTag with empty name should fail")
	}

	err = database.AddTag("fzf", "   ")
	if err == nil {
		t.Error("AddTag with whitespace-only name should fail")
	}
}

func TestAddTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	if err := database.AddTag("fzf", "  CLI  "); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 1 {
		t.Errorf("Expected 1 normalized tag, got %d: %v", len(tags), tags)
	}
	if len(tags) > 0 && tags[0] != "cli" {
		t.Errorf("Expected normalized tag 'cli', got %q", tags[0])
	}
}

func TestRemoveTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	err := database.RemoveTag("fzf", "  CLI  ")
	if err != nil {
		t.Errorf("RemoveTag with different case/whitespace should work: %v", err)
	}

	tags, _ := database.GetTags("fzf")
	if len(tags) != 0 {
		t.Errorf("Expected 0 tags after removal, got %d", len(tags))
	}
}

func TestRemoveTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")

	err := database.RemoveTag("fzf", "")
	if err == nil {
		t.Error("RemoveTag with empty name should fail")
	}
}

func TestGetToolsByTagNormalizesInput(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	seedTool(t, database, "fzf", "fzf")
	if err := database.AddTag("fzf", "cli"); err != nil {
		t.Fatalf("AddTag failed: %v", err)
	}

	tools, err := database.GetToolsByTag("  CLI  ")
	if err != nil {
		t.Errorf("GetToolsByTag with different case/whitespace should work: %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestGetToolsByTagEmptyName(t *testing.T) {
	database := setupTestDB(t)
	defer checkClose(t, database)

	_, err := database.GetToolsByTag("")
	if err == nil {
		t.Error("GetToolsByTag with empty name should fail")
	}
}
