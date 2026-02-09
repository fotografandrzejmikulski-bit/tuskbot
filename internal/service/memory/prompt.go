package memory

import (
	"os"

	"github.com/sandevgo/tuskbot/internal/config"
	"github.com/sandevgo/tuskbot/internal/core"
)

type PromptBuilder struct {
	cfg *config.AppConfig
}

func NewPromptBuilder(cfg *config.AppConfig) *PromptBuilder {
	return &PromptBuilder{
		cfg: cfg,
	}
}

func (p *PromptBuilder) Build() []core.Message {
	messages := make([]core.Message, 0)
	readFile := func(path string) string {
		content, err := os.ReadFile(path)
		if err != nil {
			return ""
		}
		return string(content)
	}

	if content := readFile(p.cfg.GetSystemPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: content})
	}
	if content := readFile(p.cfg.GetIdentityPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "\n### YOUR IDENTITY:\n" + content})
	}
	if content := readFile(p.cfg.GetUserProfilePath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "\n### ABOUT THE USER:\n" + content})
	}
	if content := readFile(p.cfg.GetMemoryPath()); content != "" {
		messages = append(messages, core.Message{Role: "system", Content: "\n### RELEVANT MEMORY:\n" + content})
	}
	return messages
}
