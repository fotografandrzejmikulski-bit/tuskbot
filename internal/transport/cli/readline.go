package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/internal/service/agent"
	"github.com/sandevgo/tuskbot/pkg/log"
)

const defaultSessionID = "cli-local"

type ReadLine struct {
	cfg   *config.AppConfig
	agent *agent.Agent
	rl    *readline.Instance
}

func NewReadLine(agent *agent.Agent, cfg *config.AppConfig) (*ReadLine, error) {
	// Ensure runtime directory exists
	if err := os.MkdirAll(cfg.RuntimePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create runtime directory: %w", err)
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          ">>> ",
		HistoryFile:     filepath.Join(cfg.RuntimePath, "input_history"),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return nil, err
	}

	return &ReadLine{
		cfg:   cfg,
		agent: agent,
		rl:    rl,
	}, nil
}

func (r *ReadLine) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("ReadLine chat started. Type 'exit' to quit.")

	for {
		// Check context before blocking read
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := r.rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if len(line) == 0 {
					return nil // Exit on Ctrl+C
				}
				continue
			} else if err == io.EOF {
				return nil
			}
			return err
		}

		line = strings.TrimSpace(line)
		if line == "exit" {
			return nil
		}
		if line == "" {
			continue
		}

		// Agent Run
		_, err = r.agent.Run(ctx, defaultSessionID, line, func(msg core.Message) {
			// Display Reasoning (Thought Chain) if present
			if msg.Reasoning != "" {
				fmt.Fprintf(r.rl.Stdout(), "\033[38;5;240m[Thinking]\n%s\033[0m\n", msg.Reasoning)
			}

			// Display content if any
			if msg.Content != "" {
				fmt.Fprintf(r.rl.Stdout(), "%s\n", msg.Content)
			}

			// Display Tool Calls
			if len(msg.ToolCalls) > 0 {
				fmt.Fprintf(r.rl.Stdout(), "[System] Processing %d tool call(s)...\n", len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					fmt.Fprintf(r.rl.Stdout(), "  > Calling %s %s...\n", tc.Function.Name, tc.Function.Arguments)
				}
			}
		})

		if err != nil {
			logger.Error().Err(err).Msg("agent run failed")
			fmt.Fprintf(r.rl.Stdout(), "Error: %v\n", err)
		}
	}
}

func (r *ReadLine) Shutdown(ctx context.Context) error {
	if r.rl != nil {
		return r.rl.Close()
	}
	return nil
}
