package config

import "context"

type ConfigContext int

const ContextKey ConfigContext = iota

func AddToContext(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, ContextKey, cfg)
}

func GetFromContext(ctx context.Context) *Config {
	cfg, ok := ctx.Value(ContextKey).(*Config)
	if ok {
		return cfg
	}
	return &Config{}
}
