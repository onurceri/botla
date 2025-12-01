package rag

import (
    "context"
    "sort"
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
    items, err := qc.SearchSimilar(ctx, queryEmbedding, chatbotID, 5)
    if err != nil {
        return "", nil, err
    }
    var metas []ChunkMetadata
    for _, it := range items {
        metas = append(metas, ChunkMetadata{SourceID: it.Payload.SourceID, SourceType: it.Payload.SourceType, ChunkIndex: it.Payload.ChunkIndex, Score: it.Score})
    }
    sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })
    const threshold = 0.2
    var parts []string
    var used []ChunkMetadata
    var tokens int
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
        nt := CountTokens(next)
        if tokens+nt > 2000 {
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

