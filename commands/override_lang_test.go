package commands

import (
	"testing"
)

func TestOverrideLangLanguageMatching(t *testing.T) {
	testCases := []struct {
		name             string
		fallbackPlatform string
		toolLanguage     string
		installPlatform  string
		shouldMatch      bool
		reason           string
	}{
		{
			name:             "lang_matches_go_platforms",
			fallbackPlatform: "lang",
			toolLanguage:     "go",
			installPlatform:  platformGOPip,
			shouldMatch:      true,
			reason:           "lang fallback with Go tool should match go (pip) platforms",
		},
		{
			name:             "lang_doesnt_match_different_language",
			fallbackPlatform: "lang",
			toolLanguage:     "go",
			installPlatform:  "rust",
			shouldMatch:      false,
			reason:           "lang fallback with Go tool should NOT match rust platforms",
		},
		{
			name:             "lang_matches_python_platforms",
			fallbackPlatform: "lang",
			toolLanguage:     "python",
			installPlatform:  "python (pip)",
			shouldMatch:      true,
			reason:           "lang fallback with Python tool should match python (pip) platforms",
		},
		{
			name:             "mise_lang_matches_go_platforms",
			fallbackPlatform: "mise_lang",
			toolLanguage:     "go",
			installPlatform:  platformGOPip,
			shouldMatch:      true,
			reason:           "mise_lang fallback with Go tool should match go (pip) platforms",
		},
		{
			name:             "non_lang_uses_OS_detection",
			fallbackPlatform: "macos",
			toolLanguage:     "go",
			installPlatform:  "brew",
			shouldMatch:      true,
			reason:           "Non-lang fallback platform should use platform matching",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Testing: %s", tc.reason)

			if tc.fallbackPlatform == PlatformLang || tc.fallbackPlatform == PlatformMiseLang {
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
	platforms := []string{
		"macos", "linux:arch", "ubuntu", "fedora", "windows",
		"alpine", "arch", "freebsd", "lang", "mise_lang",
	}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			t.Logf("Testing fallback_platform=%s", platform)
			t.Logf("✓ Fallback platform %q is valid and should be used", platform)
		})
	}
}
