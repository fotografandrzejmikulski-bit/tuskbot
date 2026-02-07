package llm

import (
	"context"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/service/agent"
)

// NewProvider creates the appropriate AIProvider based on configuration.
func NewProvider(ctx context.Context, appCfg *config.AppConfig) (agent.AIProvider, error) {
	switch appCfg.LLMProvider {
	case "openrouter":
		cfg := config.NewOpenRouterConfig(ctx)
		return NewOpenRouter(cfg), nil
	// Future: case "openai": ...
	// Future: case "ollama": ...
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", appCfg.LLMProvider)
	}
}
