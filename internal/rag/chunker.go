package rag

import (
    "errors"
    "math"
	"os"
    "regexp"
    "strings"

	"github.com/neurosnap/sentences"
	"github.com/neurosnap/sentences/english"
	"github.com/onurceri/botla-co/pkg/langconfig"
)

type Chunk struct {
    Text       string
    TokenCount int
}

// ChunkText splits raw text into token-aware chunks preserving paragraph and sentence boundaries.
// targetTokens defines desired max tokens per chunk; chunks include ~15% tail overlap from previous chunk.
func ChunkText(text string, targetTokens int, langCode string) ([]Chunk, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return nil, nil
	}
	if targetTokens <= 0 {
		return nil, errors.New("targetTokens must be > 0")
	}

	paras := splitParagraphs(s)
	var chunks []Chunk
	var current []string
	var prevTail []string

	flush := func(force bool) {
		if len(current) == 0 {
			return
		}
		joined := joinSentences(current)
		tokens := CountTokens(joined, langCode)
		if tokens == 0 && !force {
			return
		}
		chunks = append(chunks, Chunk{Text: joined, TokenCount: tokens})
		// compute tail ~15% of chunk tokens using sentence boundaries
		tailTokens := int(math.Round(float64(tokens) * 0.15))
		prevTail = tailFromSentences(current, tailTokens, langCode)
		current = nil
	}

	for _, p := range paras {
		sentences := splitSentences(p, langCode)
		for i := 0; i < len(sentences); i++ {
			if len(current) == 0 && len(prevTail) > 0 {
				current = append(current, prevTail...)
			}
			// try to add next sentence
			candidate := append(current, sentences[i])
			candText := joinSentences(candidate)
			candTokens := CountTokens(candText, langCode)
			if candTokens <= targetTokens || len(current) == 0 {
				current = candidate
				continue
			}
			// would exceed: flush current, start new chunk seeded with tail
			flush(false)
			// after flush, seed new chunk with previous tail and reprocess same sentence
			if len(prevTail) > 0 {
				current = append([]string{}, prevTail...)
				cand := append(current, sentences[i])
				if CountTokens(joinSentences(cand), langCode) <= targetTokens || len(prevTail) == 0 {
					current = cand
				} else {
					// very long sentence: emit alone
					current = append([]string{}, sentences[i])
				}
			} else {
				// no tail, start with the long sentence
				current = append([]string{}, sentences[i])
			}
		}
	}
	if len(current) > 0 {
		flush(true)
	}
	return chunks, nil
}

func splitParagraphs(s string) []string {
	re := regexp.MustCompile(`\n{2,}`)
	parts := re.Split(s, -1)
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func joinSentences(sents []string) string {
	return strings.TrimSpace(strings.Join(sents, " "))
}

func tailFromSentences(sents []string, target int, langCode string) []string {
	if target <= 0 || len(sents) == 0 {
		return nil
	}
	var acc []string
	var tok int
	for i := len(sents) - 1; i >= 0; i-- {
		acc = append([]string{sents[i]}, acc...)
		tok = CountTokens(joinSentences(acc), langCode)
		if tok >= target {
			break
		}
	}
	return acc
}

// splitSentences splits sentences using a hybrid approach:
// 1. Tries to load a pre-trained Punkt model from langconfig.TokenizerData.
// 2. Applies a "patch" layer for hardcoded abbreviations from langconfig.Abbreviations.
// 3. Falls back to English tokenizer or Regex if model loading fails.
func splitSentences(text string, langCode string) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	cfg := langconfig.Get(langCode)

	// protect common abbreviations to avoid splitting on period
	// We map abbreviations to a temporary placeholder
	repl := text
	protected := make(map[string]string)
	for i, abbr := range cfg.Abbreviations {
		placeholder := "ABBR_" + string(rune('A'+i)) + "_"
		protected[placeholder] = abbr
		repl = strings.ReplaceAll(repl, abbr, placeholder)
	}

	var matches []string
	var tok sentences.SentenceTokenizer
	var err error

	// Try to load trained model
	if cfg.TokenizerData != "" {
		// In a real app, we should cache the tokenizer instance
		// For now, we load it per request (optimization needed later)
		// Read the file content
		content, rerr := os.ReadFile(cfg.TokenizerData)
		if rerr == nil {
			data, berr := sentences.LoadTraining(content)
			if berr == nil {
				tok = sentences.NewSentenceTokenizer(data)
			}
		}
	}

	// Fallback to English tokenizer if no specific model loaded
	if tok == nil {
		tok, err = english.NewSentenceTokenizer(nil)
	}

	if tok != nil && err == nil {
		ss := tok.Tokenize(repl)
		for _, s := range ss {
			matches = append(matches, s.Text)
		}
	}

	// fallback to regex if tokenizer fails
	if len(matches) == 0 {
		re := regexp.MustCompile(`(?m)([^.!?…]+[.!?…]+(?:[\)"]+)?)`)
		matches = re.FindAllString(repl, -1)
	}
	if len(matches) == 0 {
		matches = []string{repl}
	}
	var out []string
	for _, m := range matches {
		s := strings.TrimSpace(m)
		if s == "" {
			continue
		}
		// restore abbreviations
		for k, v := range protected {
			s = strings.ReplaceAll(s, k, v)
		}
		out = append(out, s)
	}
	return out
}

