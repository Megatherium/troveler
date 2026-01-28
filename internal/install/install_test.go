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
