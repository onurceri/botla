package rag

import (
	"context"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSearchContextTiered_Unit(t *testing.T) {
	ctx := context.Background()
	mockVC := &MockVectorClient{}

	queryEmbedding := []float32{0.1, 0.2, 0.3}
	chatbotID := "test-bot"

	t.Run("High Tier Match", func(t *testing.T) {
		items := []SearchResult{
			{
				ID:    "1",
				Score: 0.9,
				Payload: EmbeddingPayload{
					OriginalText: "High quality info",
					SourceID:     "s1",
					SourceType:   "file",
				},
			},
		}

		mockVC.On("SearchSimilar", ctx, queryEmbedding, chatbotID, 5).Return(items, nil).Once()

		cfg := &models.ThresholdConfig{
			HighThreshold:   0.8,
			MediumThreshold: 0.4,
		}

		result, err := SearchContextTiered(ctx, mockVC, queryEmbedding, chatbotID, 5, 1000, cfg)

		assert.NoError(t, err)
		assert.Equal(t, TierHigh, result.Tier)
		assert.Contains(t, result.ContextText, "High quality info")
		mockVC.AssertExpectations(t)
	})

	t.Run("Medium Tier Match", func(t *testing.T) {
		items := []SearchResult{
			{
				ID:    "1",
				Score: 0.6,
				Payload: EmbeddingPayload{
					OriginalText: "Medium quality info",
					SourceID:     "s1",
					SourceType:   "file",
				},
			},
		}

		mockVC.On("SearchSimilar", ctx, queryEmbedding, chatbotID, 5).Return(items, nil).Once()

		cfg := &models.ThresholdConfig{
			HighThreshold:   0.8,
			MediumThreshold: 0.4,
		}

		result, err := SearchContextTiered(ctx, mockVC, queryEmbedding, chatbotID, 5, 1000, cfg)

		assert.NoError(t, err)
		assert.Equal(t, TierMedium, result.Tier)
		assert.Contains(t, result.ContextText, "Medium quality info")
		mockVC.AssertExpectations(t)
	})

	t.Run("Low Tier Match", func(t *testing.T) {
		items := []SearchResult{
			{
				ID:    "1",
				Score: 0.2,
				Payload: EmbeddingPayload{
					OriginalText: "Low quality info",
				},
			},
		}

		mockVC.On("SearchSimilar", ctx, queryEmbedding, chatbotID, 5).Return(items, nil).Once()

		cfg := &models.ThresholdConfig{
			HighThreshold:   0.8,
			MediumThreshold: 0.4,
		}

		result, err := SearchContextTiered(ctx, mockVC, queryEmbedding, chatbotID, 5, 1000, cfg)

		assert.NoError(t, err)
		assert.Equal(t, TierLow, result.Tier)
		assert.Empty(t, result.ContextText)
		mockVC.AssertExpectations(t)
	})
}

func TestEmbeddingService_GenerateForSource_Unit(t *testing.T) {
	ctx := context.Background()
	mockEmb := &MockEmbeddingClient{}
	mockVC := &MockVectorClient{}

	chunks := []models.Chunk{
		{Text: "chunk 1", TokenCount: 10},
		{Text: "chunk 2", TokenCount: 10},
	}
	chatbotID := "bot-1"
	sourceID := "src-1"
	sourceType := "web"

	t.Run("Successful Generation", func(t *testing.T) {
		vectors := [][]float32{
			{0.1, 0.1},
			{0.2, 0.2},
		}

		mockEmb.On("CreateEmbeddingsBatch", mock.Anything, []string{"chunk 1", "chunk 2"}).Return(vectors, nil).Once()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, vectors[0], mock.Anything).Return(nil).Once()
		mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, vectors[1], mock.Anything).Return(nil).Once()

		svc := NewEmbeddingService(mockEmb, mockVC, nil)
		err := svc.GenerateForSource(ctx, chunks, chatbotID, sourceID, sourceType)

		assert.NoError(t, err)
		mockEmb.AssertExpectations(t)
		mockVC.AssertExpectations(t)
	})

	t.Run("Embedding Failure", func(t *testing.T) {
		mockEmb.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return(nil, assert.AnError).Twice() // Includes retry

		svc := NewEmbeddingService(mockEmb, mockVC, nil)
		err := svc.GenerateForSource(ctx, chunks, chatbotID, sourceID, sourceType)

		assert.Error(t, err)
		mockEmb.AssertExpectations(t)
	})
}
