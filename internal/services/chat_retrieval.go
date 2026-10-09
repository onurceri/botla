package services

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/rag"
)

// RetrievalOutcome is the result of RetrieveContext.
type RetrievalOutcome struct {
	// Result is the tiered context selection (never nil).
	Result *rag.TieredSearchResult
	// SubQueries holds the standalone queries that were searched when query
	// decomposition split the question; it is nil for single-part questions.
	SubQueries []string
	// LLMTokens is the token usage of the decomposition call, if any. It is
	// added to the turn's token count so quota accounting includes it.
	LLMTokens int
}

// RetrieveContext gathers the RAG context for a question with the service's
// retrieval strategy (s.Retrieval). With decomposition enabled, one small LLM
// call (the bot's own model, through the regular client factory) splits a
// multi-part question into sub-queries; the question and each sub-query are
// embedded and searched, and the rankings are fused by rag.SearchContextMulti
// under the same top-K and token budget. Single-part questions, and any decomposition failure, fall
// back to searching the question verbatim, exactly as without decomposition.
//
// The eval harness calls this too, so the evals measure the product path.
func (s *ChatService) RetrieveContext(ctx context.Context, bot *models.Chatbot, question string, ragCfg models.RAGConfig, thresholds *models.ThresholdConfig) RetrievalOutcome {
	low := RetrievalOutcome{Result: &rag.TieredSearchResult{Tier: rag.TierLow}}
	embedder := s.getEmbedder()
	qc := s.getQdrantClient()
	if embedder == nil || qc == nil {
		return low
	}

	// Embed the question while (optionally) decomposing it, so single-part
	// questions pay only the decomposition latency, not an extra embedding.
	type embedResult struct {
		vec []float32
		err error
	}
	embedCh := make(chan embedResult, 1)
	go func() {
		v, err := embedder.CreateEmbedding(ctx, question)
		embedCh <- embedResult{v, err}
	}()

	var out RetrievalOutcome
	var queryEmbeddings [][]float32
	if s.Retrieval.Decompose {
		out.SubQueries, out.LLMTokens = s.decompose(ctx, bot, question)
		if len(out.SubQueries) > 1 {
			embs, err := embedder.CreateEmbeddingsBatch(ctx, out.SubQueries)
			if err == nil && len(embs) == len(out.SubQueries) {
				queryEmbeddings = embs // the original question's ranking is prepended below
			} else {
				s.warn("subquery_embedding_error", bot.ID, err)
				out.SubQueries = nil
			}
		} else {
			out.SubQueries = nil
		}
	}
	orig := <-embedCh
	switch {
	case orig.err == nil:
		// The original question is always searched, and ranked first in the
		// fusion, so a poor split can add documents but cannot displace the
		// question's own best match.
		queryEmbeddings = append([][]float32{orig.vec}, queryEmbeddings...)
	case queryEmbeddings == nil:
		low.LLMTokens = out.LLMTokens
		return low
	}

	result, err := rag.SearchContextMulti(ctx, qc, queryEmbeddings, bot.ID, rag.SearchOptions{
		TopK:             ragCfg.TopK,
		MaxContextTokens: ragCfg.MaxContextTokens,
		Thresholds:       thresholds,
	})
	if err != nil {
		s.warn("tiered_search_error", bot.ID, err)
	}
	if result == nil {
		result = &rag.TieredSearchResult{Tier: rag.TierLow}
	}
	out.Result = result
	return out
}

// decompose runs rag.DecomposeQuery with the bot's model. Failures are logged
// and yield the original question.
func (s *ChatService) decompose(ctx context.Context, bot *models.Chatbot, question string) ([]string, int) {
	client, model, err := s.getToolsClient(bot.Model)
	if err != nil {
		s.warn("query_decomposition_error", bot.ID, err)
		return []string{question}, 0
	}
	queries, tokens, err := rag.DecomposeQuery(ctx, client, model, question)
	if err != nil {
		s.warn("query_decomposition_error", bot.ID, err)
	}
	return queries, tokens
}

func (s *ChatService) warn(event, chatbotID string, err error) {
	if s.Log == nil {
		return
	}
	fields := map[string]any{"chatbot_id": chatbotID}
	if err != nil {
		fields["error"] = err.Error()
	}
	s.Log.Warn(event, fields)
}
