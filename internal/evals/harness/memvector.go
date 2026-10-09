// Package harness drives Botla's real RAG code paths for evaluation:
// ingestion (normalize -> rag.ChunkText -> rag.EmbeddingService), retrieval
// (VectorClient.SearchSimilar + rag.SearchContextTiered) and answer
// generation (services.ChatService.ProcessChat, the same entry point the HTTP
// chat handlers use). Only persistence (Postgres repositories) is replaced by
// in-memory fakes; prompts, thresholds, tools and fallbacks are the product's.
package harness

import (
	"context"
	"errors"
	"math"
	"sort"
	"sync"

	"github.com/onurceri/botla-app/internal/rag"
)

// MemoryVectorStore is an exact (brute-force cosine) rag.VectorClient. It is
// the seam that lets the suite run without Qdrant (unit tests, quick local
// runs). Qdrant's HNSW search is approximate, so on large collections results
// may differ slightly; at eval-corpus size both are effectively exact.
type MemoryVectorStore struct {
	mu     sync.RWMutex
	points map[string]memPoint
}

type memPoint struct {
	vector  []float32
	norm    float64
	payload rag.EmbeddingPayload
}

// NewMemoryVectorStore returns an empty store.
func NewMemoryVectorStore() *MemoryVectorStore {
	return &MemoryVectorStore{points: map[string]memPoint{}}
}

var _ rag.VectorClient = (*MemoryVectorStore)(nil)

// EnsureEmbeddingsCollection is a no-op for the in-memory store.
func (m *MemoryVectorStore) EnsureEmbeddingsCollection(context.Context) error { return nil }

// UpsertEmbedding stores or replaces a point.
func (m *MemoryVectorStore) UpsertEmbedding(_ context.Context, id interface{}, vector []float32, payload rag.EmbeddingPayload) error {
	key, ok := id.(string)
	if !ok || key == "" {
		return errors.New("memory vector store: point id must be a non-empty string")
	}
	v := append([]float32(nil), vector...)
	m.mu.Lock()
	m.points[key] = memPoint{vector: v, norm: norm(v), payload: payload}
	m.mu.Unlock()
	return nil
}

// SearchSimilar returns the topK points of the chatbot by cosine similarity.
func (m *MemoryVectorStore) SearchSimilar(_ context.Context, embedding []float32, chatbotID string, topK int) ([]rag.SearchResult, error) {
	qn := norm(embedding)
	m.mu.RLock()
	out := make([]rag.SearchResult, 0, len(m.points))
	for id, p := range m.points {
		if p.payload.ChatbotID != chatbotID {
			continue
		}
		out = append(out, rag.SearchResult{ID: id, Score: cosine(embedding, qn, p.vector, p.norm), Payload: p.payload})
	}
	m.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].ID.(string) < out[j].ID.(string) // deterministic tie-break
	})
	if topK > 0 && len(out) > topK {
		out = out[:topK]
	}
	return out, nil
}

// DeleteBySourceID removes every point of a source.
func (m *MemoryVectorStore) DeleteBySourceID(_ context.Context, sourceID string) error {
	m.mu.Lock()
	for id, p := range m.points {
		if p.payload.SourceID == sourceID {
			delete(m.points, id)
		}
	}
	m.mu.Unlock()
	return nil
}

// ScrollChunks returns the points of a source ordered by chunk index. The
// in-memory store never paginates, so the next offset is always nil.
func (m *MemoryVectorStore) ScrollChunks(_ context.Context, sourceID string, limit int, _ interface{}) ([]rag.SearchResult, *string, error) {
	m.mu.RLock()
	var out []rag.SearchResult
	for id, p := range m.points {
		if p.payload.SourceID == sourceID {
			out = append(out, rag.SearchResult{ID: id, Payload: p.payload})
		}
	}
	m.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool { return out[i].Payload.ChunkIndex < out[j].Payload.ChunkIndex })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil, nil
}

// Len returns the number of stored points.
func (m *MemoryVectorStore) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.points)
}

func norm(v []float32) float64 {
	var s float64
	for _, x := range v {
		s += float64(x) * float64(x)
	}
	return math.Sqrt(s)
}

func cosine(a []float32, an float64, b []float32, bn float64) float64 {
	if an == 0 || bn == 0 || len(a) != len(b) {
		return 0
	}
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot / (an * bn)
}
