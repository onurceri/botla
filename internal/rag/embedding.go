package rag

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/logger"
)

// GenerateEmbeddings orchestrates batch embedding creation and Qdrant upsert.
// - Batching: 25 chunks per request
// - Rate limiting: ~58 req/sec (~3480/min)
// - Retry: up to 3x per batch with exponential backoff handled in client
// - Error recovery: skip failed items; continue others
// - Cost tracking: logs approximate cost based on chunk token counts
func GenerateEmbeddings(ctx context.Context, emb EmbeddingClient, vc VectorClient, chunks []models.Chunk, chatbotID string) error {
	if len(chunks) == 0 || chatbotID == "" {
		return nil
	}
	log := logger.New("INFO")

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
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		texts := make([]string, len(batch))
		for i, ch := range batch {
			texts[i] = ch.Text
			totalTokens += ch.TokenCount
		}
		vectors, berr := emb.CreateEmbeddingsBatch(batchCtx, texts)
		cancel()
		if berr != nil {
			log.Warn("embedding_batch_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
			// retry one more time quickly honoring limiter
			<-ticker.C
			batchCtx2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
			vectors, berr = emb.CreateEmbeddingsBatch(batchCtx2, texts)
			cancel2()
		}
		if berr != nil {
			log.Error("embedding_batch_final_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
			return fmt.Errorf("create embeddings batch: %w", berr)
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
			batchCtx3, cancel3 := context.WithTimeout(ctx, 10*time.Second)
			if err := vc.UpsertEmbedding(batchCtx3, id, vectors[i], payload); err != nil {
				log.Warn("qdrant_upsert_failed", map[string]any{"error": err.Error(), "id": id})
			}
			cancel3()
		}
	}

	cost := float64(totalTokens) * 0.02 / 1_000_000.0
	log.Info("embedding_pipeline_completed", map[string]any{"chunks": len(chunks), "total_tokens": totalTokens, "estimated_cost_usd": cost})
	return nil
}

func GenerateEmbeddingsForSource(ctx context.Context, emb EmbeddingClient, vc VectorClient, chunks []models.Chunk, chatbotID, sourceID, sourceType string) error {
	if len(chunks) == 0 || chatbotID == "" || sourceID == "" {
		return nil
	}
	ticker := time.NewTicker(time.Second / 58)
	defer ticker.Stop()
	const batchSize = 25
	for start := 0; start < len(chunks); start += batchSize {
		end := start + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[start:end]
		<-ticker.C
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		texts := make([]string, len(batch))
		for i, ch := range batch {
			texts[i] = ch.Text
		}
		vectors, berr := emb.CreateEmbeddingsBatch(batchCtx, texts)
		cancel()
		if berr != nil {
			<-ticker.C
			batchCtx2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
			vectors, berr = emb.CreateEmbeddingsBatch(batchCtx2, texts)
			cancel2()
		}
		if berr != nil {
			return fmt.Errorf("create embeddings batch: %w", berr)
		}
		for i := range vectors {
			pid := MakePointID(sourceID, start+i)
			payload := EmbeddingPayload{ChatbotID: chatbotID, SourceID: sourceID, ChunkIndex: start + i, OriginalText: batch[i].Text, SourceType: sourceType, CreatedAt: time.Now()}
			batchCtx3, cancel3 := context.WithTimeout(ctx, 10*time.Second)
			if err := vc.UpsertEmbedding(batchCtx3, pid, vectors[i], payload); err != nil {
				cancel3()
				return fmt.Errorf("upsert embedding: %w", err)
			}
			cancel3()
		}
	}
	return nil
}

func MakePointID(sourceID string, index int) string {
	s := sourceID + ":" + strconv.Itoa(index)
	h := sha256.Sum256([]byte(s))
	h[6] = (h[6] & 0x0f) | 0x30
	h[8] = (h[8] & 0x3f) | 0x80
	u := h[:16]
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}
