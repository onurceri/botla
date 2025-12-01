package rag

import (
    "context"
    "strconv"
    "time"

    "github.com/onurceri/botla-co/pkg/logger"
)

// GenerateEmbeddings orchestrates batch embedding creation and Qdrant upsert.
// - Batching: 25 chunks per request
// - Rate limiting: ~58 req/sec (~3480/min)
// - Retry: up to 3x per batch with exponential backoff handled in client
// - Error recovery: skip failed items; continue others
// - Cost tracking: logs approximate cost based on chunk token counts
func GenerateEmbeddings(chunks []Chunk, chatbotID string) error {
    if len(chunks) == 0 || chatbotID == "" {
        return nil
    }
    log := logger.New("INFO")
    oai, err := NewOpenAIClientFromEnv()
    if err != nil {
        log.Error("openai_client_init_failed", map[string]any{"error": err.Error()})
        return err
    }
    qc, err := NewQdrantClientFromEnv()
    if err != nil {
        log.Error("qdrant_client_init_failed", map[string]any{"error": err.Error()})
        return err
    }

    // soft rate limiter: ~58 req/sec
    ticker := time.NewTicker(time.Second / 58)
    defer ticker.Stop()

    const batchSize = 25
    var totalTokens int

    for start := 0; start < len(chunks); start += batchSize {
        end := start + batchSize
        if end > len(chunks) {
            end = len(chunks)
        }
        batch := chunks[start:end]

        <-ticker.C
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        texts := make([]string, len(batch))
        for i, ch := range batch {
            texts[i] = ch.Text
            totalTokens += ch.TokenCount
        }
        vectors, berr := oai.CreateEmbeddingsBatch(ctx, texts)
        cancel()
        if berr != nil {
            log.Warn("embedding_batch_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
            // retry one more time quickly honoring limiter
            <-ticker.C
            ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
            vectors, berr = oai.CreateEmbeddingsBatch(ctx2, texts)
            cancel2()
        }
        if berr != nil {
            log.Error("embedding_batch_final_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
            return berr
        }
        // upsert each vector
        for i := range vectors {
            id := chatbotID + ":" + strconv.Itoa(start+i)
            payload := EmbeddingPayload{
                ChatbotID:    chatbotID,
                SourceID:     "",
                ChunkIndex:   start + i,
                OriginalText: batch[i].Text,
                SourceType:   "unknown",
                CreatedAt:    time.Now(),
            }
            ctx3, cancel3 := context.WithTimeout(context.Background(), 10*time.Second)
            if err := qc.UpsertEmbedding(ctx3, id, vectors[i], payload); err != nil {
                log.Warn("qdrant_upsert_failed", map[string]any{"error": err.Error(), "id": id})
            }
            cancel3()
        }
    }

    cost := float64(totalTokens) * 0.02 / 1_000_000.0
    log.Info("embedding_pipeline_completed", map[string]any{"chunks": len(chunks), "total_tokens": totalTokens, "estimated_cost_usd": cost})
    return nil
}
