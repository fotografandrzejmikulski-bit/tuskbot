package config

import (
	"context"

	"github.com/caarlos0/env/v11"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type OpenRouterConfig struct {
	APIKey string `env:"OPENROUTER_API_KEY,required,notEmpty"`
	Model  string `env:"OPENROUTER_MODEL,required,notEmpty" envDefault:"google/gemma-3-27b-it:free"`
}

func NewOpenRouterConfig(ctx context.Context) *OpenRouterConfig {
	c := &OpenRouterConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse Badger config")
	}
	return c
}
