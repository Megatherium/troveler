package platform

import (
	"testing"

	"troveler/db"
)

func TestMatchPlatform(t *testing.T) {
	tests := []struct {
		name            string
		detectedID      string
		installPlatform string
		want            bool
	}{
		{
			name:            "fedora matches linux:fedora",
			detectedID:      "fedora",
			installPlatform: "linux:fedora",
			want:            true,
		},
		{
			name:            "fedora does not match linux:arch",
			detectedID:      "fedora",
			installPlatform: "linux:arch",
			want:            false,
		},
		{
			name:            "pure go matches",
			detectedID:      "go",
			installPlatform: "go",
			want:            true,
		},
		{
			name:            "pure rust doesn't match go",
			detectedID:      "rust",
			installPlatform: "go",
			want:            false,
		},
		{
			name:            "method with distro aliases",
			detectedID:      "ubuntu",
			installPlatform: "linux:ubuntu / debian",
			want:            true,
		},
		{
			name:            "macos matches",
			detectedID:      "macos",
			installPlatform: "macos:brew",
			want:            true,
		},
		{
			name:            "freebsd matches",
			detectedID:      "freebsd",
			installPlatform: "bsd:freebsd",
			want:            true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPlatform(tt.detectedID, tt.installPlatform)
			if got != tt.want {
				t.Errorf("MatchPlatform(%q, %q) = %v, want %v", tt.detectedID, tt.installPlatform, got, tt.want)
			}
		})
	}
}

func TestNormalizeOSInfo(t *testing.T) {
	tests := []struct {
		name   string
		input  *OSInfo
		wantID string
	}{
		{
			name:   "ubuntu variants",
			input:  &OSInfo{ID: "linuxmint"},
			wantID: OSUbuntu,
		},
		{
			name:   "rhel variants",
			input:  &OSInfo{ID: "centos"},
			wantID: OSRHEL,
		},
		{
			name:   "arch variants",
			input:  &OSInfo{ID: "manjaro"},
			wantID: OSArch,
		},
		{
			name:   "fedora stays fedora",
			input:  &OSInfo{ID: "fedora"},
			wantID: OSFedora,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeOSInfo(tt.input)
			if got.ID != tt.wantID {
				t.Errorf("normalizeOSInfo(%+v).ID = %v, want %v", tt.input, got.ID, tt.wantID)
			}
		})
	}
}

func TestSelectorPriority(t *testing.T) {
	tests := []struct {
		name           string
		cliOverride    string
		configOverride string
		fallback       string
		detectedOS     string
		expected       string
	}{
		{
			name:           "cli_override_highest_priority",
			cliOverride:    "macos",
			configOverride: "linux",
			fallback:       "windows",
			detectedOS:     "freebsd",
			expected:       "macos",
		},
		{
			name:           "config_override_beats_os_and_fallback",
			cliOverride:    "",
			configOverride: "linux",
			fallback:       "windows",
			detectedOS:     "macos",
			expected:       "linux",
		},
		{
			name:           "os_detection_beats_fallback",
			cliOverride:    "",
			configOverride: "",
			fallback:       "windows",
			detectedOS:     "macos",
			expected:       "macos",
		},
		{
			name:           "fallback_used_when_no_os",
			cliOverride:    "",
			configOverride: "",
			fallback:       "windows",
			detectedOS:     "",
			expected:       "windows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSelector(tt.cliOverride, tt.configOverride, tt.fallback, "go")
			result := s.Select(tt.detectedOS)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFilterInstallsLANG(t *testing.T) {
	installs := []db.InstallInstruction{
		{Platform: "go", Command: "go install"},
		{Platform: "go (cargo)", Command: "cargo install"},
		{Platform: "rust", Command: "cargo install"},
		{Platform: "python (pip)", Command: "pip install"},
	}

	matched, usedFallback := FilterDBInstalls(installs, "LANG", "go")

	if usedFallback {
		t.Error("Expected normal match, got fallback")
	}

	if len(matched) != 2 {
		t.Errorf("Expected 2 matches for go language, got %d", len(matched))
	}

	for _, m := range matched {
		if m.Platform != "go" && m.Platform != "go (cargo)" {
			t.Errorf("Unexpected platform match: %s", m.Platform)
		}
	}
}

func TestResolveVirtual(t *testing.T) {
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
			result := ResolveVirtual(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveVirtual(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSelectDefault(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "brew", Command: "brew install"},
		{ID: "2", Platform: "cargo", Command: "cargo install"},
	}

	defaultCmd := SelectDefaultDBInstalls(installs, false, "ubuntu")

	if defaultCmd == nil {
		t.Fatal("Expected default command, got nil")
	}

	if defaultCmd.ID != "1" {
		t.Errorf("Expected first command (ID=1), got ID=%s", defaultCmd.ID)
	}
}

func TestGenerateVirtualInstalls(t *testing.T) {
	tests := []struct {
		name              string
		input             []db.InstallInstruction
		expectedCount     int
		expectedPlatforms []string
	}{
		{
			name: "go_install_generates_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "go install github.com/user/repo@latest"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:go"},
		},
		{
			name: "cargo_install_generates_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "cargo install some-crate"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:cargo"},
		},
		{
			name: "npm_install_generates_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "npm install -g package-name"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:npm"},
		},
		{
			name: "yarn_generates_npm_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "yarn global add package-name"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:npm"},
		},
		{
			name: "pip_install_generates_pipx_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "pip install package-name"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:pipx"},
		},
		{
			name: "eget_generates_github_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "eget foo/bar"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:github"},
		},
		{
			name: "multiple_backends_all_generate_virtuals",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "go install github.com/user/repo"},
				{ID: "2", Platform: "linux", Command: "cargo install some-crate"},
				{ID: "3", Platform: "linux", Command: "npm install -g package"},
				{ID: "4", Platform: "linux", Command: "pip install py-package"},
				{ID: "5", Platform: "linux", Command: "eget user/repo"},
			},
			expectedCount:     5,
			expectedPlatforms: []string{"mise:go", "mise:cargo", "mise:npm", "mise:pipx", "mise:github"},
		},
		{
			name: "npm_variants_generate_single_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "npm install -g package"},
				{ID: "2", Platform: "linux", Command: "yarn global add package"},
				{ID: "3", Platform: "linux", Command: "pnpm add -g package"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:npm"},
		},
		{
			name: "brew_install_no_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "brew install package"},
			},
			expectedCount:     0,
			expectedPlatforms: []string{},
		},
		{
			name: "apt_install_no_virtual",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "linux", Command: "apt install package"},
			},
			expectedCount:     0,
			expectedPlatforms: []string{},
		},
		{
			name: "existing_mise_platform_skipped",
			input: []db.InstallInstruction{
				{ID: "1", Platform: "mise:go", Command: "mise use --global go:github.com/user/repo"},
				{ID: "2", Platform: "linux", Command: "go install github.com/user/repo"},
			},
			expectedCount:     1,
			expectedPlatforms: []string{"mise:go"},
		},
		{
			name:              "empty_input",
			input:             []db.InstallInstruction{},
			expectedCount:     0,
			expectedPlatforms: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateVirtualInstalls(tt.input)

			if len(result) != tt.expectedCount {
				t.Errorf("Expected %d virtuals, got %d", tt.expectedCount, len(result))
			}

			if len(result) > 0 {
				actualPlatforms := make([]string, len(result))
				for i, v := range result {
					actualPlatforms[i] = v.Platform
				}

				for i, expected := range tt.expectedPlatforms {
					if i >= len(actualPlatforms) {
						t.Errorf("Missing expected platform %s at index %d", expected, i)

						continue
					}
					if actualPlatforms[i] != expected {
						t.Errorf("Platform at index %d: expected %s, got %s", i, expected, actualPlatforms[i])
					}
				}
			}
		})
	}
}

func TestExtractBackend(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Backend
	}{
		{
			name:     "go_backend",
			input:    "mise use --global go:github.com/user/repo",
			expected: BackendGo,
		},
		{
			name:     "cargo_backend",
			input:    "mise use --global cargo:crate-name",
			expected: BackendCargo,
		},
		{
			name:     "npm_backend",
			input:    "mise use --global npm:package-name",
			expected: BackendNPM,
		},
		{
			name:     "pipx_backend",
			input:    "mise use --global pipx:package-name",
			expected: BackendPipx,
		},
		{
			name:     "github_backend",
			input:    "mise use --global github:user/repo",
			expected: BackendGithub,
		},
		{
			name:     "invalid_backend",
			input:    "mise use --global invalid:package",
			expected: "",
		},
		{
			name:     "no_backend",
			input:    "mise use --global package",
			expected: "",
		},
		{
			name:     "not_mise_command",
			input:    "brew install package",
			expected: "",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractBackend(tt.input)
			if result != tt.expected {
				t.Errorf("extractBackend(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTransformToMise(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "go_install",
			input:    "go install github.com/user/repo",
			expected: "mise use --global go:github.com/user/repo",
		},
		{
			name:     "cargo_install",
			input:    "cargo install some-crate",
			expected: "mise use --global cargo:some-crate",
		},
		{
			name:     "npm_install_global",
			input:    "npm install -g package",
			expected: "mise use --global npm:package",
		},
		{
			name:     "pip_install",
			input:    "pip install package",
			expected: "mise use --global pipx:package",
		},
		{
			name:     "eget",
			input:    "eget user/repo",
			expected: "mise use --global github:user/repo",
		},
		{
			name:     "brew_no_transform",
			input:    "brew install package",
			expected: "brew install package",
		},
		{
			name:     "apt_no_transform",
			input:    "apt install package",
			expected: "apt install package",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TransformToMise(tt.input)
			if result != tt.expected {
				t.Errorf("TransformToMise(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
