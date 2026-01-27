package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DSN     string        `toml:"dsn"`
	Install InstallConfig `toml:"install"`
	Search  SearchConfig  `toml:"search"`
}

type InstallConfig struct {
	FallbackPlatform string `toml:"fallback_platform"`
	PlatformOverride string `toml:"platform_override"`
	AlwaysRun        bool   `toml:"always_run"`
	UseSudo          string `toml:"use_sudo"`
}

type SearchConfig struct {
	TaglineWidth int `toml:"tagline_width"`
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
