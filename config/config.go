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
}

type InstallConfig struct {
	FallbackPlatform string `toml:"fallback_platform"`
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
