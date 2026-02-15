package commands

import (
	"context"
	"os"
	"testing"

	"troveler/config"
	"troveler/db"
	"troveler/internal/platform"
)

func TestMiseModeOverridePriority(t *testing.T) {
	configPath := t.TempDir() + "/config.toml"
	configContent := `[install]
mise_mode = true`
	//nolint:gosec // G306: test file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	tests := []struct {
		name          string
		miseFlag      bool
		cliOverride   string
		expectedMatch string
		description   string
	}{
		{
			name:          "mise_config_only_no_cli_override",
			miseFlag:      false,
			cliOverride:   "",
			expectedMatch: "go",
			description:   "With mise_mode=true in config and no CLI override, should use LANG",
		},
		{
			name:          "mise_config_with_cli_override_platform",
			miseFlag:      false,
			cliOverride:   "cargo",
			expectedMatch: "rust (cargo)",
			description:   "With mise_mode=true and CLI override 'cargo', CLI override should win",
		},
		{
			name:          "mise_config_with_cli_override_github",
			miseFlag:      false,
			cliOverride:   "github",
			expectedMatch: "github",
			description:   "With mise_mode=true and CLI override 'github', CLI override should win",
		},
		{
			name:          "mise_flag_with_cli_override_github",
			miseFlag:      true,
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

			miseEnabled := tt.miseFlag || cfg.Install.MiseMode

			var matched []db.InstallInstruction
			override := tt.cliOverride

			if miseEnabled && override == "" {
				override = "LANG"
			}

			if override != "" {
				if override == "LANG" {
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

func extractPlatforms(installs []db.InstallInstruction) []string {
	platforms := make([]string, len(installs))
	for i, inst := range installs {
		platforms[i] = inst.Platform
	}

	return platforms
}
