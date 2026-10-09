package harness

import (
	"context"
	"fmt"

	"github.com/onurceri/botla-app/internal/evals/cache"
	"github.com/onurceri/botla-app/internal/rag"
)

const embeddingNamespace = "embeddings"

// CachingEmbedder wraps the product embedding client with a content-hash
// cache. Embeddings for a fixed corpus and question set are deterministic, so
// re-runs (and the generation phase re-embedding the same questions as the
// retrieval phase) cost nothing.
type CachingEmbedder struct {
	Inner rag.EmbeddingClient
	Cache *cache.Cache
	// Model is part of the cache key. The product hard-codes
	// text-embedding-3-small in rag.OpenAIClient.
	Model string
}

var _ rag.EmbeddingClient = (*CachingEmbedder)(nil)

func (c *CachingEmbedder) key(text string) string { return cache.Key(c.Model, text) }

// CreateEmbedding implements rag.EmbeddingClient.
func (c *CachingEmbedder) CreateEmbedding(ctx context.Context, text string) ([]float32, error) {
	var v []float32
	if ok, err := c.Cache.Get(embeddingNamespace, c.key(text), &v); err == nil && ok && len(v) > 0 {
		return v, nil
	}
	v, err := c.Inner.CreateEmbedding(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("create embedding: %w", err)
	}
	_ = c.Cache.Put(embeddingNamespace, c.key(text), v) // best-effort
	return v, nil
}

// CreateEmbeddingsBatch implements rag.EmbeddingClient, only sending cache misses upstream.
func (c *CachingEmbedder) CreateEmbeddingsBatch(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	var missIdx []int
	var missTexts []string
	for i, t := range texts {
		var v []float32
		if ok, err := c.Cache.Get(embeddingNamespace, c.key(t), &v); err == nil && ok && len(v) > 0 {
			out[i] = v
			continue
		}
		missIdx = append(missIdx, i)
		missTexts = append(missTexts, t)
	}
	if len(missTexts) == 0 {
		return out, nil
	}
	vecs, err := c.Inner.CreateEmbeddingsBatch(ctx, missTexts)
	if err != nil {
		return nil, fmt.Errorf("create embeddings batch: %w", err)
	}
	if len(vecs) != len(missTexts) {
		return nil, fmt.Errorf("embedding batch returned %d vectors for %d texts", len(vecs), len(missTexts))
	}
	for j, i := range missIdx {
		out[i] = vecs[j]
		_ = c.Cache.Put(embeddingNamespace, c.key(texts[i]), vecs[j])
	}
	return out, nil
}
