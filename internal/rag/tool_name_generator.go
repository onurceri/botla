package rag

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ToolNameGenerator generates API-compatible tool names using LLM
type ToolNameGenerator struct {
	client ToolsLLMClient
}

// NewToolNameGenerator creates a new ToolNameGenerator
func NewToolNameGenerator(client ToolsLLMClient) *ToolNameGenerator {
	return &ToolNameGenerator{client: client}
}

const toolNameGeneratorPrompt = `Generate a valid function/tool name for an AI assistant based on the following action.

Action Name: %s
Action Description: %s

Requirements:
- MUST match regex: ^[a-z][a-z0-9_]{0,62}[a-z0-9]$ (lowercase letters, numbers, underscores only)
- Use snake_case format (e.g., get_weather, create_order, check_status)
- Be descriptive but concise (2-4 words preferred)
- Use English even if input is in another language
- Start with a verb when possible (get_, create_, update_, search_, check_, send_, fetch_, list_)
- Maximum 64 characters

Examples:
- "Hava durumunu sorgula" + "Şehir için hava durumu" → get_weather
- "Sipariş durumu" + "Sipariş numarasına göre sorgula" → check_order_status
- "E-posta gönder" + "Müşteriye e-posta gönderir" → send_email

Return ONLY the tool name, nothing else. No quotes, no explanation.`

// Generate creates an API-compatible tool name from action name and description
func (g *ToolNameGenerator) Generate(ctx context.Context, name, description string) (string, error) {
	prompt := fmt.Sprintf(toolNameGeneratorPrompt, name, description)
	msg := prompt

	messages := []ChatMessage{
		{Role: "user", Content: &msg},
	}

	resp, err := g.client.CreateCompletionWithTools(ctx, messages, nil, "gpt-4o-mini", 0.1, 50)
	if err != nil {
		return "", fmt.Errorf("failed to generate tool name: %w", err)
	}

	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == nil {
		return "", fmt.Errorf("no tool name generated")
	}

	toolName := strings.TrimSpace(*resp.Choices[0].Message.Content)
	toolName = strings.Trim(toolName, "\"'`")
	toolName = strings.ToLower(toolName)

	// Validate the generated tool name
	if !IsValidToolName(toolName) {
		return "", fmt.Errorf("LLM generated invalid tool name: %q (must match ^[a-zA-Z0-9_-]{1,64}$)", toolName)
	}

	return toolName, nil
}

// ValidateToolName checks if a tool name is valid for OpenAI function calling
func ValidateToolName(name string) error {
	if name == "" {
		return fmt.Errorf("tool name cannot be empty")
	}
	if len(name) > 64 {
		return fmt.Errorf("tool name must be at most 64 characters")
	}
	if !IsValidToolName(name) {
		return fmt.Errorf("tool name must match ^[a-zA-Z0-9_-]{1,64}$")
	}
	return nil
}

// toolNameCleanupRE matches invalid characters
var toolNameCleanupRE = regexp.MustCompile(`[^a-z0-9_]`)

// SanitizeToolName attempts to clean up a tool name to make it valid
// This is a fallback helper, not meant for primary use
func SanitizeToolName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "_")
	s = toolNameCleanupRE.ReplaceAllString(s, "")
	if len(s) > 64 {
		s = s[:64]
	}
	if s == "" {
		return "action"
	}
	return s
}
