package rag

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/pkg/llamacpp"
	"github.com/sandevgo/tuskbot/pkg/log"
)

// TODO: move to config
const (
	defaultChunkSize    = 500
	defaultChunkOverlap = 100
)

var (
	reSpaces   = regexp.MustCompile(`[\t\p{Zs}]+`)
	reNewlines = regexp.MustCompile(`\n{2,}`)
)

type Embedder struct {
	llm *llamacpp.LlamaEmbedder
}

func NewEmbedder(cfg *config.RAGConfig) (*Embedder, error) {
	llm, err := llamacpp.NewLlamaEmbedder(cfg.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize llama embedder: %w", err)
	}
	return &Embedder{llm: llm}, nil
}

func (e *Embedder) Embed(ctx context.Context, text string) ([][]float32, error) {
	text = normalizeText(text)
	chunks := chunkText(text, defaultChunkSize, defaultChunkOverlap)
	embeddings := make([][]float32, 0, len(chunks))
	for _, chunk := range chunks {
		log.FromCtx(ctx).Debug().Str("chunk", chunk).Msg("embedding chunk")
		emb, err := e.llm.Embed(chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk: %w", err)
		}
		embeddings = append(embeddings, emb)
	}
	return embeddings, nil
}

func (e *Embedder) Shutdown() error {
	e.llm.Free()
	return nil
}

func normalizeText(text string) string {
	text = reSpaces.ReplaceAllString(text, " ")
	text = reNewlines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

func chunkText(text string, chunkSize int, overlap int) []string {
	runes := []rune(text)
	var chunks []string
	for i := 0; i < len(runes); {
		start := i
		end := i + chunkSize

		if end >= len(runes) {
			chunks = append(chunks, string(runes[i:]))
			break
		}

		actualEnd := end
		for j := end; j > i; j-- {
			if runes[j] == ' ' || runes[j] == '\n' {
				actualEnd = j
				break
			}
		}

		if actualEnd == i {
			actualEnd = end
		}
		chunks = append(chunks, string(runes[i:actualEnd]))

		i = actualEnd - overlap

		if i <= start {
			i = actualEnd
		}
	}
	return chunks
}
