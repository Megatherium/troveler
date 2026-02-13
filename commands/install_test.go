package commands

import (
	"testing"

	"troveler/db"
)

func TestResolveVirtualPlatform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "mise_github_to_github",
			input:    "mise:github",
			expected: "github",
		},
		{
			name:     "mise_go_to_go",
			input:    "mise:go",
			expected: "go",
		},
		{
			name:     "mise_cargo_to_cargo",
			input:    "mise:cargo",
			expected: "cargo",
		},
		{
			name:     "mise_npm_to_npm",
			input:    "mise:npm",
			expected: "npm",
		},
		{
			name:     "mise_pipx_to_pipx",
			input:    "mise:pipx",
			expected: "pipx",
		},
		{
			name:     "regular_platform_unchanged",
			input:    "github",
			expected: "github",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "non_mise_prefix",
			input:    "docker:github",
			expected: "docker:github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveVirtualPlatform(tt.input)
			if result != tt.expected {
				t.Errorf("resolveVirtualPlatform(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFindMatchingInstallsWithVirtualPlatforms(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "github", Command: "eget user/repo"},
		{ID: "2", Platform: "go", Command: "go install github.com/user/repo"},
		{ID: "3", Platform: "cargo", Command: "cargo install some-crate"},
	}

	tests := []struct {
		name        string
		override    string
		expectCount int
		platform    string
	}{
		{
			name:        "virtual_mise_github",
			override:    "mise:github",
			expectCount: 1,
			platform:    "github",
		},
		{
			name:        "virtual_mise_go",
			override:    "mise:go",
			expectCount: 1,
			platform:    "go",
		},
		{
			name:        "virtual_mise_cargo",
			override:    "mise:cargo",
			expectCount: 1,
			platform:    "cargo",
		},
		{
			name:        "regular_github",
			override:    "github",
			expectCount: 1,
			platform:    "github",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			override := ResolveVirtualPlatform(tt.override)
			matched := FindMatchingInstalls(override, installs)

			if len(matched) != tt.expectCount {
				t.Errorf("Expected %d matches for override %s, got %d",
					tt.expectCount, tt.override, len(matched))
			}

			if tt.expectCount > 0 {
				if matched[0].Platform != tt.platform {
					t.Errorf("Expected platform %q, got %q",
						tt.platform, matched[0].Platform)
				}
			}
		})
	}
}
