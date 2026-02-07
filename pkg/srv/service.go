package srv

import (
	"context"

	"github.com/sandevgo/tuskbot/pkg/log"
)

type Service interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

func StartServices(ctx context.Context, services []Service) {
	logger := log.FromCtx(ctx)
	for _, service := range services {
		go func(service Service) {
			if err := service.Start(ctx); err != nil {
				logger.Fatal().Err(err).Msgf("%T failed to start", service)
			}
		}(service)
	}
}

func ShutdownServices(ctx context.Context, services []Service) {
	<-ctx.Done()
	for _, service := range services {
		if err := service.Shutdown(ctx); err != nil {
			log.FromCtx(ctx).Error().Err(err).Msgf("%T failed to shutdown", service)
		}
	}
}
