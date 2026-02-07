package main

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/providers/llm"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
	"github.com/sandevgo/tuskbot/internal/providers/rag"
	"github.com/sandevgo/tuskbot/internal/providers/tools"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/internal/storage/sqlite"
	"github.com/sandevgo/tuskbot/internal/transport/telegram"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/sandevgo/tuskbot/pkg/srv"
)

func NewServices(ctx context.Context) []srv.Service {
	logger := log.FromCtx(ctx)
	services := make([]srv.Service, 0)

	// 1. Configuration
	appCfg := config.NewAppConfig(ctx)
	ragCfg := config.NewRAGConfig(ctx)

	// 2. Storage
	db, historyRepo, err := initStorage(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize storage")
	}
	services = append(services, srv.NewCleanup(db.Close))

	// 3. AI Provider
	aiProvider, err := llm.NewProvider(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize LLM provider")
	}

	// 4. RAG Provider
	embedder, err := rag.NewEmbedder(ragCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize RAG embedder")
	}
	services = append(services, srv.NewCleanup(embedder.Shutdown))

	// 5. MCP & Tools
	mcpManager, err := initMCP(ctx, appCfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize MCP manager")
	}
	services = append(services, mcpManager)

	// 6. Agent Service
	ag := agent.NewAgent(appCfg, aiProvider, mcpManager, historyRepo, embedder)

	// 7. Transports
	transports, err := initTransports(ctx, appCfg, ag)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialize transports")
	}
	services = append(services, transports...)

	return services
}

func initStorage(ctx context.Context, cfg *config.AppConfig) (*sql.DB, agent.HistoryRepository, error) {
	db, err := sqlite.NewDB(ctx, cfg.GetDatabasePath())
	if err != nil {
		return nil, nil, err
	}
	return db, sqlite.NewHistory(db), nil
}

func initMCP(ctx context.Context, cfg *config.AppConfig) (*mcp.Manager, error) {
	mgr, err := mcp.NewManager(ctx, cfg.GetMCPConfigPath())
	if err != nil {
		return nil, err
	}

	// Helper to register a toolset
	register := func(t interface {
		GetDefinitions() map[string]struct {
			Description string
			Schema      string
			Handler     func(context.Context, json.RawMessage) (string, error)
		}
	}) {
		for name, def := range t.GetDefinitions() {
			mgr.RegisterNativeTool(name, def.Description, json.RawMessage(def.Schema), def.Handler)
		}
	}

	// Register Core Tools
	register(tools.NewFilesystem(cfg.GetRuntimePath()))
	register(tools.NewShell(cfg.GetRuntimePath()))
	register(tools.NewFetch())

	return mgr, nil
}

func initTransports(ctx context.Context, cfg *config.AppConfig, ag *agent.Agent) ([]srv.Service, error) {
	var services []srv.Service

	// CLI Chat used for test only
	//if cfg.EnableCLI {
	//	rlChat, err := cli.NewReadLine(ag, cfg)
	//	if err != nil {
	//		return nil, err
	//	}
	//	services = append(services, rlChat)
	//}

	// Telegram Bot
	if cfg.EnableTelegram {
		tgCfg := config.NewTelegramConfig(ctx)
		bot, err := telegram.NewBot(ctx, tgCfg, ag)
		if err != nil {
			return nil, err
		}
		services = append(services, bot)
	}

	return services, nil
}
