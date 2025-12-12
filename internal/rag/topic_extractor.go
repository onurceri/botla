package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/config"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

// ExtractTopics remains for backward compatibility.
func ExtractTopics(ctx context.Context, client LLMClient, content string, langCode string) (string, error) {
	if len(content) > 2000 {
		content = content[:2000]
	}
	cfg := langconfig.Get(langCode)
	sp := cfg.ResponseTemplates.TopicExtractionSystemPrompt
	ct := fmt.Sprintf("Metin:\n%s", content)
	um := cfg.ResponseTemplates.TopicExtractionUserPrompt

	params := models.CompletionParams{
		SystemPrompt: sp,
		Context:      ct,
		UserMessage:  um,
		Model:        config.ModelGPT4oMini,
		Temperature:  0.0,
		MaxTokens:    150,
	}

	result, err := client.CreateCompletion(ctx, params)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}
	return strings.TrimSpace(result.Content), nil
}

var fenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)\\s*```")

// ExtractIngestionMetadata asks the LLM for structured output with capability summary and example questions.
func ExtractIngestionMetadata(ctx context.Context, client LLMClient, content string, langCode string) (models.IngestionMetadata, error) {
	if len(content) > 4000 {
		content = content[:4000]
	}
	cfg := langconfig.Get(langCode)
	sp := cfg.ResponseTemplates.TopicExtractionSystemPrompt
	ct := fmt.Sprintf("Metin:\n%s", content)
	// Strict JSON instruction; model will often still wrap in fences — we handle both.
	um := cfg.ResponseTemplates.TopicExtractionUserPrompt + "\n\nYanıtı YALNIZCA JSON olarak ver. Şu formatta:\n{\n  \"capability_summary\": <kısa cümle>,\n  \"suggested_questions\": [<3-6 kısa ve farklı soru>]\n}\nEkstra açıklama, ön/son metin ekleme. Soruları \"" + cfg.Code + "\" dilinde yaz."

	params := models.CompletionParams{
		SystemPrompt: sp,
		Context:      ct,
		UserMessage:  um,
		Model:        config.ModelGPT4oMini,
		Temperature:  0.0,
		MaxTokens:    300,
	}

	result, err := client.CreateCompletion(ctx, params)
	if err != nil {
		return models.IngestionMetadata{}, fmt.Errorf("llm call failed: %w", err)
	}
	out := result.Content
	raw := strings.TrimSpace(out)
	// If fenced code, extract inner JSON
	if m := fenceRe.FindStringSubmatch(raw); len(m) == 2 {
		raw = strings.TrimSpace(m[1])
	}
	var im models.IngestionMetadata
	if jerr := json.Unmarshal([]byte(raw), &im); jerr != nil {
		// Fallback: derive minimal metadata from legacy summary
		sum, serr := ExtractTopics(ctx, client, content, langCode)
		if serr != nil {
			return models.IngestionMetadata{}, errors.New("failed to parse and fallback summary")
		}
		im = models.IngestionMetadata{CapabilitySummary: strings.TrimSpace(sum), SuggestedQuestions: deriveQuestionsFromSummary(sum, cfg.Code)}
	}

	// If JSON parsed but questions are empty, try to derive from summary
	if len(im.SuggestedQuestions) == 0 {
		if im.CapabilitySummary != "" {
			im.SuggestedQuestions = deriveQuestionsFromSummary(im.CapabilitySummary, cfg.Code)
		} else {
			// Last resort: try legacy extraction
			if sum, serr := ExtractTopics(ctx, client, content, langCode); serr == nil {
				im.CapabilitySummary = strings.TrimSpace(sum)
				im.SuggestedQuestions = deriveQuestionsFromSummary(sum, cfg.Code)
			}
		}
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
