package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(`[install]
fallback_platform = "LANG"`), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if cfg.DSN == "" {
		t.Error("Expected DSN to have default value")
	}

	if cfg.Install.FallbackPlatform != "LANG" {
		t.Errorf("Expected fallback_platform 'LANG', got '%s'", cfg.Install.FallbackPlatform)
	}

	if cfg.Search.TaglineWidth != 50 {
		t.Errorf("Expected default TaglineWidth 50, got %d", cfg.Search.TaglineWidth)
	}
}

func TestLoadConfigWithCustomTaglineWidth(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(`[search]
tagline_width = 30`), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if cfg.Search.TaglineWidth != 30 {
		t.Errorf("Expected TaglineWidth 30, got %d", cfg.Search.TaglineWidth)
	}
}

func TestLoadConfigWithZeroTaglineWidth(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	//nolint:gosec // G306: test file
	_ = os.WriteFile(configPath, []byte(`[search]
tagline_width = 0`), 0644)

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load config failed: %v", err)
	}

	if cfg.Search.TaglineWidth != 50 {
		t.Errorf("Expected default TaglineWidth 50 when set to 0, got %d", cfg.Search.TaglineWidth)
	}
}
