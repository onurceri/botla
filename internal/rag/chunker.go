package rag

import (
    "errors"
    "math"
    "regexp"
    "strings"

    "github.com/neurosnap/sentences/english"
)

type Chunk struct {
    Text       string
    TokenCount int
}

// ChunkText splits raw text into token-aware chunks preserving paragraph and sentence boundaries.
// targetTokens defines desired max tokens per chunk; chunks include ~15% tail overlap from previous chunk.
func ChunkText(text string, targetTokens int) ([]Chunk, error) {
    s := strings.TrimSpace(text)
    if s == "" {
        return nil, nil
    }
    if targetTokens <= 0 {
        return nil, errors.New("targetTokens must be > 0")
    }

    paras := splitParagraphsTR(s)
    var chunks []Chunk
    var current []string
    var prevTail []string

    flush := func(force bool) {
        if len(current) == 0 {
            return
        }
        joined := joinSentences(current)
        tokens := CountTokens(joined)
        if tokens == 0 && !force {
            return
        }
        chunks = append(chunks, Chunk{Text: joined, TokenCount: tokens})
        // compute tail ~15% of chunk tokens using sentence boundaries
        tailTokens := int(math.Round(float64(tokens) * 0.15))
        prevTail = tailFromSentences(current, tailTokens)
        current = nil
    }

    for _, p := range paras {
        sentences := splitSentencesTR(p)
        for i := 0; i < len(sentences); i++ {
            if len(current) == 0 && len(prevTail) > 0 {
                current = append(current, prevTail...)
            }
            // try to add next sentence
            candidate := append(current, sentences[i])
            candText := joinSentences(candidate)
            candTokens := CountTokens(candText)
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
                if CountTokens(joinSentences(cand)) <= targetTokens || len(prevTail) == 0 {
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

func splitParagraphsTR(s string) []string {
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

func tailFromSentences(sents []string, target int) []string {
    if target <= 0 || len(sents) == 0 {
        return nil
    }
    var acc []string
    var tok int
    for i := len(sents) - 1; i >= 0; i-- {
        acc = append([]string{sents[i]}, acc...)
        tok = CountTokens(joinSentences(acc))
        if tok >= target {
            break
        }
    }
    return acc
}

// splitSentencesTR splits sentences with simple heuristics and Turkish abbreviations handling.
func splitSentencesTR(text string) []string {
    if strings.TrimSpace(text) == "" {
        return nil
    }
    // protect common Turkish abbreviations to avoid splitting on period
    protected := map[string]string{
        "Dr.":   "DR_ABBR",
        "Prof.": "PROF_ABBR",
        "vb.":   "VB_ABBR",
    }
    repl := text
    for k, v := range protected {
        repl = strings.ReplaceAll(repl, k, v)
    }
    // use neurosnap/sentences (english tokenizer works reasonably for punctuation)
    tok, err := english.NewSentenceTokenizer(nil)
    var matches []string
    if err == nil && tok != nil {
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
            s = strings.ReplaceAll(s, v, k)
        }
        out = append(out, s)
    }
    return out
}
