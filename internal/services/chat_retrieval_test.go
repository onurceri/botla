package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
)

// retrievalEmbedder maps each known text to a one-dimensional "embedding"
// whose value identifies the query in retrievalVector.
type retrievalEmbedder struct {
	ids       map[string]float32
	failOne   bool
	failBatch bool
}

func (e *retrievalEmbedder) CreateEmbedding(_ context.Context, text string) ([]float32, error) {
	if e.failOne {
		return nil, errors.New("embed down")
	}
	return []float32{e.ids[text]}, nil
}

func (e *retrievalEmbedder) CreateEmbeddingsBatch(_ context.Context, texts []string) ([][]float32, error) {
	if e.failBatch {
		return nil, errors.New("batch down")
	}
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = []float32{e.ids[t]}
	}
	return out, nil
}

type retrievalVector struct {
	byQuery  map[float32][]rag.SearchResult
	err      error
	searched []float32
}

func (v *retrievalVector) EnsureEmbeddingsCollection(context.Context) error { return nil }
func (v *retrievalVector) UpsertEmbedding(context.Context, interface{}, []float32, rag.EmbeddingPayload) error {
	return nil
}
func (v *retrievalVector) DeleteBySourceID(context.Context, string) error { return nil }
func (v *retrievalVector) ScrollChunks(context.Context, string, int, interface{}) ([]rag.SearchResult, *string, error) {
	return nil, nil, nil
}
func (v *retrievalVector) SearchSimilar(_ context.Context, emb []float32, _ string, topK int) ([]rag.SearchResult, error) {
	v.searched = append(v.searched, emb[0])
	if v.err != nil {
		return nil, v.err
	}
	res := v.byQuery[emb[0]]
	if len(res) > topK {
		res = res[:topK]
	}
	return res, nil
}

// decomposerLLM is a tools-capable LLM client that returns a fixed reply.
type decomposerLLM struct {
	reply string
	err   error
	calls int
}

func (d *decomposerLLM) CreateCompletion(context.Context, models.CompletionParams) (*models.CompletionResult, error) {
	return nil, errors.New("not used")
}
func (d *decomposerLLM) GetModelInfo() models.ModelInfo { return models.ModelInfo{} }
func (d *decomposerLLM) CreateCompletionWithTools(context.Context, []rag.ChatMessage, []rag.Tool, string, float32, int) (*rag.ChatResponseWithTools, error) {
	d.calls++
	if d.err != nil {
		return nil, d.err
	}
	r := &rag.ChatResponseWithTools{}
	r.Usage.TotalTokens = 30
	content := d.reply
	r.Choices = append(r.Choices, struct {
		Message      rag.ChatMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	}{Message: rag.ChatMessage{Role: "assistant", Content: &content}})
	return r, nil
}

const (
	twoPartQ = "Who founded Halvane, and when did the Beacon S2 launch?"
	subQ1    = "Who founded Halvane?"
	subQ2    = "When did the Beacon S2 launch?"
)

func sr(src string, score float64) rag.SearchResult {
	return rag.SearchResult{Score: score, Payload: rag.EmbeddingPayload{SourceID: src, OriginalText: "text of " + src}}
}

type retrievalFixture struct {
	svc *ChatService
	emb *retrievalEmbedder
	vec *retrievalVector
	llm *decomposerLLM
	bot *models.Chatbot
}

func newRetrievalFixture(reply string) *retrievalFixture {
	emb := &retrievalEmbedder{ids: map[string]float32{twoPartQ: 1, subQ1: 2, subQ2: 3}}
	vec := &retrievalVector{byQuery: map[float32][]rag.SearchResult{
		// The full question is dominated by the product part.
		1: {sr("specs", 0.60), sr("firmware", 0.55), sr("pricing", 0.50), sr("company", 0.31)},
		2: {sr("company", 0.58), sr("specs", 0.40)},
		3: {sr("specs", 0.66), sr("firmware", 0.45)},
	}}
	llm := &decomposerLLM{reply: reply}
	factory := rag.NewClientFactory(&config.Config{OPENAI_API_KEY: "test"})
	factory.RegisterClient("openai", llm)
	svc := &ChatService{Factory: factory, Embedder: emb, QC: vec, Log: logger.New("ERROR")}
	return &retrievalFixture{svc: svc, emb: emb, vec: vec, llm: llm, bot: &models.Chatbot{ID: "bot", Model: "gpt-4o-mini"}}
}

func (f *retrievalFixture) retrieve(question string) RetrievalOutcome {
	return f.svc.RetrieveContext(context.Background(), f.bot, question, models.RAGConfig{TopK: 2, MaxContextTokens: 1000}, models.DefaultThresholdConfig())
}

func contextDocs(r *rag.TieredSearchResult) []string {
	out := make([]string, 0, len(r.Chunks))
	for _, c := range r.Chunks {
		out = append(out, c.SourceID)
	}
	return out
}

func TestRetrieveContext_DecompositionOffIsSingleQuery(t *testing.T) {
	f := newRetrievalFixture(`{"queries": ["` + subQ1 + `", "` + subQ2 + `"]}`)
	out := f.retrieve(twoPartQ)
	assert.Zero(t, f.llm.calls, "no LLM call when the strategy is off")
	assert.Nil(t, out.SubQueries)
	assert.Zero(t, out.LLMTokens)
	assert.Equal(t, []float32{1}, f.vec.searched)
	assert.Equal(t, []string{"specs", "firmware"}, contextDocs(out.Result))

	direct, err := rag.SearchContextTiered(context.Background(), f.vec, []float32{1}, "bot", 2, 1000, models.DefaultThresholdConfig())
	require.NoError(t, err)
	assert.Equal(t, direct, out.Result, "identical to the original single-query search")
}

func TestRetrieveContext_DecomposesMultiPartQuestion(t *testing.T) {
	f := newRetrievalFixture(`{"queries": ["` + subQ1 + `", "` + subQ2 + `"]}`)
	f.svc.Retrieval.Decompose = true
	out := f.retrieve(twoPartQ)
	assert.Equal(t, 1, f.llm.calls)
	assert.Equal(t, []string{subQ1, subQ2}, out.SubQueries)
	assert.Equal(t, 30, out.LLMTokens)
	assert.Equal(t, []float32{1, 2, 3}, f.vec.searched, "the original question is searched first, then each sub-query")
	assert.Equal(t, []string{"specs", "company"}, contextDocs(out.Result), "the second part's document now reaches the context")
	assert.Equal(t, rag.TierHigh, out.Result.Tier)
	assert.InDelta(t, 0.66, out.Result.HighestScore, 1e-9, "tier uses the best score across sub-queries")
}

func TestRetrieveContext_SinglePartUnchanged(t *testing.T) {
	f := newRetrievalFixture(`{"queries": ["Who founded the company?"]}`)
	f.svc.Retrieval.Decompose = true
	out := f.retrieve(twoPartQ)
	assert.Nil(t, out.SubQueries)
	assert.Equal(t, 30, out.LLMTokens, "the call is still metered")
	assert.Equal(t, []float32{1}, f.vec.searched)
	assert.Equal(t, []string{"specs", "firmware"}, contextDocs(out.Result))
}

func TestRetrieveContext_Fallbacks(t *testing.T) {
	t.Run("decomposition error", func(t *testing.T) {
		f := newRetrievalFixture("")
		f.llm.err = errors.New("llm down")
		f.svc.Retrieval.Decompose = true
		out := f.retrieve(twoPartQ)
		assert.Nil(t, out.SubQueries)
		assert.Equal(t, []string{"specs", "firmware"}, contextDocs(out.Result))
	})
	t.Run("no LLM client", func(t *testing.T) {
		f := newRetrievalFixture("")
		f.svc.Factory = rag.NewClientFactory(nil)
		f.svc.Retrieval.Decompose = true
		out := f.retrieve(twoPartQ)
		assert.Nil(t, out.SubQueries)
		assert.Equal(t, []float32{1}, f.vec.searched)
	})
	t.Run("sub-query embedding error", func(t *testing.T) {
		f := newRetrievalFixture(`{"queries": ["` + subQ1 + `", "` + subQ2 + `"]}`)
		f.emb.failBatch = true
		f.svc.Retrieval.Decompose = true
		out := f.retrieve(twoPartQ)
		assert.Nil(t, out.SubQueries)
		assert.Equal(t, []float32{1}, f.vec.searched)
	})
	t.Run("question embedding error uses sub-queries", func(t *testing.T) {
		f := newRetrievalFixture(`{"queries": ["` + subQ1 + `", "` + subQ2 + `"]}`)
		f.emb.failOne = true
		f.svc.Retrieval.Decompose = true
		out := f.retrieve(twoPartQ)
		assert.Equal(t, []float32{2, 3}, f.vec.searched)
		assert.Equal(t, []string{"company", "specs"}, contextDocs(out.Result))
	})
	t.Run("question embedding error without decomposition", func(t *testing.T) {
		f := newRetrievalFixture("")
		f.emb.failOne = true
		out := f.retrieve(twoPartQ)
		assert.Equal(t, rag.TierLow, out.Result.Tier)
		assert.Empty(t, f.vec.searched)
	})
	t.Run("vector search error", func(t *testing.T) {
		f := newRetrievalFixture("")
		f.vec.err = errors.New("qdrant down")
		out := f.retrieve(twoPartQ)
		require.NotNil(t, out.Result)
		assert.Equal(t, rag.TierLow, out.Result.Tier)
	})
	t.Run("no embedder or vector store", func(t *testing.T) {
		f := newRetrievalFixture("")
		f.svc.QC = nil
		f.svc.Log = nil
		out := f.retrieve(twoPartQ)
		assert.Equal(t, rag.TierLow, out.Result.Tier)
	})
}

func TestPerformRAGSearch_CountsDecompositionTokens(t *testing.T) {
	f := newRetrievalFixture(`{"queries": ["` + subQ1 + `", "` + subQ2 + `"]}`)
	f.svc.Retrieval.Decompose = true
	cc := &chatContext{
		Bot:          f.bot,
		Request:      models.ChatRequest{Message: twoPartQ},
		RAGConfig:    models.RAGConfig{TopK: 2, MaxContextTokens: 1000},
		ThresholdCfg: models.DefaultThresholdConfig(),
		TotalTokens:  5,
	}
	f.svc.performRAGSearch(context.Background(), cc)
	assert.Equal(t, 35, cc.TotalTokens)
	assert.Len(t, cc.ChunkMetas, 2)
	assert.Len(t, cc.Sources, 2)
	assert.Equal(t, rag.TierHigh, cc.SearchResult.Tier)
}

func TestNewChatService_ReadsRetrievalStrategyFromEnv(t *testing.T) {
	t.Setenv(rag.EnvQueryDecomposition, "true")
	svc := NewChatService(nil, nil, nil, nil, nil, nil, rag.NewClientFactory(nil), &retrievalEmbedder{}, nil, nil, nil)
	assert.True(t, svc.Retrieval.Decompose)
}
