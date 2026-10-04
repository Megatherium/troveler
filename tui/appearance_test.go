package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"

	"troveler/config"
	"troveler/db"
)

func TestModelAppearanceColors(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previousProfile) })
	for _, test := range []struct {
		name, settings string
		colors         []string
	}{
		{"builtin gradient", "", []string{"38;2;144;238;144", "38;2;139;232;140", "38;2;134;226;136"}},
		{"plain default", `theme = "default"`, []string{"38;2;255;255;255", "38;2;255;255;255", "38;2;255;255;255"}},
		{"custom palette", "theme = 'custom'\ngradient_colors = ['#123456', '#FF0000']", []string{"38;2;18;52;86", "38;2;255;0;0", "38;2;18;52;86"}},
		{"gradient palette override", `gradient_colors = ["#123456", "#ff0000"]`, []string{"38;2;18;52;86", "38;2;255;0;0", "38;2;18;52;86"}},
		{"plain palette override", "theme = 'default'\ngradient_colors = ['#123456', '#FF0000']", []string{"38;2;18;52;86", "38;2;18;52;86", "38;2;18;52;86"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg := loadAppearanceForTest(t, test.settings)
			m := NewModel(nil, cfg)
			m.toolsPanel.SetSize(140, 12)
			m.toolsPanel.SetTools([]db.SearchResult{
				{Tool: db.Tool{ID: "a", Name: "AppearanceAlpha"}},
				{Tool: db.Tool{ID: "b", Name: "AppearanceBeta"}},
				{Tool: db.Tool{ID: "c", Name: "AppearanceGamma"}},
			})
			// A model must retain its own palette when config or another model changes.
			if len(cfg.TUI.GradientColors) > 0 {
				cfg.TUI.GradientColors[0] = "#ffffff"
			}
			_ = NewModel(nil, &config.Config{TUI: config.TUIConfig{Theme: "custom", GradientColors: []string{"#abcdef"}}})
			for _, state := range []string{"normal", "selected", "marked selected", "marked blurred"} {
				if state == "selected" {
					m.toolsPanel.Focus()
				} else if state == "marked selected" {
					m.toolsPanel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
				} else if state == "marked blurred" {
					m.toolsPanel.Blur()
				}
				lines := strings.Split(m.toolsPanel.View(), "\n")
				for row, wantColor := range test.colors {
					line := lines[row+2]
					if !strings.Contains(line, wantColor) {
						t.Fatalf("%s row %d lacks configured RGB color %q: %q", state, row, wantColor, line)
					}
				}
				var background string
				switch state {
				case "selected":
					background = "48;2;0;51;51"
				case "marked selected":
					background = "48;2;85;85;0"
				case "marked blurred":
					background = "48;2;51;51;0"
				}
				if background != "" && !strings.Contains(lines[2], background) {
					t.Fatalf("%s lost its visible background: %q", state, lines[2])
				}
				if strings.HasPrefix(state, "marked") && (!m.toolsPanel.IsMarked("a") || !strings.Contains(lines[2], "●")) {
					t.Fatalf("%s did not render a marked tool: %q", state, lines[2])
				}
			}
		})
	}
}

func TestRunRejectsInvalidManualAppearance(t *testing.T) {
	err := Run(nil, &config.Config{TUI: config.TUIConfig{Theme: "unknown"}})
	if err == nil || !strings.Contains(err.Error(), "invalid TUI config") || !strings.Contains(err.Error(), "tui.theme") {
		t.Fatalf("expected appearance validation before starting the terminal: %v", err)
	}
}

func TestModelAppearanceTaglineWidth(t *testing.T) {
	for _, test := range []struct {
		name, tagline                string
		limit, panelWidth, wantWidth int
	}{
		{"configured", strings.Repeat("abcdefghij", 10), 12, 140, 12},
		{"default", strings.Repeat("abcdefghij", 10), 0, 140, 40},
		{"terminal limited", strings.Repeat("abcdefghij", 10), 40, 80, 21},
		{"large limit", strings.Repeat("abcdefghij", 10), 1000, 140, 81},
		{"one column", "abcdefghij", 1, 140, 1},
		{"two columns", "abcdefghij", 2, 140, 2},
		{"three columns", "abcdefghij", 3, 140, 3},
		{"wide unicode", strings.Repeat("界", 20), 12, 140, 12},
		{"graphemes", strings.Repeat("🇩🇪e\u0301", 20), 12, 140, 12},
	} {
		t.Run(test.name, func(t *testing.T) {
			cfg := loadAppearanceForTest(t, fmt.Sprintf("tagline_max_width = %d", test.limit))
			m := NewModel(nil, cfg)
			m.toolsPanel.SetSize(test.panelWidth, 12)
			m.toolsPanel.SetTools([]db.SearchResult{{Tool: db.Tool{Name: "WidthProbe", Tagline: test.tagline}}})
			view := ansi.Strip(m.toolsPanel.View())
			lines := strings.Split(view, "\n")
			if len(lines) != 4 || !utf8.ValidString(view) {
				t.Fatalf("table must stay valid UTF-8 with one physical line per row: %q", view)
			}
			cells := strings.Split(lines[2], " │ ")
			if len(cells) != 4 || ansi.StringWidth(cells[1]) != test.wantWidth {
				t.Fatalf("tagline column = %q; want width %d", cells, test.wantWidth)
			}
			if test.wantWidth >= 3 && !strings.HasSuffix(strings.TrimSpace(cells[1]), "...") {
				t.Fatalf("long tagline must show truncation: %q", cells[1])
			}
		})
	}
}

func loadAppearanceForTest(t *testing.T, settings string) *config.Config {
	t.Helper()
	t.Setenv("TROVELER_DSN", "")
	path := filepath.Join(t.TempDir(), "appearance.toml")
	if err := os.WriteFile(path, []byte("[tui]\n"+settings+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}
