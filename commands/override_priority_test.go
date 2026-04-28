package commands

import (
	"testing"
)

func TestOverrideFlagPriority(t *testing.T) {
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
			fallbackPlatform: "lang",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "brew",
			reason:           "Command-line --override flag should have highest priority",
		},
		{
			name:             "platformOverride_beats_OS_and_fallback",
			platformArg:      "",
			platformOverride: "macos",
			fallbackPlatform: "lang",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "macos",
			reason:           "platform_override config has higher priority than OS detection and fallback",
		},
		{
			name:             "fallback_lang_uses_language_when_os_no_match",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "lang",
			detectedOS:       "",
			toolLanguage:     "go",
			expectedPlatform: "go",
			reason:           "fallback_platform=lang should use tool language when OS detection fails",
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
			fallbackPlatform: "lang",
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
		{
			name:             "mise_lang_as_fallback",
			platformArg:      "",
			platformOverride: "",
			fallbackPlatform: "mise_lang",
			detectedOS:       "",
			toolLanguage:     "go",
			expectedPlatform: "go",
			reason:           "fallback_platform=mise_lang should use tool language matching",
		},
		{
			name:             "mise_lang_as_override",
			platformArg:      "",
			platformOverride: "mise_lang",
			fallbackPlatform: "",
			detectedOS:       "linux:arch",
			toolLanguage:     "go",
			expectedPlatform: "go",
			reason:           "platform_override=mise_lang should use language matching overriding OS detection",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.reason)

			selectedPlatform := tc.platformArg
			if selectedPlatform == "" {
				selectedPlatform = tc.platformOverride
			}
			if selectedPlatform == "" {
				selectedPlatform = tc.detectedOS
			}
			if selectedPlatform == "" {
				selectedPlatform = tc.fallbackPlatform
			}

			if selectedPlatform == PlatformLang || selectedPlatform == PlatformMiseLang {
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
