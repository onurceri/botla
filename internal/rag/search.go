package rag

import (
	"context"

	"github.com/onurceri/botla-app/internal/models"
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

// SearchContextTiered performs a tiered similarity search for one query
// embedding using ThresholdConfig. It is SearchContextMulti with a single
// query and no over-fetch or per-document cap.
func SearchContextTiered(ctx context.Context, vc VectorClient, queryEmbedding []float32, chatbotID string, limitTopK int, limitMaxTokens int, thresholdConfig *models.ThresholdConfig) (*TieredSearchResult, error) {
	return SearchContextMulti(ctx, vc, [][]float32{queryEmbedding}, chatbotID, SearchOptions{
		TopK:             limitTopK,
		MaxContextTokens: limitMaxTokens,
		Thresholds:       thresholdConfig,
	})
}
