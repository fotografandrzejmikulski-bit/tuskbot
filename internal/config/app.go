package config

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type AppConfig struct {
	RuntimePath      string `env:"TUSK_RUNTIME_PATH" envDefault:".tuskbot"`
	MainModel        string `env:"TUSK_MAIN_MODEL" envDefault:"openrouter/google/gemma-3-27b-it:free"`
	OpenRouterAPIKey string `env:"TUSK_OPENROUTER_API_KEY,required,notEmpty"`

	// Transport Flags
	EnableTelegram bool `env:"TUSK_ENABLE_TELEGRAM" envDefault:"false"`
	EnableCLI      bool `env:"TUSK_ENABLE_CLI" envDefault:"true"`

	// Context Management
	ContextWindowSize int `env:"TUSK_CONTEXT_WINDOW_SIZE" envDefault:"30"`

	Provider string
	Model    string
}

func NewAppConfig(ctx context.Context) *AppConfig {
	c := &AppConfig{}
	if err := env.Parse(c); err != nil {
		log.FromCtx(ctx).Fatal().Err(err).Msg("failed to parse App config")
	}

	i := strings.Index(c.MainModel, "/")
	if i > 0 {
		c.Provider = c.MainModel[:i]
		c.Model = c.MainModel[i+1:]
	}

	return c
}

func (c *AppConfig) GetRuntimePath() string {
	return c.RuntimePath
}

func (c *AppConfig) GetSystemPath() string {
	return filepath.Join(c.RuntimePath, "SYSTEM.md")
}

func (c *AppConfig) GetIdentityPath() string {
	return filepath.Join(c.RuntimePath, "IDENTITY.md")
}

func (c *AppConfig) GetUserProfilePath() string {
	return filepath.Join(c.RuntimePath, "USER.md")
}

func (c *AppConfig) GetMemoryPath() string {
	return filepath.Join(c.RuntimePath, "MEMORY.md")
}

func (c *AppConfig) GetDatabasePath() string {
	return filepath.Join(c.RuntimePath, "tuskbot.db")
}

func (c *AppConfig) GetMCPConfigPath() string {
	return filepath.Join(c.RuntimePath, "mcp_config.json")
}
