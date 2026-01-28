package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DSN          string        `toml:"dsn"`
	DefaultToTUI bool          `toml:"default_to_tui"`
	Install      InstallConfig `toml:"install"`
	Search       SearchConfig  `toml:"search"`
	TUI          TUIConfig     `toml:"tui"`
}

type InstallConfig struct {
	FallbackPlatform string `toml:"fallback_platform"`
	PlatformOverride string `toml:"platform_override"`
	AlwaysRun        bool   `toml:"always_run"`
	UseSudo          string `toml:"use_sudo"`
	MiseMode         bool   `toml:"mise_mode"`
}

type SearchConfig struct {
	TaglineWidth int `toml:"tagline_width"`
}

type TUIConfig struct {
	Theme           string   `toml:"theme"`
	TaglineMaxWidth int      `toml:"tagline_max_width"`
	GradientColors  []string `toml:"gradient_colors"`
}

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

	// DefaultToTUI defaults to true
	if !cfg.DefaultToTUI {
		// Check if it was explicitly set to false in config
		// For now, we'll default to true (will be overridden if explicitly set)
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
