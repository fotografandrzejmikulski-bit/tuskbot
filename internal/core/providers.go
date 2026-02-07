package core

import "context"

type AIProvider interface {
	Chat(ctx context.Context, history []Message, tools []Tool) (Message, error)
}

type Embedder interface {
	Embed(ctx context.Context, text string) ([][]float32, error)
}

type MCPServer interface {
	GetTools(ctx context.Context) ([]Tool, error)
	CallTool(ctx context.Context, name string, args string) (string, error)
}
