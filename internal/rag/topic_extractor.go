package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

// LLMClient defines the interface for interacting with the LLM provider.
type LLMClient interface {
	CreateCompletion(ctx context.Context, systemPrompt, contextText, userMessage string, model string, temperature float32, maxTokens int) (string, int, error)
}

// ExtractTopics generates a concise summary of the capabilities/topics provided by the source text.
// It uses gpt-4o-mini for cost-effectiveness.
func ExtractTopics(ctx context.Context, client LLMClient, content string, langCode string) (string, error) {
	// Truncate content to first ~2000 chars to save cost and focus on intro/headers
	if len(content) > 2000 {
		content = content[:2000]
	}

	config := langconfig.Get(langCode)
	systemPrompt := config.ResponseTemplates.TopicExtractionSystemPrompt
	contextText := fmt.Sprintf("Metin:\n%s", content)
	userMessage := config.ResponseTemplates.TopicExtractionUserPrompt

	// Using CreateCompletion which supports chat models (gpt-4o-mini is a chat model)
	summary, _, err := client.CreateCompletion(ctx, systemPrompt, contextText, userMessage, "gpt-4o-mini", 0.0, 150)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	return strings.TrimSpace(summary), nil
}
