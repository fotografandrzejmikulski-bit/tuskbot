package config

import (
	"context"

	"github.com/caarlos0/env/v11"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type TelegramConfig struct {
	Token   string `env:"TELEGRAM_TOKEN,required,notEmpty"`
	OwnerID int64  `env:"TELEGRAM_OWNER_ID,required"`
}

func NewTelegramConfig(ctx context.Context) *TelegramConfig {
	c := &TelegramConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse Telegram config")
	}
	return c
}
