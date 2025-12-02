package rag

import (
	"context"
	"os"
	"sort"
	"strconv"
	"strings"
)

type ChunkMetadata struct {
	SourceID   string
	SourceType string
	ChunkIndex int
	Score      float64
}

func SearchContext(queryEmbedding []float32, chatbotID string) (string, []ChunkMetadata, error) {
	if len(queryEmbedding) == 0 || chatbotID == "" {
		return "", nil, nil
	}
	qc, err := NewQdrantClientFromEnv()
	if err != nil {
		return "", nil, err
	}
	ctx := context.Background()
	topK := 5
	if v := os.Getenv("RAG_TOPK"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			topK = n
		}
	}
	items, err := qc.SearchSimilar(ctx, queryEmbedding, chatbotID, topK)
	if err != nil {
		return "", nil, err
	}
	var metas []ChunkMetadata
	for _, it := range items {
		metas = append(metas, ChunkMetadata{SourceID: it.Payload.SourceID, SourceType: it.Payload.SourceType, ChunkIndex: it.Payload.ChunkIndex, Score: it.Score})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
	threshold := 0.2
	if v := os.Getenv("RAG_SCORE_THRESHOLD"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			threshold = f
		}
	}
	var parts []string
	var used []ChunkMetadata
	var tokens int
	maxCtx := 2000
	if v := os.Getenv("RAG_MAX_CONTEXT_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxCtx = n
		}
	}
	for _, it := range items {
		if it.Score < threshold {
			continue
		}
		t := strings.TrimSpace(it.Payload.OriginalText)
		if t == "" {
			continue
		}
		next := t
		if len(parts) > 0 {
			next = "\n---\n" + next
		}
		nt := CountTokens(next, "tr") // Default to TR for search context estimation if language unknown
		if tokens+nt > maxCtx {
			break
		}
		parts = append(parts, next)
		tokens += nt
		used = append(used, ChunkMetadata{SourceID: it.Payload.SourceID, SourceType: it.Payload.SourceType, ChunkIndex: it.Payload.ChunkIndex, Score: it.Score})
	}
	if len(parts) == 0 {
		return "", metas, nil
	}
	body := strings.Join(parts, "")
	formatted := "Aşağıdaki belgeler sorgularına cevap vermek için kullanılmıştır:\n\n" + body
	return formatted, used, nil
}
