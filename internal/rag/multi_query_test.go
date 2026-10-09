package rag

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurceri/botla-app/internal/models"
)

// fakeVector returns a fixed ranking per query, keyed by the first component
// of the query embedding, and records the requested topK.
type fakeVector struct {
	byQuery   map[float32][]SearchResult
	err       error
	requested []int
}

func (f *fakeVector) EnsureEmbeddingsCollection(context.Context) error { return nil }
func (f *fakeVector) UpsertEmbedding(context.Context, interface{}, []float32, EmbeddingPayload) error {
	return nil
}
func (f *fakeVector) DeleteBySourceID(context.Context, string) error { return nil }
func (f *fakeVector) ScrollChunks(context.Context, string, int, interface{}) ([]SearchResult, *string, error) {
	return nil, nil, nil
}
func (f *fakeVector) SearchSimilar(_ context.Context, emb []float32, _ string, topK int) ([]SearchResult, error) {
	f.requested = append(f.requested, topK)
	if f.err != nil {
		return nil, f.err
	}
	res := append([]SearchResult(nil), f.byQuery[emb[0]]...)
	if len(res) > topK {
		res = res[:topK]
	}
	return res, nil
}

func hit(src string, idx int, score float64, text string) SearchResult {
	return SearchResult{Score: score, Payload: EmbeddingPayload{SourceID: src, ChunkIndex: idx, OriginalText: text}}
}

func docs(cs []models.ChunkMetadata) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.SourceID
	}
	return out
}

var testThresholds = &models.ThresholdConfig{HighThreshold: 0.5, MediumThreshold: 0.3}

func TestRetrievalStrategyFromEnv(t *testing.T) {
	t.Setenv(EnvQueryDecomposition, "")
	assert.Equal(t, RetrievalStrategy{Decompose: DefaultQueryDecomposition}, RetrievalStrategyFromEnv())
	t.Setenv(EnvQueryDecomposition, "true")
	assert.True(t, RetrievalStrategyFromEnv().Decompose)
	t.Setenv(EnvQueryDecomposition, "0")
	assert.False(t, RetrievalStrategyFromEnv().Decompose)
	t.Setenv(EnvQueryDecomposition, "maybe")
	assert.Equal(t, DefaultQueryDecomposition, RetrievalStrategyFromEnv().Decompose, "invalid values fall back to the default")
}

func TestFuseRoundRobin(t *testing.T) {
	assert.Empty(t, FuseRoundRobin(nil))

	single := []SearchResult{hit("", 0, 0.9, "a"), hit("", 0, 0.8, "b")}
	assert.Equal(t, single, FuseRoundRobin([][]SearchResult{single}), "a single ranking is passed through, duplicates included")

	q1 := []SearchResult{hit("a", 0, 0.9, ""), hit("b", 0, 0.8, ""), hit("c", 0, 0.7, "")}
	q2 := []SearchResult{hit("d", 0, 0.6, ""), hit("b", 0, 0.85, "")}
	q3 := []SearchResult{hit("a", 0, 0.95, "")}
	got := FuseRoundRobin([][]SearchResult{q1, q2, q3})
	var ids []string
	for _, h := range got {
		ids = append(ids, h.Payload.SourceID)
	}
	assert.Equal(t, []string{"a", "d", "b", "c"}, ids, "rank 1 of every query first, duplicates skipped")
	assert.InDelta(t, 0.95, got[0].Score, 1e-9, "a duplicate keeps the best score")
	assert.InDelta(t, 0.85, got[2].Score, 1e-9)

	sameDocOtherChunk := FuseRoundRobin([][]SearchResult{{hit("a", 0, 0.9, "")}, {hit("a", 1, 0.8, "")}})
	assert.Len(t, sameDocOtherChunk, 2, "chunks of one document are distinct")
}

func TestSearchContextMulti_SingleQueryMatchesTiered(t *testing.T) {
	vc := &fakeVector{byQuery: map[float32][]SearchResult{
		1: {hit("b", 0, 0.4, "B"), hit("a", 0, 0.6, "A"), hit("c", 0, 0.2, "C")},
	}}
	multi, err := SearchContextMulti(context.Background(), vc, [][]float32{{1}}, "bot", SearchOptions{TopK: 5, MaxContextTokens: 1000, Thresholds: testThresholds})
	require.NoError(t, err)
	tiered, err := SearchContextTiered(context.Background(), vc, []float32{1}, "bot", 5, 1000, testThresholds)
	require.NoError(t, err)
	assert.Equal(t, tiered, multi)
	assert.Equal(t, TierHigh, multi.Tier)
	assert.Equal(t, []string{"a", "b"}, docs(multi.Chunks), "sorted by score, below-threshold chunk dropped")
	assert.Equal(t, "A\n---\nB", multi.ContextText)
	assert.Len(t, multi.AllChunks, 3)
	assert.InDelta(t, 0.5, multi.AverageScore, 1e-9)
}

func TestSearchContextMulti_FusesSubQueriesUnderTopK(t *testing.T) {
	vc := &fakeVector{byQuery: map[float32][]SearchResult{
		1: {hit("a", 0, 0.45, "A"), hit("b", 0, 0.44, "B"), hit("c", 0, 0.43, "C")},
		2: {hit("x", 0, 0.62, "X"), hit("y", 0, 0.35, "Y"), hit("a", 0, 0.31, "A")},
	}}
	res, err := SearchContextMulti(context.Background(), vc, [][]float32{{1}, {2}, nil}, "bot", SearchOptions{TopK: 3, MaxContextTokens: 1000, Thresholds: testThresholds})
	require.NoError(t, err)
	assert.Equal(t, TierHigh, res.Tier, "tier comes from the best score across queries")
	assert.InDelta(t, 0.62, res.HighestScore, 1e-9)
	assert.Equal(t, []string{"a", "x", "b"}, docs(res.Chunks), "round-robin, capped at top-K")
	assert.Equal(t, []int{3, 3}, vc.requested, "empty embeddings are skipped")
}

func TestSearchContextMulti_EdgeCases(t *testing.T) {
	ctx := context.Background()
	var nilVC *fakeVector
	_, err := SearchContextMulti(ctx, nilVC, [][]float32{{1}}, "bot", SearchOptions{})
	assert.Error(t, err)
	_, err = SearchContextMulti(ctx, nil, [][]float32{{1}}, "bot", SearchOptions{})
	assert.Error(t, err)

	vc := &fakeVector{byQuery: map[float32][]SearchResult{}}
	res, err := SearchContextMulti(ctx, vc, nil, "bot", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, TierLow, res.Tier)
	res, err = SearchContextMulti(ctx, vc, [][]float32{{1}}, "", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, TierLow, res.Tier)
	res, err = SearchContextMulti(ctx, vc, [][]float32{{1}}, "bot", SearchOptions{})
	require.NoError(t, err)
	assert.Equal(t, TierLow, res.Tier, "no hits")

	_, err = SearchContextMulti(ctx, &fakeVector{err: errors.New("down")}, [][]float32{{1}}, "bot", SearchOptions{})
	assert.ErrorContains(t, err, "search similar failed")

	low := &fakeVector{byQuery: map[float32][]SearchResult{1: {hit("a", 0, 0.1, "A")}}}
	res, err = SearchContextMulti(ctx, low, [][]float32{{1}}, "bot", SearchOptions{Thresholds: testThresholds})
	require.NoError(t, err)
	assert.Equal(t, TierLow, res.Tier)
	assert.Len(t, res.AllChunks, 1)

	medium := &fakeVector{byQuery: map[float32][]SearchResult{1: {hit("a", 0, 0.4, "A")}}}
	res, err = SearchContextMulti(ctx, medium, [][]float32{{1}}, "bot", SearchOptions{})
	require.NoError(t, err, "nil thresholds use defaults")
	assert.NotEqual(t, TierHigh, res.Tier)

	empty := &fakeVector{byQuery: map[float32][]SearchResult{1: {hit("a", 0, 0.9, "   ")}}}
	res, err = SearchContextMulti(ctx, empty, [][]float32{{1}}, "bot", SearchOptions{Thresholds: testThresholds})
	require.NoError(t, err)
	assert.Equal(t, TierLow, res.Tier, "chunks without text never reach the context")

	big := &fakeVector{byQuery: map[float32][]SearchResult{1: {hit("a", 0, 0.9, "short"), hit("b", 0, 0.8, "a much longer chunk of text that will not fit")}}}
	res, err = SearchContextMulti(ctx, big, [][]float32{{1}}, "bot", SearchOptions{MaxContextTokens: CountTokens("short", "tr"), Thresholds: testThresholds})
	require.NoError(t, err)
	assert.Equal(t, []string{"a"}, docs(res.Chunks), "stops at the token budget")
}

func TestResolveLimitsFromEnv(t *testing.T) {
	t.Setenv("RAG_TOPK", "")
	t.Setenv("RAG_MAX_CONTEXT_TOKENS", "")
	assert.Equal(t, 5, resolveTopK(0))
	assert.Equal(t, 2000, resolveMaxContextTokens(0))
	assert.Equal(t, 7, resolveTopK(7))
	assert.Equal(t, 900, resolveMaxContextTokens(900))
	t.Setenv("RAG_TOPK", "8")
	t.Setenv("RAG_MAX_CONTEXT_TOKENS", "3000")
	assert.Equal(t, 8, resolveTopK(0))
	assert.Equal(t, 3000, resolveMaxContextTokens(0))
}
