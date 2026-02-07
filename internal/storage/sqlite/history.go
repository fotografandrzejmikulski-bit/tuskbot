package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/sandevgo/tuskbot/internal/core"
	"github.com/sandevgo/tuskbot/pkg/log"
)

type History struct {
	db *sql.DB
}

func NewHistory(db *sql.DB) *History {
	return &History{db: db}
}

func (h *History) AddMessage(ctx context.Context, sessionID string, msg core.Message) error {
	toolCallsJSON, err := json.Marshal(msg.ToolCalls)
	if err != nil {
		return fmt.Errorf("failed to marshal tool calls: %w", err)
	}

	// If ToolCalls is "null" (empty slice), store as empty string to save space
	tcStr := string(toolCallsJSON)
	if tcStr == "null" {
		tcStr = ""
	}

	query := `INSERT INTO messages (session_id, role, content, tool_calls, tool_call_id) VALUES (?, ?, ?, ?, ?)`
	_, err = h.db.ExecContext(ctx, query, sessionID, msg.Role, msg.Content, tcStr, msg.ToolCallID)
	if err != nil {
		return fmt.Errorf("failed to insert message: %w", err)
	}
	return nil
}

func (h *History) GetMessages(ctx context.Context, sessionID string, limit int) ([]core.Message, error) {
	// Fetch the LAST 'limit' messages by ordering DESC
	query := `SELECT role, content, tool_calls, tool_call_id FROM messages WHERE session_id = ? ORDER BY id DESC LIMIT ?`

	rows, err := h.db.QueryContext(ctx, query, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}
	defer rows.Close()

	var messages []core.Message
	for rows.Next() {
		var msg core.Message
		var content, toolCallsStr, toolCallID sql.NullString

		// Use NullString to safely handle potential NULLs in DB
		if err := rows.Scan(&msg.Role, &content, &toolCallsStr, &toolCallID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		msg.Content = content.String
		msg.ToolCallID = toolCallID.String

		if toolCallsStr.Valid && toolCallsStr.String != "" && toolCallsStr.String != "null" {
			if err := json.Unmarshal([]byte(toolCallsStr.String), &msg.ToolCalls); err != nil {
				return nil, fmt.Errorf("failed to unmarshal tool calls: %w", err)
			}
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// The query returned messages in Reverse Chronological Order (Newest -> Oldest).
	// We need to reverse them back to Chronological Order (Oldest -> Newest) for the LLM.
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	log.FromCtx(ctx).Debug().Int("count", len(messages)).Msg("loaded history messages")
	return messages, nil
}
