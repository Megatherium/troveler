package commands

import (
	"context"
	"testing"

	"troveler/db"
	"troveler/internal/platform"
)

func TestMiseLangOverridePriority(t *testing.T) {
	tests := []struct {
		name          string
		cliOverride   string
		expectedMatch string
		description   string
	}{
		{
			name:          "mise_lang_no_cli_override",
			cliOverride:   "",
			expectedMatch: "go",
			description:   "With mise_lang as override and no CLI override, should use language matching",
		},
		{
			name:          "mise_lang_with_cli_override_platform",
			cliOverride:   "cargo",
			expectedMatch: "rust (cargo)",
			description:   "With mise_lang config and CLI override 'cargo', CLI override should win",
		},
		{
			name:          "mise_lang_with_cli_override_github",
			cliOverride:   "github",
			expectedMatch: "github",
			description:   "With mise_lang config and CLI override 'github', CLI override should win",
		},
		{
			name:          "mise_flag_with_cli_override_github",
			cliOverride:   "github",
			expectedMatch: "github",
			description:   "--mise flag with CLI override 'github' should respect CLI override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbPath := t.TempDir() + "/test.db"
			database, err := db.New(dbPath)
			if err != nil {
				t.Fatalf("Failed to create database: %v", err)
			}
			defer func() { _ = database.Close() }()

			ctx := context.Background()
			tool := &db.Tool{
				ID:       "test-tool-id",
				Slug:     "test-tool",
				Name:     "Test Tool",
				Language: "go",
			}
			if err := database.UpsertTool(ctx, tool); err != nil {
				t.Fatalf("Failed to insert tool: %v", err)
			}

			installs := []db.InstallInstruction{
				{ID: "1", ToolID: tool.ID, Platform: "go", Command: "go install github.com/user/repo"},
				{ID: "2", ToolID: tool.ID, Platform: "rust (cargo)", Command: "cargo install some-crate"},
				{ID: "3", ToolID: tool.ID, Platform: "github", Command: "eget user/repo"},
			}
			for _, inst := range installs {
				if err := database.UpsertInstallInstruction(ctx, &inst); err != nil {
					t.Fatalf("Failed to insert install instruction: %v", err)
				}
			}

			var matched []db.InstallInstruction
			override := tt.cliOverride

			// Simulate --mise flag behavior: if no CLI override, use mise_lang
			if override == "" {
				override = PlatformMiseLang
			}

			if platform.IsLangPlatform(override) {
				for _, inst := range installs {
					if platform.MatchLanguage(tool.Language, inst.Platform) {
						matched = append(matched, inst)
					}
				}
			} else {
				platformID := platform.Normalize(override)
				for _, inst := range installs {
					if platform.MatchPlatform(platformID, inst.Platform) {
						matched = append(matched, inst)
					}
				}
			}

			if len(matched) == 0 {
				t.Errorf("%s: Expected to find at least one match, got none", tt.description)

				return
			}

			found := false
			for _, m := range matched {
				if m.Platform == tt.expectedMatch {
					found = true

					break
				}
			}
			if !found {
				t.Errorf("%s: Expected to find platform %q in matches, got %v",
					tt.description, tt.expectedMatch, extractPlatforms(matched))
			}
		})
	}
}

func TestLangOverrideMatching(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	database, err := db.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer func() { _ = database.Close() }()

	ctx := context.Background()
	tool := &db.Tool{
		ID:       "test-tool-id",
		Slug:     "test-tool",
		Name:     "Test Tool",
		Language: "go",
	}
	if err := database.UpsertTool(ctx, tool); err != nil {
		t.Fatalf("Failed to insert tool: %v", err)
	}

	installs := []db.InstallInstruction{
		{ID: "1", ToolID: tool.ID, Platform: "go", Command: "go install github.com/user/repo"},
		{ID: "2", ToolID: tool.ID, Platform: "rust (cargo)", Command: "cargo install some-crate"},
	}
	for _, inst := range installs {
		if err := database.UpsertInstallInstruction(ctx, &inst); err != nil {
			t.Fatalf("Failed to insert install instruction: %v", err)
		}
	}

	// Test that "lang" triggers language matching (same as "mise_lang" for matching purposes)
	matched, usedFallback := platform.FilterDBInstalls(installs, "lang", "go")
	if usedFallback {
		t.Error("Expected normal match for lang platform")
	}
	if len(matched) != 1 || matched[0].Platform != "go" {
		t.Errorf("Expected to match 'go' platform, got %v", extractPlatforms(matched))
	}
}

func extractPlatforms(installs []db.InstallInstruction) []string {
	platforms := make([]string, len(installs))
	for i, inst := range installs {
		platforms[i] = inst.Platform
	}

	return platforms
}
