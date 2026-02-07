package config

import (
	"context"
	"path/filepath"

	"github.com/caarlos0/env/v9"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type AppConfig struct {
	RuntimePath string `env:"TUSKBOT_RUNTIME_PATH" envDefault:".tuskbot"`
	// Allow selecting the provider
	LLMProvider string `env:"LLM_PROVIDER" envDefault:"openrouter"`

	// Transport Flags
	EnableTelegram bool `env:"ENABLE_TELEGRAM" envDefault:"false"`
	EnableCLI      bool `env:"ENABLE_CLI" envDefault:"true"`

	// Context Management
	ContextWindowSize int `env:"CONTEXT_WINDOW_SIZE" envDefault:"30"`
}

func NewAppConfig(ctx context.Context) *AppConfig {
	c := &AppConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse App config")
	}
	return c
}

func (c AppConfig) GetRuntimePath() string {
	return c.RuntimePath
}

func (c AppConfig) GetSystemPath() string {
	return filepath.Join(c.RuntimePath, "SYSTEM.md")
}

func (c AppConfig) GetIdentityPath() string {
	return filepath.Join(c.RuntimePath, "IDENTITY.md")
}

func (c AppConfig) GetUserProfilePath() string {
	return filepath.Join(c.RuntimePath, "USER.md")
}

func (c AppConfig) GetMemoryPath() string {
	return filepath.Join(c.RuntimePath, "MEMORY.md")
}

func (c AppConfig) GetDatabasePath() string {
	return filepath.Join(c.RuntimePath, "tuskbot.db")
}

func (c AppConfig) GetMCPConfigPath() string {
	return filepath.Join(c.RuntimePath, "mcp_config.json")
}
