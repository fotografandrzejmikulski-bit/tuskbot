package memory

import (
	"context"

	"github.com/sandevgo/tuskbot/internal/core"
)

type Repository interface {
	AddMessage(ctx context.Context, sessionID string, msg core.Message) error
	GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error)
}
