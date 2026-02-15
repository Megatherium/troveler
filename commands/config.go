package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"troveler/config"
	"troveler/db"
)

type contextKey struct{}

// WithConfig adds a config to the context.
func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

// GetConfig retrieves the config from the context.
func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(contextKey{}).(*config.Config); ok {
		return cfg
	}

	return nil
}

// LoadConfig loads the configuration from a file.
func LoadConfig(path string) (*config.Config, error) {
	return config.Load(path)
}

// WithDB provides a database connection to a command function.
func WithDB(cmd *cobra.Command, fn func(ctx context.Context, database *db.SQLiteDB) error) error {
	cfg := GetConfig(cmd.Context())
	if cfg == nil {
		return fmt.Errorf("config not loaded")
	}

	database, err := db.New(cfg.DSN)
	if err != nil {
		return fmt.Errorf("db init: %w", err)
	}
	defer func() { _ = database.Close() }()

	return fn(cmd.Context(), database)
}
