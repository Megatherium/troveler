// Package config handles application configuration loading and defaults.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds all application configuration.
type Config struct {
	DSN          string        `toml:"dsn"`
	DefaultToTUI bool          `toml:"default_to_tui"`
	Install      InstallConfig `toml:"install"`
	Search       SearchConfig  `toml:"search"`
	TUI          TUIConfig     `toml:"tui"`
}

// InstallConfig holds install-related settings.
type InstallConfig struct {
	FallbackPlatform string `toml:"fallback_platform"`
	PlatformOverride string `toml:"platform_override"`
	AlwaysRun        bool   `toml:"always_run"`
	UseSudo          string `toml:"use_sudo"`
}

// SearchConfig holds search-related settings.
type SearchConfig struct {
	TaglineWidth int `toml:"tagline_width"`
}

// TUIConfig holds TUI-related settings.
type TUIConfig struct {
	Theme           string   `toml:"theme"`
	TaglineMaxWidth int      `toml:"tagline_max_width"`
	GradientColors  []string `toml:"gradient_colors"`
}

// Load reads configuration from the given path, applying defaults.
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = defaultConfigPath()
	}

	cfg := &Config{}

	if _, err := os.Stat(configPath); err == nil {
		if _, err := toml.DecodeFile(configPath, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	if cfg.DSN == "" {
		cfg.DSN = defaultDSN()
	}

	if dsn := os.Getenv("TROVELER_DSN"); dsn != "" {
		cfg.DSN = dsn
	}

	if cfg.Search.TaglineWidth == 0 {
		cfg.Search.TaglineWidth = 50
	}

	// TUI defaults
	if cfg.TUI.Theme == "" {
		cfg.TUI.Theme = "gradient"
	}

	if cfg.TUI.TaglineMaxWidth == 0 {
		cfg.TUI.TaglineMaxWidth = 40
	}

	return cfg, nil
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}

	return filepath.Join(home, ".config", "troveler", "config.toml")
}

func defaultDSN() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "file:troveler.db?cache=shared&mode=rwc"
	}
	dbPath := filepath.Join(home, ".local", "share", "troveler", "troveler.db")

	return "file:" + dbPath + "?cache=shared&mode=rwc"
}
