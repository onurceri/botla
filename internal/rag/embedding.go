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

// EmbeddingService handles batch embedding creation and vector storage.
// It provides rate limiting, retry logic, and cost tracking.
type EmbeddingService struct {
	embedder EmbeddingClient
	vector   VectorClient
	log      *logger.Logger
}

// NewEmbeddingService creates a new EmbeddingService with the given clients.
func NewEmbeddingService(embedder EmbeddingClient, vector VectorClient, log *logger.Logger) *EmbeddingService {
	if log == nil {
		log = logger.New("INFO")
	}
	return &EmbeddingService{
		embedder: embedder,
		vector:   vector,
		log:      log,
	}
}

// Generate orchestrates batch embedding creation and Qdrant upsert.
// - Batching: 25 chunks per request
// - Rate limiting: ~58 req/sec (~3480/min)
// - Retry: up to 2x per batch
// - Error recovery: returns on first failure
// - Cost tracking: logs approximate cost based on chunk token counts
func (s *EmbeddingService) Generate(ctx context.Context, chunks []models.Chunk, chatbotID string) error {
	if len(chunks) == 0 || chatbotID == "" {
		return nil
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
		batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		texts := make([]string, len(batch))
		for i, ch := range batch {
			texts[i] = ch.Text
			totalTokens += ch.TokenCount
		}
		vectors, berr := s.embedder.CreateEmbeddingsBatch(batchCtx, texts)
		cancel()
		if berr != nil {
			s.log.Warn("embedding_batch_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
			// retry one more time quickly honoring limiter
			<-ticker.C
			batchCtx2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
			vectors, berr = s.embedder.CreateEmbeddingsBatch(batchCtx2, texts)
			cancel2()
		}
		if berr != nil {
			s.log.Error("embedding_batch_final_failed", map[string]any{"error": berr.Error(), "start_index": start, "count": len(batch)})
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
			if err := s.vector.UpsertEmbedding(batchCtx3, id, vectors[i], payload); err != nil {
				s.log.Warn("qdrant_upsert_failed", map[string]any{"error": err.Error(), "id": id})
			}
			cancel3()
		}
	}

	cost := float64(totalTokens) * 0.02 / 1_000_000.0
	s.log.Info("embedding_pipeline_completed", map[string]any{"chunks": len(chunks), "total_tokens": totalTokens, "estimated_cost_usd": cost})
	return nil
}

// GenerateForSource generates embeddings for a specific source.
// Each embedding includes source metadata for filtering and deletion.
func (s *EmbeddingService) GenerateForSource(ctx context.Context, chunks []models.Chunk, chatbotID, sourceID, sourceType string) error {
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
		vectors, berr := s.embedder.CreateEmbeddingsBatch(batchCtx, texts)
		cancel()
		if berr != nil {
			<-ticker.C
			batchCtx2, cancel2 := context.WithTimeout(ctx, 30*time.Second)
			vectors, berr = s.embedder.CreateEmbeddingsBatch(batchCtx2, texts)
			cancel2()
		}
		if berr != nil {
			return fmt.Errorf("create embeddings batch: %w", berr)
		}
		for i := range vectors {
			pid := MakePointID(sourceID, start+i)
			payload := EmbeddingPayload{ChatbotID: chatbotID, SourceID: sourceID, ChunkIndex: start + i, OriginalText: batch[i].Text, SourceType: sourceType, CreatedAt: time.Now()}
			batchCtx3, cancel3 := context.WithTimeout(ctx, 10*time.Second)
			if err := s.vector.UpsertEmbedding(batchCtx3, pid, vectors[i], payload); err != nil {
				cancel3()
				return fmt.Errorf("upsert embedding: %w", err)
			}
			cancel3()
		}
	}
	return nil
}

// MakePointID creates a deterministic UUID-like ID from sourceID and chunk index.
func MakePointID(sourceID string, index int) string {
	s := sourceID + ":" + strconv.Itoa(index)
	h := sha256.Sum256([]byte(s))
	h[6] = (h[6] & 0x0f) | 0x30
	h[8] = (h[8] & 0x3f) | 0x80
	u := h[:16]
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}
