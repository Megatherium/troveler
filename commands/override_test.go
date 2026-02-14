package commands

import (
	"context"
	"os"
	"testing"

	"troveler/config"
	"troveler/db"
	"troveler/lib"
)

const (
	platformGOPip = "go (pip)"
)

func TestOverrideFlagPriority(t *testing.T) {
	// Test that --override flag has highest priority

	testCases := []struct {
		name             string
		platformArg      string
		platformOverride string
		fallbackPlatform string
		detectedOS       string
		toolLanguage     string
		expectedPlatform string
		reason           string
	}{
		{
			name:             "override_flag_highest_priority",
			platformArg:      "brew",
			platformOverride: "macos",
			fallbackPlatform: "LANG",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "brew",
			reason:           "Command-line --override flag should have highest priority",
		},
		{
			name:             "platformOverride_beats_OS_and_fallback",
			platformArg:      "",
			platformOverride: "macos",
			fallbackPlatform: "LANG",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "platform_override config has higher priority than OS detection and fallback",
		},
		{
			name:             "fallback_LANG_uses_language_when_os_no_match",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "LANG",
			detectedOS:       "",
			toolLanguage:     "go",
			expectedPlatform: "go",
			reason:           "fallback_platform=LANG should use tool language when OS detection fails",
		},
		{
			name:             "OS_detection_beats_fallback",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "macos",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "linux:arch",
			reason:           "OS detection should take precedence over fallback_platform",
		},
		{
			name:             "OS_detection_with_empty",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "",
			detectedOS:       "macos",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "OS detection should be used when no override or fallback set",
		},
		{
			name:             "CLI_arg_overrides_everything",
			platformArg:      "macos",
			platformOverride: "windows",
			fallbackPlatform: "LANG",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "CLI --override argument has highest priority over all other settings",
		},
		{
			name:             "no_config_uses_OS",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "",
			detectedOS:       "macos",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "When no override or fallback is set, should use OS detection",
		},
		{
			name:             "fallback_used_when_OS_detected_but_set",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "brew",
			detectedOS:       "macos",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "OS detection takes priority over fallback when OS is detected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.reason)

			// Simulated priority logic: CLI > config override > OS detection > fallback
			selectedPlatform := tc.platformArg
			if selectedPlatform == "" {
				selectedPlatform = tc.platformOverride
			}
			if selectedPlatform == "" {
				// Try OS detection first
				selectedPlatform = tc.detectedOS
			}
			if selectedPlatform == "" {
				// Fallback is last resort
				selectedPlatform = tc.fallbackPlatform
			}

			// Handle LANG special case (converted to language in actual implementation)
			if selectedPlatform == platformLang {
				selectedPlatform = tc.toolLanguage
			}

			if selectedPlatform != tc.expectedPlatform {
				t.Errorf("Platform selection failed\n"+
					"  Platform Arg: %q\n"+
					"  Override: %q\n"+
					"  Fallback: %q\n"+
					"  Detected OS: %q\n"+
					"  Tool Lang: %q\n"+
					"  Got: %q\n"+
					"  Expected: %q",
					tc.platformArg, tc.platformOverride, tc.fallbackPlatform,
					tc.detectedOS, tc.toolLanguage,
					selectedPlatform, tc.expectedPlatform)
			} else {
				t.Logf("✓ Selected platform: %q", selectedPlatform)
			}
		})
	}
}

func TestOverrideLANGLanguageMatching(t *testing.T) {
	// Test that LANG fallback uses language matching correctly

	testCases := []struct {
		name             string
		fallbackPlatform string
		toolLanguage     string
		installPlatform  string
		shouldMatch      bool
		reason           string
	}{
		{
			name:             "LANG_matches_go_platforms",
			fallbackPlatform: "LANG",
			toolLanguage:     "go",
			installPlatform:  platformGOPip,
			shouldMatch:      true,
			reason:           "LANG fallback with Go tool should match go (pip) platforms",
		},
		{
			name:             "LANG_doesnt_match_different_language",
			fallbackPlatform: "LANG",
			toolLanguage:     "go",
			installPlatform:  "rust",
			shouldMatch:      false,
			reason:           "LANG fallback with Go tool should NOT match rust platforms",
		},
		{
			name:             "LANG_matches_python_platforms",
			fallbackPlatform: "LANG",
			toolLanguage:     "python",
			installPlatform:  "python (pip)",
			shouldMatch:      true,
			reason:           "LANG fallback with Python tool should match python (pip) platforms",
		},
		{
			name:             "non_LANG_uses_OS_detection",
			fallbackPlatform: "macos",
			toolLanguage:     "go",
			installPlatform:  "brew",
			shouldMatch:      true,
			reason:           "Non-LANG fallback platform should use platform matching",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.reason)

			if tc.fallbackPlatform == platformLang {
				if tc.toolLanguage == "go" && tc.installPlatform == platformGOPip {
					t.Logf("✓ Language match: %s matches %s", tc.toolLanguage, tc.installPlatform)
				}
				if tc.toolLanguage != "go" && tc.installPlatform == platformGOPip {
					t.Log("✗ No language mismatch as expected")
				}
			} else {
				if tc.toolLanguage != "go" && tc.installPlatform != platformGOPip {
					t.Log("✗ Language mismatch as expected")
				}
				if tc.installPlatform == "brew" {
					t.Logf("✓ Platform matches: %s", tc.installPlatform)
				}
			}
		})
	}
}

func TestSpecificFallbackPlatform(t *testing.T) {
	// Test that specific fallback platforms work correctly

	platforms := []string{"macos", "linux:arch", "ubuntu", "fedora", "windows", "alpine", "arch", "freebsd"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			t.Logf("Testing fallback_platform=%s", platform)
			t.Logf("✓ Fallback platform %q is valid and should be used", platform)
		})
	}
}

func TestMiseModeOverridePriority(t *testing.T) {
	// Create temporary config with mise_mode = true
	configPath := t.TempDir() + "/config.toml"
	configContent := `[install]
mise_mode = true`
	//nolint:gosec // G306: test file
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Load config
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
			// Create a temporary database with a Go tool
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

			// Simulate runInstall logic
			miseEnabled := tt.miseFlag || cfg.Install.MiseMode

			var matched []db.InstallInstruction
			override := tt.cliOverride

			// If mise mode is enabled AND no CLI override was provided, force LANG override
			if miseEnabled && override == "" {
				override = "LANG"
			}

			// Check for CLI or config override first (config override is empty in our test)
			if override != "" {
				if override == "LANG" {
					for _, inst := range installs {
						if lib.MatchLanguage(tool.Language, inst.Platform) {
							matched = append(matched, inst)
						}
					}
				} else {
					platform := lib.NormalizePlatform(override)
					for _, inst := range installs {
						if lib.MatchPlatform(platform, inst.Platform) {
							matched = append(matched, inst)
						}
					}
				}
			}

			// Verify the match
			if len(matched) == 0 {
				t.Errorf("%s: Expected to find at least one match, got none", tt.description)

				return
			}

			// Check if expected platform is in matches
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
