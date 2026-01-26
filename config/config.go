package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBPath string
}

func Load() (*Config, error) {
	dbPath := os.Getenv("TROVELER_DB_PATH")
	if dbPath == "" {
		dbPath = defaultDBPath()
	}

	return &Config{
		DBPath: dbPath,
	}, nil
}

func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "troveler.db"
	}
	return fmt.Sprintf("%s/.local/share/troveler/troveler.db", home)
}
