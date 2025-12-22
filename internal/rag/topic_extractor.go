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

// =============================================================================
// LLM PROMPTS FOR TOPIC EXTRACTION - ALWAYS IN ENGLISH
// Language-specific output is controlled via language directive in the prompt
// =============================================================================

// topicExtractionSystemPrompt is the system prompt for capability extraction.
const topicExtractionSystemPrompt = "You are a helpful assistant."

// topicExtractionUserPromptTemplate is for simple capability extraction.
// %s placeholder is for the target language name.
const topicExtractionUserPromptTemplate = `The text above is a knowledge source for a chatbot. Write a single sentence summarizing what capabilities or information this text provides to the chatbot.
Only rely on the information present in the text. If the text is meaningless or contains no information, state that.
Example: "Provides information about the company's history and vision."

IMPORTANT: Write the summary in %s.
Summary:`

// topicExtractionJSONPromptTemplate is for structured metadata extraction.
// %s placeholder is for the target language name.
// topicExtractionJSONPromptTemplate is for structured metadata extraction.
// First %d is for max questions, second %s is for language name.
const topicExtractionJSONPromptTemplate = `The text above is a knowledge source for a chatbot. 
Write a single sentence summarizing what capabilities or information this text provides to the chatbot.
Only rely on the information present in the text. If the text is meaningless or contains no information, state that.
Example: "Provides information about the company's history and vision."

Respond ONLY in JSON format:
{
  "capability_summary": <short sentence>,
  "suggested_questions": [<%d short and varied questions>]
}
No extra explanation or text. Write the summary and questions in %s.`

// DefaultMaxSuggestedQuestions is the fallback when plan limit is not specified.
const DefaultMaxSuggestedQuestions = 6

// extractTopicsFallback is used as a fallback when structured metadata extraction fails.
func extractTopicsFallback(ctx context.Context, client LLMClient, content string, langCode string) (string, error) {
	if len(content) > 2000 {
		content = content[:2000]
	}
	cfg := langconfig.Get(langCode)
	sp := topicExtractionSystemPrompt
	ct := fmt.Sprintf("Text:\n%s", content)
	um := fmt.Sprintf(topicExtractionUserPromptTemplate, cfg.Name)

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
// maxQuestions specifies the plan-based limit for suggested questions.
func ExtractIngestionMetadata(ctx context.Context, client LLMClient, content string, langCode string, maxQuestions int) (models.IngestionMetadata, error) {
	if len(content) > 4000 {
		content = content[:4000]
	}
	if maxQuestions <= 0 {
		maxQuestions = DefaultMaxSuggestedQuestions
	}
	cfg := langconfig.Get(langCode)
	sp := topicExtractionSystemPrompt
	ct := fmt.Sprintf("Text:\n%s", content)
	// English prompt with language directive for output
	um := fmt.Sprintf(topicExtractionJSONPromptTemplate, maxQuestions, cfg.Name)

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
		sum, serr := extractTopicsFallback(ctx, client, content, langCode)
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
			if sum, serr := extractTopicsFallback(ctx, client, content, langCode); serr == nil {
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
