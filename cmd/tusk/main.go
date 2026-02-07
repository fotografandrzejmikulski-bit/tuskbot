package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/pkg/log"
	"github.com/sandevgo/tuskbot/pkg/srv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// logger setup
	var flushLog func()
	ctx, flushLog = log.NewContextWithLogger(ctx, config.IsDebug())
	defer flushLog()

	logger := log.FromCtx(ctx)
	logger.Info().Msg("starting tuskbot")

	// Define services
	services := NewServices(ctx)

	// Start services
	srv.StartServices(ctx, services)

	// Wait for shutdown signal
	srv.ShutdownServices(ctx, services)
	logger.Info().Msg("tuskbot has been shut down gracefully")
}
