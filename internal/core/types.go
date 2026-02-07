package core

import "encoding/json"

const (
	TuskName          = "TuskBot"
	TuskUserAgent     = "TuskBot-Agent/0.1"
	TuskRepositoryURL = "https://github.com/sandevgo/tuskbot"
	TaskVersion       = "0.1.0"
)

type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Reasoning  string     `json:"reasoning,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}
