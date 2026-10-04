package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadTUIValidation(t *testing.T) {
	for _, test := range []struct {
		name, settings, errorField string
	}{
		{"unknown theme", `theme = "solarized"`, "tui.theme"},
		{"negative width", `tagline_max_width = -1`, "tui.tagline_max_width"},
		{"custom without palette", `theme = "custom"`, "tui.gradient_colors"},
		{"named color", `gradient_colors = ["red"]`, "tui.gradient_colors[0]"},
		{"short color", `gradient_colors = ["#123"]`, "tui.gradient_colors[0]"},
		{"invalid hex", `gradient_colors = ["#GG1122"]`, "tui.gradient_colors[0]"},
		{"alpha color", `gradient_colors = ["#11223344"]`, "tui.gradient_colors[0]"},
		{"invalid later color", `gradient_colors = ["#112233", ""]`, "tui.gradient_colors[1]"},
		{"gradient defaults", "", ""},
		{"default theme", `theme = "default"`, ""},
		{"single custom color", "theme = 'custom'\ngradient_colors = ['#aabbcc']", ""},
		{"gradient override", `gradient_colors = ["#123456", "#abcdef"]`, ""},
		{"zero width uses default", `tagline_max_width = 0`, ""},
		{"one column", `tagline_max_width = 1`, ""},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("TROVELER_DSN", "")
			path := filepath.Join(root, "appearance.toml")
			if err := os.WriteFile(path, []byte("[tui]\n"+test.settings+"\n"), 0600); err != nil {
				t.Fatal(err)
			}
			cfg, err := Load(path)
			if test.errorField != "" {
				if err == nil || cfg != nil || !strings.Contains(err.Error(), test.errorField) || !strings.Contains(err.Error(), path) {
					t.Fatalf("expected %s error identifying config path: cfg = %v, error = %v", test.errorField, cfg, err)
				}
			} else if err != nil || cfg == nil {
				t.Fatalf("valid appearance rejected: cfg = %v, error = %v", cfg, err)
			}
		})
	}
}
