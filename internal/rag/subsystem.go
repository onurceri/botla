package rag

import (
	"context"
	"errors"

	"github.com/onurceri/botla-app/internal/models"
)

var (
	ErrNilEmbedder     = errors.New("rag subsystem: embedder is nil")
	ErrNilVectorClient = errors.New("rag subsystem: vector client is nil")
	ErrNilLLMClient    = errors.New("rag subsystem: llm client is nil")
)

type ChunkWithEmbedding struct {
	Chunk     string
	Embedding []float32
}

type SearchRequest struct {
	Query     string
	ChatbotID string
	Limit     int
}

type RAGSubsystem interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)
	Store(ctx context.Context, chunk ChunkWithEmbedding) error
	Search(ctx context.Context, req SearchRequest) ([]SearchResult, error)
	DeleteBySource(ctx context.Context, sourceID string) error
	Complete(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error)
	Ready() bool
}

type ragSubsystem struct {
	embedder EmbeddingClient
	vector   VectorClient
	llm      LLMClient
}

func NewRAGSubsystem(embedder EmbeddingClient, vector VectorClient, llm LLMClient) RAGSubsystem {
	return &ragSubsystem{
		embedder: embedder,
		vector:   vector,
		llm:      llm,
	}
}

func (r *ragSubsystem) Embed(ctx context.Context, text string) ([]float32, error) {
	if r.embedder == nil {
		return nil, ErrNilEmbedder
	}
	return r.embedder.CreateEmbedding(ctx, text)
}

func (r *ragSubsystem) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if r.embedder == nil {
		return nil, ErrNilEmbedder
	}
	return r.embedder.CreateEmbeddingsBatch(ctx, texts)
}

func (r *ragSubsystem) Store(ctx context.Context, chunk ChunkWithEmbedding) error {
	if r.vector == nil {
		return ErrNilVectorClient
	}
	return r.vector.UpsertEmbedding(ctx, chunk.Embedding, chunk.Embedding, EmbeddingPayload{
		OriginalText: chunk.Chunk,
	})
}

func (r *ragSubsystem) Search(ctx context.Context, req SearchRequest) ([]SearchResult, error) {
	if r.vector == nil {
		return nil, ErrNilVectorClient
	}
	embedding, err := r.embedder.CreateEmbedding(ctx, req.Query)
	if err != nil {
		return nil, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}
	return r.vector.SearchSimilar(ctx, embedding, req.ChatbotID, limit)
}

func (r *ragSubsystem) DeleteBySource(ctx context.Context, sourceID string) error {
	if r.vector == nil {
		return ErrNilVectorClient
	}
	return r.vector.DeleteBySourceID(ctx, sourceID)
}

func (r *ragSubsystem) Complete(ctx context.Context, params models.CompletionParams) (*models.CompletionResult, error) {
	if r.llm == nil {
		return nil, ErrNilLLMClient
	}
	return r.llm.CreateCompletion(ctx, params)
}

func (r *ragSubsystem) Ready() bool {
	return r.embedder != nil && r.vector != nil && r.llm != nil
}
