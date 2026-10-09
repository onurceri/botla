package harness

import (
	"context"
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/onurceri/botla-app/internal/evals/cache"
	"github.com/onurceri/botla-app/internal/evals/dataset"
	"github.com/onurceri/botla-app/internal/evals/usage"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/pkg/config"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// lexicalEmbedding is a deterministic bag-of-words vector, so that the fake
// provider's similarity search behaves like a (crude) real one.
func lexicalEmbedding(text string, dim int) []float64 {
	v := make([]float64, dim)
	for _, w := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if len(w) < 3 {
			continue
		}
		h := fnv.New32a()
		_, _ = h.Write([]byte(w))
		v[int(h.Sum32())%dim]++
	}
	var n float64
	for _, x := range v {
		n += x * x
	}
	if n > 0 {
		for i := range v {
			v[i] /= math.Sqrt(n)
		}
	}
	return v
}

// fakeOpenAI serves /v1/embeddings and /v1/chat/completions in the OpenAI wire
// format, so the product's real rag.OpenAIClient can be exercised offline.
type fakeOpenAI struct {
	mu           sync.Mutex
	chatRequests []map[string]any
	answer       string
	failChat     bool
	failEmbed    bool
}

func (f *fakeOpenAI) handler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		require.NoError(t, json.Unmarshal(raw, &body))
		switch r.URL.Path {
		case "/v1/embeddings":
			if f.failEmbed {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			var inputs []string
			switch in := body["input"].(type) {
			case string:
				inputs = []string{in}
			case []any:
				for _, s := range in {
					inputs = append(inputs, s.(string))
				}
			}
			data := make([]map[string]any, len(inputs))
			for i, s := range inputs {
				data[i] = map[string]any{"embedding": lexicalEmbedding(s, 64)}
			}
			writeJSON(w, map[string]any{"model": "text-embedding-3-small", "data": data, "usage": map[string]any{"prompt_tokens": 10 * len(inputs), "total_tokens": 10 * len(inputs)}})
		case "/v1/chat/completions":
			f.mu.Lock()
			f.chatRequests = append(f.chatRequests, body)
			f.mu.Unlock()
			if f.failChat {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			content := f.answer
			if strings.Contains(string(raw), "capability_summary") {
				content = `{"capability_summary":"Provides information about ` + firstWords(string(raw)) + `","suggested_questions":["q1?"]}`
			}
			writeJSON(w, map[string]any{
				"model":   "gpt-4o-mini-2024-07-18",
				"choices": []any{map[string]any{"message": map[string]any{"role": "assistant", "content": content}, "finish_reason": "stop"}},
				"usage":   map[string]any{"prompt_tokens": 100, "completion_tokens": 20, "total_tokens": 120},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
}

func firstWords(s string) string {
	if i := strings.Index(s, "# "); i >= 0 {
		s = s[i+2:]
	}
	if len(s) > 12 {
		s = s[:12]
	}
	return strings.Map(func(r rune) rune {
		if r == '"' || r == '\\' || r == '\n' {
			return ' '
		}
		return r
	}, s)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func corpusDir(t *testing.T) *dataset.Corpus {
	t.Helper()
	dir := t.TempDir()
	docs := map[string]string{
		"warranty": "# Warranty\n\nEvery sensor has a 24-month warranty from the delivery date.",
		"pricing":  "# Pricing\n\nThe Starter plan costs 9 euros per sensor per month.",
		"leds":     "# LEDs\n\nBlinking amber means the sensor is not connected to Wi-Fi.",
	}
	for id, txt := range docs {
		require.NoError(t, os.WriteFile(filepath.Join(dir, id+".md"), []byte(txt), 0o600))
	}
	c, err := dataset.LoadCorpus(dir)
	require.NoError(t, err)
	return c
}

type env struct {
	fake      *fakeOpenAI
	pipe      *Pipeline
	vector    *MemoryVectorStore
	transport *usage.Transport
}

func newEnv(t *testing.T) *env {
	t.Helper()
	fake := &fakeOpenAI{answer: "Every sensor has a 24-month warranty."}
	srv := httptest.NewServer(fake.handler(t))
	t.Cleanup(srv.Close)

	tr := &usage.Transport{Pricing: &usage.Pricing{Models: map[string]usage.Price{"gpt-4o-mini": {Input: 0.15, Output: 0.6}}}}
	cfg := &config.Config{OPENAI_API_KEY: "test", OPENAI_API_BASE: srv.URL}
	oai, err := rag.NewOpenAIClient(cfg, rag.WithHTTPClient(&http.Client{Transport: tr}))
	require.NoError(t, err)
	factory := rag.NewClientFactory(cfg)
	factory.RegisterClient("openai", oai)

	store := cache.New(t.TempDir())
	emb := &CachingEmbedder{Inner: oai, Cache: store, Model: "text-embedding-3-small"}
	vec := NewMemoryVectorStore()
	log := logger.New("ERROR")
	ing := &Ingestor{Embedder: emb, Vector: vec, LLM: oai, Cache: store, Log: log}
	ix, err := ing.Ingest(context.Background(), "bot-1", corpusDir(t))
	require.NoError(t, err)

	pipe := &Pipeline{
		Factory: factory, Embedder: emb, Vector: vec, Index: ix, Log: log,
		Bot: models.Chatbot{
			ID: "bot-1", UserID: "u", Name: "Test Bot", LanguageCode: "en", Model: "gpt-4o-mini", MaxTokens: 256,
			// The lexical fake embedding gives low absolute similarities; relax
			// the thresholds so the high-tier path is exercised.
			ThresholdConfig: &models.ThresholdConfig{HighThreshold: 0.2, MediumThreshold: 0.1, FallbackMode: "smart"},
		},
		Plan: models.Plan{Limits: &models.PlanLimits{ChatRAGTopK: 2, ChatRAGMaxContextTokens: 2000, GuardrailsCanUseSmartFallback: true}},
	}
	return &env{fake: fake, pipe: pipe, vector: vec, transport: tr}
}

func TestIngest_UsesProductChunkingAndMetadata(t *testing.T) {
	e := newEnv(t)
	ix := e.pipe.Index
	assert.Equal(t, 3, ix.TotalChunks(), "short docs produce one chunk each")
	assert.Equal(t, 3, e.vector.Len())
	txt, ok := ix.ChunkText("warranty", 0)
	require.True(t, ok)
	assert.Equal(t, "# Warranty Every sensor has a 24-month warranty from the delivery date.", txt, "text.NormalizeTR collapses whitespace like the product")
	assert.Len(t, ix.Capabilities, 3)
	assert.True(t, strings.HasPrefix(ix.Capabilities[0], "Provides information"))

	// Re-ingesting replaces points instead of duplicating them.
	ing := &Ingestor{Embedder: e.pipe.Embedder, Vector: e.vector}
	_, err := ing.Ingest(context.Background(), "bot-1", corpusDir(t))
	require.NoError(t, err)
	assert.Equal(t, 3, e.vector.Len())
}

func TestRetrieve(t *testing.T) {
	e := newEnv(t)
	r, err := e.pipe.Retrieve(context.Background(), "How long is the sensor warranty from delivery?", 10)
	require.NoError(t, err)
	require.Len(t, r.Ranked, 3)
	assert.Equal(t, "warranty", r.RankedDocs()[0])
	assert.Equal(t, "high", r.Tier)
	assert.Greater(t, r.HighestScore, 0.2)
	assert.Contains(t, r.ContextDocs(), "warranty")
	assert.LessOrEqual(t, len(r.Context), 2, "plan top-k caps the context")
}

func TestGenerate_RunsProductChatPipeline(t *testing.T) {
	e := newEnv(t)
	m := usage.NewMeter()
	ctx := usage.WithMeter(context.Background(), m)
	g, err := e.pipe.Generate(ctx, "How long is the sensor warranty from delivery?", "s1")
	require.NoError(t, err)
	assert.Equal(t, "Every sensor has a 24-month warranty.", g.Answer)
	assert.Equal(t, "high", g.Tier)
	assert.Equal(t, 120, g.TokensUsed)
	assert.Contains(t, g.ContextDocs(), "warranty")
	assert.Contains(t, g.Context, "24-month warranty")
	assert.Positive(t, g.Latency)

	// The request the product sent is the product's prompt, not ours.
	e.fake.mu.Lock()
	last := e.fake.chatRequests[len(e.fake.chatRequests)-1]
	e.fake.mu.Unlock()
	msgs := last["messages"].([]any)
	system := msgs[0].(map[string]any)["content"].(string)
	user := msgs[len(msgs)-1].(map[string]any)["content"].(string)
	assert.Contains(t, system, `You are an AI assistant named "Test Bot"`)
	assert.Contains(t, system, "Provides information", "capability summaries reach the system prompt")
	assert.Contains(t, user, "The following documents are relevant to the query")
	assert.NotEmpty(t, last["tools"], "built-in tools are offered, as in production")

	// Usage is metered from provider responses: one question embedding + one chat call.
	by := m.ByModel()
	chat := by["gpt-4o-mini-2024-07-18"]
	assert.Equal(t, 1, chat.Calls)
	assert.Equal(t, 100, chat.PromptTokens)
	assert.Equal(t, 20, chat.CompletionTokens)
	assert.InDelta(t, (100*0.15+20*0.6)/1e6, chat.CostUSD, 1e-12)
	assert.Equal(t, 1, by["text-embedding-3-small"].Calls)
}

func TestGenerate_LowTierUsesSmartFallback(t *testing.T) {
	e := newEnv(t)
	e.fake.answer = "My knowledge sources don't cover that."
	g, err := e.pipe.Generate(context.Background(), "zebra quantum lasagna", "s2")
	require.NoError(t, err)
	assert.Equal(t, "low", g.Tier)
	assert.Empty(t, g.ContextChunks)
	assert.Empty(t, g.Context)
	assert.Equal(t, "My knowledge sources don't cover that.", g.Answer)
}

func TestGenerate_SwallowedLLMErrorSurfacesAsFallbackText(t *testing.T) {
	e := newEnv(t)
	e.fake.failChat = true
	g, err := e.pipe.Generate(context.Background(), "How long is the sensor warranty from delivery?", "s3")
	require.NoError(t, err, "the product swallows LLM errors")
	assert.Equal(t, "An error occurred, please try again later.", g.Answer)
}

func TestRetrieveAndIngestErrors(t *testing.T) {
	e := newEnv(t)
	e.fake.failEmbed = true
	_, err := e.pipe.Retrieve(context.Background(), "a question never embedded before", 5)
	assert.ErrorContains(t, err, "embed question")

	ing := &Ingestor{Embedder: e.pipe.Embedder, Vector: e.vector}
	_, err = ing.Ingest(context.Background(), "bot-2", corpusDir(t))
	require.NoError(t, err, "embeddings for the corpus are served from cache")

	failing := &Ingestor{Embedder: &CachingEmbedder{Inner: failingEmbedder{}, Cache: cache.New("")}, Vector: NewMemoryVectorStore()}
	_, err = failing.Ingest(context.Background(), "bot-3", corpusDir(t))
	assert.ErrorContains(t, err, "embed")

	_, err = (&Ingestor{Vector: brokenVector{ensureErr: errors.New("down")}}).Ingest(context.Background(), "b", corpusDir(t))
	assert.ErrorContains(t, err, "ensure vector collection")
	_, err = (&Ingestor{Vector: brokenVector{deleteErr: errors.New("down")}}).Ingest(context.Background(), "b", corpusDir(t))
	assert.ErrorContains(t, err, "clear previous points")

	p := *e.pipe
	p.Embedder = fixedEmbedder{}
	p.Vector = brokenVector{searchErr: errors.New("qdrant down")}
	_, err = p.Retrieve(context.Background(), "q", 5)
	assert.ErrorContains(t, err, "vector search")
}

func TestCapabilityExtractionFailureIsNonFatal(t *testing.T) {
	e := newEnv(t)
	e.fake.failChat = true
	ing := &Ingestor{Embedder: e.pipe.Embedder, Vector: NewMemoryVectorStore(), LLM: failingLLM{}, Cache: cache.New(""), Log: logger.New("ERROR")}
	ix, err := ing.Ingest(context.Background(), "bot-x", corpusDir(t))
	require.NoError(t, err)
	assert.Empty(t, ix.Capabilities)
}

func TestGenerate_CancelledContextDegradesLikeProduct(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// ProcessChat does not return an error here: repositories are in-memory and
	// the product converts the failed LLM call into user-facing fallback text.
	// The runner detects such answers via Config.ErrorMessages.
	g, err := e.pipe.Generate(ctx, "warranty?", "s4")
	require.NoError(t, err)
	assert.NotEmpty(t, g.Answer)
}

// --- MemoryVectorStore ---

func TestMemoryVectorStore(t *testing.T) {
	ctx := context.Background()
	m := NewMemoryVectorStore()
	require.NoError(t, m.EnsureEmbeddingsCollection(ctx))
	require.NoError(t, m.UpsertEmbedding(ctx, "p1", []float32{1, 0}, rag.EmbeddingPayload{ChatbotID: "b", SourceID: "s1", ChunkIndex: 1}))
	require.NoError(t, m.UpsertEmbedding(ctx, "p0", []float32{0.9, 0.1}, rag.EmbeddingPayload{ChatbotID: "b", SourceID: "s1", ChunkIndex: 0}))
	require.NoError(t, m.UpsertEmbedding(ctx, "p2", []float32{0, 1}, rag.EmbeddingPayload{ChatbotID: "b", SourceID: "s2"}))
	require.NoError(t, m.UpsertEmbedding(ctx, "other", []float32{1, 0}, rag.EmbeddingPayload{ChatbotID: "x", SourceID: "s9"}))
	require.NoError(t, m.UpsertEmbedding(ctx, "zero", []float32{0, 0}, rag.EmbeddingPayload{ChatbotID: "b", SourceID: "s3"}))
	assert.Error(t, m.UpsertEmbedding(ctx, 42, []float32{1}, rag.EmbeddingPayload{}))

	res, err := m.SearchSimilar(ctx, []float32{1, 0}, "b", 2)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "p1", res[0].ID)
	assert.InDelta(t, 1, res[0].Score, 1e-9)
	assert.Equal(t, "p0", res[1].ID)

	all, err := m.SearchSimilar(ctx, []float32{1, 0}, "b", 0)
	require.NoError(t, err)
	assert.Len(t, all, 4, "other chatbots are filtered out")
	assert.InDelta(t, 0, all[3].Score, 1e-9, "zero vector scores 0")

	// Equal scores tie-break by ID for determinism.
	require.NoError(t, m.UpsertEmbedding(ctx, "a-dup", []float32{1, 0}, rag.EmbeddingPayload{ChatbotID: "b", SourceID: "s4"}))
	res, err = m.SearchSimilar(ctx, []float32{1, 0}, "b", 2)
	require.NoError(t, err)
	assert.Equal(t, "a-dup", res[0].ID)

	chunks, next, err := m.ScrollChunks(ctx, "s1", 0, nil)
	require.NoError(t, err)
	assert.Nil(t, next)
	require.Len(t, chunks, 2)
	assert.Equal(t, 0, chunks[0].Payload.ChunkIndex)
	chunks, _, err = m.ScrollChunks(ctx, "s1", 1, nil)
	require.NoError(t, err)
	assert.Len(t, chunks, 1)

	require.NoError(t, m.DeleteBySourceID(ctx, "s1"))
	assert.Equal(t, 4, m.Len())
	assert.InDelta(t, 0, cosine([]float32{1}, 1, []float32{1, 2}, 1), 1e-9, "dimension mismatch")
}

// --- CachingEmbedder ---

type countingEmbedder struct {
	mu      sync.Mutex
	batches [][]string
	singles int
	shortBy int
	err     error
}

func (c *countingEmbedder) CreateEmbedding(_ context.Context, text string) ([]float32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.singles++
	if c.err != nil {
		return nil, c.err
	}
	return []float32{float32(len(text))}, nil
}

func (c *countingEmbedder) CreateEmbeddingsBatch(_ context.Context, texts []string) ([][]float32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.batches = append(c.batches, texts)
	if c.err != nil {
		return nil, c.err
	}
	out := make([][]float32, len(texts)-c.shortBy)
	for i := range out {
		out[i] = []float32{float32(len(texts[i]))}
	}
	return out, nil
}

func TestCachingEmbedder(t *testing.T) {
	inner := &countingEmbedder{}
	c := &CachingEmbedder{Inner: inner, Cache: cache.New(t.TempDir()), Model: "m"}
	ctx := context.Background()

	v, err := c.CreateEmbedding(ctx, "abc")
	require.NoError(t, err)
	assert.Equal(t, []float32{3}, v)
	_, err = c.CreateEmbedding(ctx, "abc")
	require.NoError(t, err)
	assert.Equal(t, 1, inner.singles)

	out, err := c.CreateEmbeddingsBatch(ctx, []string{"abc", "de", "f"})
	require.NoError(t, err)
	assert.Equal(t, [][]float32{{3}, {2}, {1}}, out)
	assert.Equal(t, [][]string{{"de", "f"}}, inner.batches, "only misses go upstream")

	_, err = c.CreateEmbeddingsBatch(ctx, []string{"de", "f"})
	require.NoError(t, err)
	assert.Len(t, inner.batches, 1, "fully cached batch makes no call")

	inner.shortBy = 1
	_, err = c.CreateEmbeddingsBatch(ctx, []string{"x", "y"})
	assert.ErrorContains(t, err, "returned 1 vectors for 2 texts")

	inner.err = errors.New("401")
	_, err = c.CreateEmbeddingsBatch(ctx, []string{"zz"})
	assert.ErrorContains(t, err, "401")
	_, err = c.CreateEmbedding(ctx, "new")
	assert.ErrorContains(t, err, "401")
}

// --- small fakes ---

type failingEmbedder struct{}

func (failingEmbedder) CreateEmbedding(context.Context, string) ([]float32, error) {
	return nil, errors.New("no")
}

func (failingEmbedder) CreateEmbeddingsBatch(context.Context, []string) ([][]float32, error) {
	return nil, errors.New("no")
}

type fixedEmbedder struct{}

func (fixedEmbedder) CreateEmbedding(context.Context, string) ([]float32, error) {
	return []float32{1}, nil
}

func (fixedEmbedder) CreateEmbeddingsBatch(_ context.Context, t []string) ([][]float32, error) {
	out := make([][]float32, len(t))
	for i := range out {
		out[i] = []float32{1}
	}
	return out, nil
}

type failingLLM struct{}

func (failingLLM) CreateCompletion(context.Context, models.CompletionParams) (*models.CompletionResult, error) {
	return nil, errors.New("llm down")
}

func (failingLLM) GetModelInfo() models.ModelInfo { return models.ModelInfo{} }

type brokenVector struct {
	ensureErr, deleteErr, searchErr error
}

func (b brokenVector) EnsureEmbeddingsCollection(context.Context) error { return b.ensureErr }
func (b brokenVector) UpsertEmbedding(context.Context, interface{}, []float32, rag.EmbeddingPayload) error {
	return nil
}
func (b brokenVector) SearchSimilar(context.Context, []float32, string, int) ([]rag.SearchResult, error) {
	return nil, b.searchErr
}
func (b brokenVector) DeleteBySourceID(context.Context, string) error { return b.deleteErr }
func (b brokenVector) ScrollChunks(context.Context, string, int, interface{}) ([]rag.SearchResult, *string, error) {
	return nil, nil, nil
}
