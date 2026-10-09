package rag

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/onurceri/botla-app/internal/models"
)

// EnvQueryDecomposition turns query decomposition on or off (see RetrievalStrategyFromEnv).
const EnvQueryDecomposition = "RAG_QUERY_DECOMPOSITION"

// MaxSubQueries is the most sub-queries a question is split into.
const MaxSubQueries = 3

// RetrievalStrategy selects how context is gathered for a question. The zero
// value is the original single-query behaviour: one embedding, top-K hits,
// threshold filter, token budget.
type RetrievalStrategy struct {
	// Decompose splits multi-part questions into standalone sub-queries with
	// one small LLM call, searches each, and fuses the rankings. Single-part
	// questions are searched exactly as before.
	Decompose bool
}

// DefaultQueryDecomposition is used when RAG_QUERY_DECOMPOSITION is unset.
// It is off: on the eval suite decomposition raised multi-hop context recall
// but not answer pass rates, and it roughly doubles median turn latency (one
// extra LLM call before retrieval). See evals/README.md, "Experiments:
// multi-hop".
const DefaultQueryDecomposition = false

// RetrievalStrategyFromEnv reads the strategy from RAG_QUERY_DECOMPOSITION
// (a strconv.ParseBool value; unset or invalid means DefaultQueryDecomposition).
func RetrievalStrategyFromEnv() RetrievalStrategy {
	s := RetrievalStrategy{Decompose: DefaultQueryDecomposition}
	if v, err := strconv.ParseBool(strings.TrimSpace(os.Getenv(EnvQueryDecomposition))); err == nil {
		s.Decompose = v
	}
	return s
}

// SearchOptions configures SearchContextMulti.
type SearchOptions struct {
	// TopK is the number of chunks that may enter the context (<= 0: RAG_TOPK or 5).
	TopK int
	// MaxContextTokens is the context token budget (<= 0: RAG_MAX_CONTEXT_TOKENS or 2000).
	MaxContextTokens int
	// Thresholds decide the confidence tier (nil: defaults).
	Thresholds *models.ThresholdConfig
}

// SearchContextMulti is the tiered context search for one or more query
// embeddings (the question and its sub-queries). Each query fetches TopK
// candidates; the per-query rankings are fused round-robin (rank 1 of every
// query first, then rank 2, ...) so each part of a multi-part question gets
// its best chunk into the context before any part gets its second. The tier is
// decided by the best score across all queries. Context selection then keeps
// at most TopK chunks above the tier's threshold and stops at the token
// budget.
//
// With a single embedding this is exactly the original single-query
// SearchContextTiered behaviour.
func SearchContextMulti(ctx context.Context, vc VectorClient, queryEmbeddings [][]float32, chatbotID string, opts SearchOptions) (*TieredSearchResult, error) {
	if vc == nil || (reflect.ValueOf(vc).Kind() == reflect.Ptr && reflect.ValueOf(vc).IsNil()) {
		return nil, fmt.Errorf("vector client is nil")
	}
	var queries [][]float32
	for _, e := range queryEmbeddings {
		if len(e) > 0 {
			queries = append(queries, e)
		}
	}
	if len(queries) == 0 || chatbotID == "" {
		return &TieredSearchResult{Tier: TierLow}, nil
	}
	thresholds := opts.Thresholds
	if thresholds == nil {
		thresholds = models.DefaultThresholdConfig()
	}
	topK := resolveTopK(opts.TopK)

	rankings := make([][]SearchResult, 0, len(queries))
	for _, q := range queries {
		items, err := vc.SearchSimilar(ctx, q, chatbotID, topK)
		if err != nil {
			return nil, fmt.Errorf("search similar failed: %w", err)
		}
		sort.SliceStable(items, func(i, j int) bool { return items[i].Score > items[j].Score })
		rankings = append(rankings, items)
	}
	items := FuseRoundRobin(rankings)
	if len(items) == 0 {
		return &TieredSearchResult{Tier: TierLow}, nil
	}

	allMetas := make([]models.ChunkMetadata, 0, len(items))
	highestScore := items[0].Score
	for _, it := range items {
		allMetas = append(allMetas, chunkMeta(it))
		if it.Score > highestScore {
			highestScore = it.Score
		}
	}

	var tier ContextTier
	switch {
	case highestScore >= thresholds.HighThreshold:
		tier = TierHigh
	case highestScore >= thresholds.MediumThreshold:
		tier = TierMedium
	default:
		return &TieredSearchResult{Tier: TierLow, AllChunks: allMetas, HighestScore: highestScore}, nil
	}
	// Both admitted tiers keep chunks down to the medium threshold.
	effectiveThreshold := thresholds.MediumThreshold

	maxCtx := resolveMaxContextTokens(opts.MaxContextTokens)
	var (
		parts      []string
		usedChunks []models.ChunkMetadata
		tokens     int
		totalScore float64
	)
	for _, it := range items {
		if len(usedChunks) >= topK {
			break
		}
		if it.Score < effectiveThreshold {
			continue
		}
		t := strings.TrimSpace(it.Payload.OriginalText)
		if t == "" {
			continue
		}
		next := t
		if len(parts) > 0 {
			next = "\n---\n" + next
		}
		nt := CountTokens(next, "tr")
		if tokens+nt > maxCtx {
			break
		}
		parts = append(parts, next)
		tokens += nt
		totalScore += it.Score
		usedChunks = append(usedChunks, chunkMeta(it))
	}

	if len(parts) == 0 {
		return &TieredSearchResult{Tier: TierLow, AllChunks: allMetas, HighestScore: highestScore}, nil
	}
	// Note: the context intro prefix is added by the chat service using langconfig.
	return &TieredSearchResult{
		ContextText:  strings.Join(parts, ""),
		Chunks:       usedChunks,
		AllChunks:    allMetas,
		Tier:         tier,
		HighestScore: highestScore,
		AverageScore: totalScore / float64(len(usedChunks)),
	}, nil
}

// FuseRoundRobin merges per-query rankings by taking rank 1 of every query,
// then rank 2, and so on, skipping chunks already taken (same source and chunk
// index). A duplicate keeps its first position and the best score seen. A
// single ranking is returned unchanged (in particular not de-duplicated), so
// single-query search behaves exactly as before.
func FuseRoundRobin(rankings [][]SearchResult) []SearchResult {
	if len(rankings) == 1 {
		return rankings[0]
	}
	type key struct {
		src string
		idx int
	}
	pos := map[key]int{}
	var out []SearchResult
	for rank := 0; ; rank++ {
		progressed := false
		for _, r := range rankings {
			if rank >= len(r) {
				continue
			}
			progressed = true
			it := r[rank]
			k := key{it.Payload.SourceID, it.Payload.ChunkIndex}
			if i, seen := pos[k]; seen {
				if it.Score > out[i].Score {
					out[i].Score = it.Score
				}
				continue
			}
			pos[k] = len(out)
			out = append(out, it)
		}
		if !progressed {
			return out
		}
	}
}

func chunkMeta(it SearchResult) models.ChunkMetadata {
	return models.ChunkMetadata{
		SourceID:   it.Payload.SourceID,
		SourceType: it.Payload.SourceType,
		ChunkIndex: it.Payload.ChunkIndex,
		Score:      it.Score,
	}
}

func resolveTopK(topK int) int {
	if topK > 0 {
		return topK
	}
	if n, err := strconv.Atoi(os.Getenv("RAG_TOPK")); err == nil && n > 0 {
		return n
	}
	return 5
}

func resolveMaxContextTokens(maxTokens int) int {
	if maxTokens > 0 {
		return maxTokens
	}
	if n, err := strconv.Atoi(os.Getenv("RAG_MAX_CONTEXT_TOKENS")); err == nil && n > 0 {
		return n
	}
	return 2000
}
