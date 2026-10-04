// Package config handles application configuration loading and defaults.
package config

import (
	"errors"
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

	defaultDatabaseDSN string
	defaultDatabaseDir string
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
// Only a missing implicit default file is optional; explicit paths must be readable.
func Load(configPath string) (*Config, error) {
	optional := configPath == ""
	if optional {
		configPath = defaultConfigPath()
	}

	cfg := &Config{}

	contents, err := os.ReadFile(configPath)
	if err != nil {
		if !optional || !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read config file %q: %w", configPath, err)
		}
	} else if _, err := toml.Decode(string(contents), cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file %q: %w", configPath, err)
	}

	if cfg.DSN == "" {
		cfg.DSN, cfg.defaultDatabaseDir = defaultDSN()
		cfg.defaultDatabaseDSN = cfg.DSN
	}

	if dsn := os.Getenv("TROVELER_DSN"); dsn != "" {
		cfg.DSN = dsn
		cfg.defaultDatabaseDir = ""
		cfg.defaultDatabaseDSN = ""
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

// EnsureDatabaseDir prepares the data directory only for the loaded default DSN.
// Explicit DSNs, including overrides made after Load, remain caller-managed.
func (c *Config) EnsureDatabaseDir() error {
	if c.defaultDatabaseDir == "" || c.DSN != c.defaultDatabaseDSN {
		return nil
	}
	if err := os.MkdirAll(c.defaultDatabaseDir, 0700); err != nil {
		return fmt.Errorf("create default database directory %q: %w", c.defaultDatabaseDir, err)
	}
	return nil
}

func defaultConfigPath() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if !filepath.IsAbs(configHome) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "config.toml"
		}
		configHome = filepath.Join(home, ".config")
	}

	return filepath.Join(configHome, "troveler", "config.toml")
}

func defaultDSN() (string, string) {
	dataHome := os.Getenv("XDG_DATA_HOME")
	if !filepath.IsAbs(dataHome) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "file:troveler.db?cache=shared&mode=rwc", ""
		}
		dataHome = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(dataHome, "troveler")
	dbPath := filepath.Join(dir, "troveler.db")

	return "file:" + dbPath + "?cache=shared&mode=rwc", dir
}
