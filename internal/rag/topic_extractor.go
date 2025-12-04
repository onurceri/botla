package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/onurceri/botla-co/pkg/langconfig"
)

// LLMClient defines the interface for interacting with the LLM provider.
type LLMClient interface {
	CreateCompletion(ctx context.Context, systemPrompt, contextText, userMessage string, model string, temperature float32, maxTokens int) (string, int, error)
}

// ExtractTopics remains for backward compatibility.
func ExtractTopics(ctx context.Context, client LLMClient, content string, langCode string) (string, error) {
	if len(content) > 2000 {
		content = content[:2000]
	}
	cfg := langconfig.Get(langCode)
	sp := cfg.ResponseTemplates.TopicExtractionSystemPrompt
	ct := fmt.Sprintf("Metin:\n%s", content)
	um := cfg.ResponseTemplates.TopicExtractionUserPrompt
	summary, _, err := client.CreateCompletion(ctx, sp, ct, um, "gpt-4o-mini", 0.0, 150)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	return strings.TrimSpace(summary), nil
}

type IngestionMetadata struct {
	CapabilitySummary  string   `json:"capability_summary"`
	SuggestedQuestions []string `json:"suggested_questions"`
}

var fenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")

// ExtractIngestionMetadata asks the LLM for structured output with capability summary and example questions.
func ExtractIngestionMetadata(ctx context.Context, client LLMClient, content string, langCode string) (IngestionMetadata, error) {
	if len(content) > 4000 {
		content = content[:4000]
	}
	cfg := langconfig.Get(langCode)
	sp := cfg.ResponseTemplates.TopicExtractionSystemPrompt
	ct := fmt.Sprintf("Metin:\n%s", content)
	// Strict JSON instruction; model will often still wrap in fences — we handle both.
	um := cfg.ResponseTemplates.TopicExtractionUserPrompt + "\n\nYanıtı YALNIZCA JSON olarak ver. Şu formatta:\n{\n  \"capability_summary\": <kısa cümle>,\n  \"suggested_questions\": [<3-6 kısa ve farklı soru>]\n}\nEkstra açıklama, ön/son metin ekleme. Soruları \"" + cfg.Code + "\" dilinde yaz."

	out, _, err := client.CreateCompletion(ctx, sp, ct, um, "gpt-4o-mini", 0.0, 300)
	if err != nil {
		return IngestionMetadata{}, fmt.Errorf("llm call failed: %w", err)
	}
	raw := strings.TrimSpace(out)
	// If fenced code, extract inner JSON
	if m := fenceRe.FindStringSubmatch(raw); len(m) == 2 {
		raw = strings.TrimSpace(m[1])
	}
	var im IngestionMetadata
	if jerr := json.Unmarshal([]byte(raw), &im); jerr != nil {
		// Fallback: derive minimal metadata from legacy summary
		sum, serr := ExtractTopics(ctx, client, content, langCode)
		if serr != nil {
			return IngestionMetadata{}, errors.New("failed to parse and fallback summary")
		}
		im = IngestionMetadata{CapabilitySummary: strings.TrimSpace(sum), SuggestedQuestions: deriveQuestionsFromSummary(sum, cfg.Code)}
	}
	// Normalize questions: trim, drop empties, cap count/length
	im.SuggestedQuestions = normalizeSuggestions(im.SuggestedQuestions)
	return im, nil
}

func deriveQuestionsFromSummary(sum string, lang string) []string {
	s := strings.TrimSpace(sum)
	if s == "" {
		return []string{}
	}
	if lang == "en" {
		return normalizeSuggestions([]string{
			"What topics can you help me with?",
			"Give me a quick overview.",
			"Where should I start?",
		})
	}
	return normalizeSuggestions([]string{
		"Hangi konularda yardımcı olabilirsin?",
		"Kısa bir genel bakış verir misin?",
		"Nereden başlamalıyım?",
	})
}

func normalizeSuggestions(in []string) []string {
	if len(in) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, q := range in {
		t := strings.TrimSpace(q)
		if t == "" {
			continue
		}
		if len(t) > 120 {
			t = t[:120]
		}
		k := strings.ToLower(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
		if len(out) >= 6 {
			break
		}
	}
	return out
}
