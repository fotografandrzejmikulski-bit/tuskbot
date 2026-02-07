package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/providers/mcp"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type HistoryRepository interface {
	AddMessage(ctx context.Context, sessionID string, msg core.Message) error
	GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error)
}

type AIProvider interface {
	Chat(ctx context.Context, history []core.Message, tools []core.Tool) (core.Message, error)
}

type Agent struct {
	appCfg   *config.AppConfig
	ai       AIProvider
	mcpMgr   *mcp.Manager
	repo     HistoryRepository
	embedder core.Embedder
}

func NewAgent(
	appCfg *config.AppConfig,
	ai AIProvider,
	mcpMgr *mcp.Manager,
	repo HistoryRepository,
	embedder core.Embedder,
) *Agent {
	return &Agent{
		appCfg:   appCfg,
		ai:       ai,
		mcpMgr:   mcpMgr,
		repo:     repo,
		embedder: embedder,
	}
}

func (a *Agent) Run(ctx context.Context, sessionID string, input string, onUpdate func(core.Message)) (string, error) {
	logger := log.FromCtx(ctx)

	userMsg := core.Message{Role: "user", Content: input}
	if err := a.repo.AddMessage(ctx, sessionID, userMsg); err != nil {
		return "", fmt.Errorf("failed to save user message: %w", err)
	}

	a.debugEmbed(ctx, "User Input", input)

	var finalContent string

	for {
		tools, err := a.mcpMgr.GetTools(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get tools: %w", err)
		}

		systemMsgs := a.buildSystemPrompt()
		history, err := a.repo.GetMessages(ctx, sessionID, a.appCfg.ContextWindowSize)
		if err != nil {
			return "", fmt.Errorf("failed to fetch history: %w", err)
		}
		messages := append(systemMsgs, history...)

		responseMsg, err := a.ai.Chat(ctx, messages, tools)
		if err != nil {
			return "", fmt.Errorf("ai chat error: %w", err)
		}

		if err := a.repo.AddMessage(ctx, sessionID, responseMsg); err != nil {
			logger.Error().Err(err).Msg("failed to save assistant message")
		}

		if responseMsg.Content != "" {
			a.debugEmbed(ctx, "AI Response", responseMsg.Content)
		}

		if onUpdate != nil {
			onUpdate(responseMsg)
		}

		if responseMsg.Content != "" {
			finalContent = responseMsg.Content
		}

		if len(responseMsg.ToolCalls) == 0 {
			break
		}

		for _, tc := range responseMsg.ToolCalls {
			logger.Info().Str("tool", tc.Function.Name).Msg("executing tool")

			result, err := a.mcpMgr.CallTool(ctx, tc.Function.Name, tc.Function.Arguments)
			if err != nil {
				result = fmt.Sprintf("Error executing tool: %v", err)
			}

			toolMsg := core.Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			}
			if err := a.repo.AddMessage(ctx, sessionID, toolMsg); err != nil {
				logger.Error().Err(err).Msg("failed to save tool message")
			}

			a.debugEmbed(ctx, fmt.Sprintf("Tool Result (%s)", tc.Function.Name), result)
		}
	}

	return finalContent, nil
}

func (a *Agent) debugEmbed(ctx context.Context, label, text string) {
	if a.embedder == nil || text == "" {
		return
	}

	logger := log.FromCtx(ctx)

	chunks, err := a.embedder.Embed(ctx, text)
	if err != nil {
		logger.Error().Err(err).Str("label", label).Msg("failed to embed text")
		return
	}

	if len(chunks) == 0 {
		logger.Debug().Str("label", label).Msg("failed to embed text: no chunks")
	}

	limit := 5
	if len(chunks[0]) < limit {
		limit = len(chunks[0])
	}
	logger.Debug().Str("label", label).Msgf("Embedding: %v", chunks[0][:limit])
}

func (a *Agent) buildSystemPrompt() []core.Message {
	messages := make([]core.Message, 0)
	readFile := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		return string(content)
	}

	if content := readFile(a.appCfg.GetSystemPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	if content := readFile(a.appCfg.GetIdentityPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "YOUR IDENTITY:\n" + content})
	}
	if content := readFile(a.appCfg.GetUserProfilePath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "ABOUT THE USER:\n" + content})
	}
	if content := readFile(a.appCfg.GetMemoryPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "RELEVANT MEMORY:\n" + content})
	}
	return messages
}
