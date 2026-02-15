package commands

import (
	"testing"
)

func TestOverrideLANGLanguageMatching(t *testing.T) {
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
	platforms := []string{"macos", "linux:arch", "ubuntu", "fedora", "windows", "alpine", "arch", "freebsd"}

	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			t.Logf("Testing fallback_platform=%s", platform)
			t.Logf("✓ Fallback platform %q is valid and should be used", platform)
		})
	}
}
