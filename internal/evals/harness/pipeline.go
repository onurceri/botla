package harness

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/logger"
)

// Pipeline runs retrieval and generation for one eval chatbot.
type Pipeline struct {
	Factory  *rag.ClientFactory
	Embedder rag.EmbeddingClient
	Vector   rag.VectorClient
	Index    *Index
	Bot      models.Chatbot
	Plan     models.Plan
	Log      *logger.Logger
	// Retrieval is the product retrieval strategy under test (query
	// decomposition). The zero value is plain single-query search.
	Retrieval rag.RetrievalStrategy
}

// RAGConfig is the per-plan retrieval configuration ProcessChatWithValidation
// would pass to ProcessChat.
func (p *Pipeline) RAGConfig() models.RAGConfig {
	return models.RAGConfig{TopK: p.Plan.Limits.ChatRAGTopK, MaxContextTokens: p.Plan.Limits.ChatRAGMaxContextTokens}
}

// RetrievedChunk is one ranked search hit.
type RetrievedChunk struct {
	SourceID   string  `json:"source_id"`
	ChunkIndex int     `json:"chunk_index"`
	Score      float64 `json:"score"`
}

// Retrieval is the outcome of the retrieval stage for one question.
type Retrieval struct {
	// Ranked is the raw similarity ranking (top-N chunks), used for rank metrics.
	Ranked []RetrievedChunk `json:"ranked"`
	// Context is what rag.SearchContextTiered admits into the prompt with the
	// plan's top-k, token budget and the bot's similarity thresholds.
	Context      []RetrievedChunk `json:"context"`
	Tier         string           `json:"tier"`
	HighestScore float64          `json:"highest_score"`
	// SubQueries are the standalone queries searched when query
	// decomposition split the question (empty for single-part questions).
	SubQueries []string `json:"sub_queries,omitempty"`
}

// RankedDocs returns the chunk ranking as source IDs.
func (r Retrieval) RankedDocs() []string { return sourceIDs(r.Ranked) }

// ContextDocs returns the source IDs that reached the LLM context.
func (r Retrieval) ContextDocs() []string { return sourceIDs(r.Context) }

func sourceIDs(cs []RetrievedChunk) []string {
	out := make([]string, len(cs))
	for i, c := range cs {
		out[i] = c.SourceID
	}
	return out
}

// Retrieve runs the product's retrieval (services.ChatService.RetrieveContext,
// the same call ProcessChat makes) and, for rank metrics, the raw similarity
// ranking at depth topN. When decomposition split the question, the ranking
// is the round-robin fusion of the question's and each sub-query's depth-topN
// ranking, the same fusion the product uses for context selection.
func (p *Pipeline) Retrieve(ctx context.Context, question string, topN int) (Retrieval, error) {
	svc := p.newService(nil)
	bot := p.Bot
	got := svc.RetrieveContext(ctx, &bot, question, p.RAGConfig(), p.Bot.ThresholdConfig)

	queries := append([]string{question}, got.SubQueries...)
	rankings := make([][]rag.SearchResult, 0, len(queries))
	for _, q := range queries {
		emb, err := p.Embedder.CreateEmbedding(ctx, q)
		if err != nil {
			return Retrieval{}, fmt.Errorf("embed question: %w", err)
		}
		hits, err := p.Vector.SearchSimilar(ctx, emb, p.Bot.ID, topN)
		if err != nil {
			return Retrieval{}, fmt.Errorf("vector search: %w", err)
		}
		sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
		rankings = append(rankings, hits)
	}
	hits := rag.FuseRoundRobin(rankings)
	if len(hits) > topN {
		hits = hits[:topN]
	}
	out := Retrieval{Ranked: make([]RetrievedChunk, len(hits)), SubQueries: got.SubQueries}
	for i, h := range hits {
		out.Ranked[i] = RetrievedChunk{SourceID: h.Payload.SourceID, ChunkIndex: h.Payload.ChunkIndex, Score: h.Score}
	}
	out.Tier = string(got.Result.Tier)
	out.HighestScore = got.Result.HighestScore
	for _, m := range got.Result.Chunks {
		out.Context = append(out.Context, RetrievedChunk{SourceID: m.SourceID, ChunkIndex: m.ChunkIndex, Score: m.Score})
	}
	return out, nil
}

// newService builds a ChatService over fresh in-memory repositories, wired to
// the eval clients and retrieval strategy. onSources, if set, receives the
// chunks ProcessChat saves as the answer's sources.
func (p *Pipeline) newService(onSources func([]models.ChunkMetadata)) *services.ChatService {
	plan := p.Plan
	planRepo := repository.NewMockPlanRepo()
	planRepo.GetByUserIDFunc = func(context.Context, string) (*models.Plan, error) { return &plan, nil }

	convRepo := &repository.MockConversationRepo{}
	convRepo.GetOrCreateBySessionIDFunc = func(_ context.Context, chatbotID, sid string) (*models.Conversation, error) {
		s := sid
		return &models.Conversation{ID: "eval-conv-" + sid, ChatbotID: chatbotID, SessionID: &s}, nil
	}
	convRepo.CreateMessageFunc = func(_ context.Context, msg *models.Message) (string, error) {
		return "eval-msg-" + msg.Role, nil
	}
	convRepo.SaveMessageSourcesFunc = func(_ context.Context, _ string, sources []models.ChunkMetadata) error {
		if onSources != nil {
			onSources(sources)
		}
		return nil
	}

	sourceRepo := repository.NewMockSourceRepo()
	caps := append([]string(nil), p.Index.Capabilities...)
	sourceRepo.GetCapabilitySummariesFunc = func(context.Context, string) ([]string, error) { return caps, nil }

	svc := services.NewChatService(
		planRepo, convRepo, repository.NewMockAnalyticsRepo(), repository.NewMockActionRepo(),
		sourceRepo, repository.NewMockHandoffRepo(), p.Factory, p.Embedder, p.Vector, nil, p.Log,
	)
	svc.SyncAnalytics = true
	svc.Retrieval = p.Retrieval
	return svc
}

// Generation is the outcome of one end-to-end chat turn.
type Generation struct {
	Answer string `json:"answer"`
	Tier   string `json:"tier"`
	// ContextChunks are the chunks ProcessChat placed in the prompt, captured
	// from the conversation repository's SaveMessageSources call.
	ContextChunks []RetrievedChunk `json:"context_chunks"`
	// Context is the text of those chunks, joined as the product joins them.
	Context string `json:"-"`
	// TokensUsed is the product's own token accounting for the turn.
	TokensUsed int           `json:"tokens_used"`
	Latency    time.Duration `json:"latency_ns"`
}

// ContextDocs returns the source IDs that reached the LLM context.
func (g Generation) ContextDocs() []string { return sourceIDs(g.ContextChunks) }

// Generate runs one single-turn conversation through
// services.ChatService.ProcessChat with in-memory repositories. A fresh
// service and fresh repositories are built per call so concurrent cases share
// no mutable state (the repository mocks record calls without locking).
func (p *Pipeline) Generate(ctx context.Context, question, sessionID string) (Generation, error) {
	var (
		mu       sync.Mutex
		captured []models.ChunkMetadata
	)
	svc := p.newService(func(sources []models.ChunkMetadata) {
		mu.Lock()
		captured = append([]models.ChunkMetadata(nil), sources...)
		mu.Unlock()
	})

	bot := p.Bot
	start := time.Now()
	res, err := svc.ProcessChat(ctx, models.ChatRequest{Message: question, SessionID: sessionID}, &bot, p.RAGConfig())
	latency := time.Since(start)
	if err != nil {
		return Generation{}, fmt.Errorf("process chat: %w", err)
	}

	mu.Lock()
	defer mu.Unlock()
	g := Generation{Answer: res.Response, Tier: res.ConfidenceTier, TokensUsed: res.TokensUsed, Latency: latency}
	parts := make([]string, 0, len(captured))
	for _, m := range captured {
		g.ContextChunks = append(g.ContextChunks, RetrievedChunk{SourceID: m.SourceID, ChunkIndex: m.ChunkIndex, Score: m.Score})
		if t, ok := p.Index.ChunkText(m.SourceID, m.ChunkIndex); ok {
			parts = append(parts, strings.TrimSpace(t))
		}
	}
	g.Context = strings.Join(parts, "\n---\n")
	return g, nil
}
