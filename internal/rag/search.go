package rag

import (
	"context"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/onurceri/botla-co/internal/models"
)

// ContextTier represents the confidence level of search results
type ContextTier string

const (
	TierHigh   ContextTier = "high"   // Strong match - normal RAG flow
	TierMedium ContextTier = "medium" // Weak match - RAG with warning
	TierLow    ContextTier = "low"    // No match - fallback mode
)

// TieredSearchResult contains search results with tier information
type TieredSearchResult struct {
	ContextText  string                 // Combined context from matching chunks
	Chunks       []models.ChunkMetadata // Metadata of used chunks
	AllChunks    []models.ChunkMetadata // All chunks found (for sources_used)
	Tier         ContextTier            // Confidence tier
	HighestScore float64                // Highest similarity score
	AverageScore float64                // Average score of used chunks
}

// SearchContextTiered performs a tiered similarity search using ThresholdConfig
func SearchContextTiered(queryEmbedding []float32, chatbotID string, limitTopK int, limitMaxTokens int, thresholdConfig *models.ThresholdConfig) (*TieredSearchResult, error) {
	if len(queryEmbedding) == 0 || chatbotID == "" {
		return &TieredSearchResult{Tier: TierLow}, nil
	}

	// Use defaults if config is nil
	if thresholdConfig == nil {
		thresholdConfig = models.DefaultThresholdConfig()
	}

	qc, err := NewQdrantClientFromEnv()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	topK := limitTopK
	if topK <= 0 {
		topK = 5
		if v := os.Getenv("RAG_TOPK"); v != "" {
			if n, errEnv := strconv.Atoi(v); errEnv == nil && n > 0 {
				topK = n
			}
		}
	}

	items, err := qc.SearchSimilar(ctx, queryEmbedding, chatbotID, topK)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return &TieredSearchResult{Tier: TierLow}, nil
	}

	// Sort by score descending
	sort.Slice(items, func(i, j int) bool { return items[i].Score > items[j].Score })

	// Collect all chunks metadata (for sources_used in response)
	var allMetas []models.ChunkMetadata
	for _, it := range items {
		allMetas = append(allMetas, models.ChunkMetadata{
			SourceID:   it.Payload.SourceID,
			SourceType: it.Payload.SourceType,
			ChunkIndex: it.Payload.ChunkIndex,
			Score:      it.Score,
		})
	}

	// Determine highest score
	highestScore := items[0].Score

	var tier ContextTier
	var effectiveThreshold float64
	switch {
	case highestScore >= thresholdConfig.HighThreshold:
		tier = TierHigh
		effectiveThreshold = thresholdConfig.MediumThreshold
	case highestScore >= thresholdConfig.MediumThreshold:
		tier = TierMedium
		effectiveThreshold = thresholdConfig.MediumThreshold
	default:
		return &TieredSearchResult{
			Tier:         TierLow,
			AllChunks:    allMetas,
			HighestScore: highestScore,
		}, nil
	}

	// Build context from chunks that pass the effective threshold
	maxCtx := limitMaxTokens
	if maxCtx <= 0 {
		maxCtx = 2000
		if v := os.Getenv("RAG_MAX_CONTEXT_TOKENS"); v != "" {
			if n, errEnv := strconv.Atoi(v); errEnv == nil && n > 0 {
				maxCtx = n
			}
		}
	}

	var parts []string
	var usedChunks []models.ChunkMetadata
	var tokens int
	var totalScore float64

	for _, it := range items {
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
		usedChunks = append(usedChunks, models.ChunkMetadata{
			SourceID:   it.Payload.SourceID,
			SourceType: it.Payload.SourceType,
			ChunkIndex: it.Payload.ChunkIndex,
			Score:      it.Score,
		})
	}

	// Calculate average score
	var avgScore float64
	if len(usedChunks) > 0 {
		avgScore = totalScore / float64(len(usedChunks))
	}

	// If no chunks passed, return low tier
	if len(parts) == 0 {
		return &TieredSearchResult{
			Tier:         TierLow,
			AllChunks:    allMetas,
			HighestScore: highestScore,
		}, nil
	}

	body := strings.Join(parts, "")
	// Note: Context intro prefix is added by chat_service using langconfig

	return &TieredSearchResult{
		ContextText:  body,
		Chunks:       usedChunks,
		AllChunks:    allMetas,
		Tier:         tier,
		HighestScore: highestScore,
		AverageScore: avgScore,
	}, nil
}


