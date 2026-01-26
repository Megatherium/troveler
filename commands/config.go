package commands

import (
	"context"

	"troveler/config"
)

type contextKey struct{}

func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

func GetConfig(ctx context.Context) *config.Config {
	if cfg, ok := ctx.Value(contextKey{}).(*config.Config); ok {
		return cfg
	}
	return nil
}

func LoadConfig(path string) (*config.Config, error) {
	return config.Load(path)
}
