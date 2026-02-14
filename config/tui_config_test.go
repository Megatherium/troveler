package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTUIDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Empty config file
	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(""), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.TUI.Theme != "gradient" {
		t.Errorf("Expected default theme 'gradient', got %s", cfg.TUI.Theme)
	}

	if cfg.TUI.TaglineMaxWidth != 40 {
		t.Errorf("Expected default tagline_max_width 40, got %d", cfg.TUI.TaglineMaxWidth)
	}
}

func TestLoadTUICustom(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(`
[tui]
theme = "custom"
tagline_max_width = 60
gradient_colors = ["#FF0000", "#00FF00", "#0000FF"]
`), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.TUI.Theme != "custom" {
		t.Errorf("Expected theme 'custom', got %s", cfg.TUI.Theme)
	}

	if cfg.TUI.TaglineMaxWidth != 60 {
		t.Errorf("Expected tagline_max_width 60, got %d", cfg.TUI.TaglineMaxWidth)
	}

	if len(cfg.TUI.GradientColors) != 3 {
		t.Errorf("Expected 3 gradient colors, got %d", len(cfg.TUI.GradientColors))
	}
}

func TestDefaultToTUI(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(`
default_to_tui = true
`), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !cfg.DefaultToTUI {
		t.Error("Expected default_to_tui to be true")
	}
}
