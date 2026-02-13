package install

import (
	"testing"

	"troveler/db"
)

func TestPlatformSelectorPriority(t *testing.T) {
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
			ps := NewPlatformSelector(tt.cliOverride, tt.configOverride, tt.fallback, "go")
			result := ps.SelectPlatform(tt.detectedOS)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFilterCommandsLANG(t *testing.T) {
	installs := []db.InstallInstruction{
		{Platform: "go", Command: "go install"},
		{Platform: "go (cargo)", Command: "cargo install"},
		{Platform: "rust", Command: "cargo install"},
		{Platform: "python (pip)", Command: "pip install"},
	}

	matched, usedFallback := FilterCommands(installs, "LANG", "go")

	if usedFallback {
		t.Error("Expected normal match, got fallback")
	}

	if len(matched) != 2 {
		t.Errorf("Expected 2 matches for go language, got %d", len(matched))
	}

	// Should match "go" and "go (cargo)"
	for _, m := range matched {
		if m.Platform != "go" && m.Platform != "go (cargo)" {
			t.Errorf("Unexpected platform match: %s", m.Platform)
		}
	}
}

func TestSelectDefaultCommand(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "brew", Command: "brew install"},
		{ID: "2", Platform: "cargo", Command: "cargo install"},
	}

	defaultCmd := SelectDefaultCommand(installs, false, "ubuntu")

	if defaultCmd == nil {
		t.Fatal("Expected default command, got nil")
	}

	if defaultCmd.ID != "1" {
		t.Errorf("Expected first command (ID=1), got ID=%s", defaultCmd.ID)
	}
}

func TestFormatCommands(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "brew", Command: "brew install xxx"},
		{ID: "2", Platform: "cargo", Command: "cargo install xxx"},
	}

	defaultCmd := &installs[0]
	formatted := FormatCommands(installs, defaultCmd)

	if len(formatted) != 2 {
		t.Fatalf("Expected 2 formatted commands, got %d", len(formatted))
	}

	if !formatted[0].IsDefault {
		t.Error("Expected first command to be marked as default")
	}

	if formatted[1].IsDefault {
		t.Error("Expected second command to NOT be marked as default")
	}
}

func TestMiseModeForcesLANG(t *testing.T) {
	installs := []db.InstallInstruction{
		{ID: "1", Platform: "go", Command: "go install github.com/user/repo"},
		{ID: "2", Platform: "rust", Command: "cargo install crate"},
	}

	matched, usedFallback := FilterCommands(installs, "LANG", "go")

	if len(matched) != 1 {
		t.Fatalf("Expected 1 match for go language, got %d", len(matched))
	}

	if matched[0].Platform != "go" {
		t.Errorf("Expected 'go' platform, got %s", matched[0].Platform)
	}

	if usedFallback {
		t.Error("Expected normal match, got fallback")
	}

	formatted := FormatCommands(matched, &matched[0])
	if len(formatted) != 1 {
		t.Fatalf("Expected 1 formatted command, got %d", len(formatted))
	}

	// Transform with mise mode
	transformed := TransformToMise(formatted[0].Command)
	expected := "mise use --global go:github.com/user/repo"
	if transformed != expected {
		t.Errorf("Expected %q, got %q", expected, transformed)
	}
}

func TestGenerateVirtualInstallInstructions(t *testing.T) {
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
			name:  "empty_input",
			input: []db.InstallInstruction{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateVirtualInstallInstructions(tt.input)

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

func TestExtractBackendType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected BackendType
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
			result := extractBackendType(tt.input)
			if result != tt.expected {
				t.Errorf("extractBackendType(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
