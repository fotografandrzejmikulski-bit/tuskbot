package memory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type Extractor struct {
	repo     core.KnowledgeRepository
	ai       core.AIProvider
	embedder core.Embedder
	interval time.Duration
}

func NewExtractor(repo core.KnowledgeRepository, ai core.AIProvider, embedder core.Embedder) *Extractor {
	return &Extractor{
		repo:     repo,
		ai:       ai,
		embedder: embedder,
		interval: 2 * time.Minute,
	}
}

func (e *Extractor) Start(ctx context.Context) error {
	logger := log.FromCtx(ctx)
	logger.Info().Msg("Knowledge Extractor service started")

	// Initial run after a short delay to let app startup
	time.AfterFunc(10*time.Second, func() {
		if err := e.processBatch(ctx); err != nil {
			logger.Error().Err(err).Msg("failed to process initial knowledge extraction batch")
		}
	})

	ticker := time.NewTicker(e.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := e.processBatch(ctx); err != nil {
				logger.Error().Err(err).Msg("failed to process knowledge extraction batch")
			}
		}
	}
}

func (e *Extractor) Shutdown(ctx context.Context) error {
	return nil
}

func (e *Extractor) processBatch(ctx context.Context) error {
	logger := log.FromCtx(ctx)

	// 1. Fetch unextracted messages
	msgs, err := e.repo.GetUnextractedMessages(ctx, 20)
	if err != nil {
		return fmt.Errorf("failed to fetch messages: %w", err)
	}

	if len(msgs) == 0 {
		return nil
	}

	logger.Debug().Int("count", len(msgs)).Msg("extracting knowledge from messages")

	// 2. Build Prompt
	var conversation strings.Builder
	var msgIDs []int64
	for _, m := range msgs {
		conversation.WriteString(fmt.Sprintf("%s: %s\n", strings.ToUpper(m.Role), m.Content))
		msgIDs = append(msgIDs, m.ID)
	}

	prompt := fmt.Sprintf(`Analyze the following conversation and extract atomic facts about the user, their projects, preferences, or specific instructions they gave.
Ignore trivial chit-chat.
Return the result as a JSON list of objects with "fact" (string) and "category" (string: "preference", "user_fact", "project", "instruction").

Conversation:
%s`, conversation.String())

	// 3. Call LLM
	resp, err := e.ai.Chat(ctx, []core.Message{
		{Role: core.RoleSystem, Content: "You are a knowledge extraction system. Output only valid JSON."},
		{Role: core.RoleUser, Content: prompt},
	}, nil)
	if err != nil {
		return fmt.Errorf("llm extraction failed: %w", err)
	}

	// 4. Parse Response
	var facts []struct {
		Fact     string `json:"fact"`
		Category string `json:"category"`
	}

	content := resp.Content
	// Attempt to clean markdown code blocks and find the JSON array
	if idx := strings.Index(content, "["); idx != -1 {
		content = content[idx:]
	}
	if idx := strings.LastIndex(content, "]"); idx != -1 {
		content = content[:idx+1]
	}

	if err := json.Unmarshal([]byte(content), &facts); err != nil {
		// If parsing fails, we log and mark messages as extracted anyway to prevent infinite loops.
		logger.Error().Err(err).Str("content", resp.Content).Msg("failed to parse extraction JSON, skipping batch")
		return e.repo.MarkMessagesExtracted(ctx, msgIDs)
	}

	// 5. Embed and Save Facts
	for _, f := range facts {
		chunks, err := e.embedder.Embed(ctx, f.Fact)
		if err != nil {
			logger.Error().Err(err).Str("fact", f.Fact).Msg("failed to embed fact")
			continue
		}
		if len(chunks) == 0 {
			continue
		}

		hash := sha256.Sum256([]byte(f.Fact))
		factHash := hex.EncodeToString(hash[:])

		storedFact := core.StoredKnowledge{
			Fact:      f.Fact,
			Category:  f.Category,
			Source:    "extracted",
			Embedding: chunks[0],
			FactHash:  factHash,
		}

		if err := e.repo.SaveFact(ctx, storedFact); err != nil {
			// Ignore unique constraint errors (duplicates)
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") && !strings.Contains(err.Error(), "constraint failed") {
				logger.Error().Err(err).Msg("failed to save fact")
			}
		} else {
			logger.Info().Str("fact", f.Fact).Msg("knowledge extracted and saved")
		}
	}

	// 6. Mark messages as processed
	if err := e.repo.MarkMessagesExtracted(ctx, msgIDs); err != nil {
		return fmt.Errorf("failed to mark messages extracted: %w", err)
	}

	return nil
}
