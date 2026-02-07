package srv

import "context"

// cleanupService implements Service interface.
type cleanupService struct {
	cleanup func() error
}

func (c *cleanupService) Start(ctx context.Context) error {
	// No-op for a cleanup-only service
	return nil
}

func (c *cleanupService) Shutdown(ctx context.Context) error {
	if c.cleanup != nil {
		return c.cleanup()
	}
	return nil
}

func NewCleanup(fn func() error) Service {
	return &cleanupService{cleanup: fn}
}
