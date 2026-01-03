package qdrant

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/integration"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealQdrant_Connection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Qdrant integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	t.Run("client is created", func(t *testing.T) {
		assert.NotNil(t, env.Qdrant)
	})

	t.Run("collection exists or can be created", func(t *testing.T) {
		ctx := t.Context()

		err := env.Qdrant.EnsureEmbeddingsCollection(ctx)
		require.NoError(t, err)
	})
}

func TestRealQdrant_CollectionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Qdrant integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := t.Context()
	collectionName := "test-collection-" + uuid.New().String()[:8]

	t.Run("create collection", func(t *testing.T) {
		client, err := rag.NewQdrantClient(&rag.QdrantConfig{
			URL:            env.Cfg.QDRANT_URL,
			APIKey:         env.Cfg.QDRANT_API_KEY,
			Timeout:        15 * time.Second,
			CollectionName: collectionName,
		})
		require.NoError(t, err)

		err = client.EnsureEmbeddingsCollection(ctx)
		assert.NoError(t, err)
	})

	t.Run("upsert vectors", func(t *testing.T) {
		client, err := rag.NewQdrantClient(&rag.QdrantConfig{
			URL:            env.Cfg.QDRANT_URL,
			APIKey:         env.Cfg.QDRANT_API_KEY,
			Timeout:        15 * time.Second,
			CollectionName: collectionName,
		})
		require.NoError(t, err)
		require.NoError(t, client.EnsureEmbeddingsCollection(ctx))

		pointID := uuid.New().String()
		payload := rag.EmbeddingPayload{
			ChatbotID:    "test-chatbot-id",
			SourceID:     "test-source-id",
			ChunkIndex:   0,
			OriginalText: "test content",
			SourceType:   "text",
		}

		vector := make([]float32, 1536)
		for i := range vector {
			vector[i] = float32(i%256) / 255.0
		}

		err = client.UpsertEmbedding(ctx, pointID, vector, payload)
		assert.NoError(t, err)
	})

	t.Run("search vectors", func(t *testing.T) {
		client, err := rag.NewQdrantClient(&rag.QdrantConfig{
			URL:            env.Cfg.QDRANT_URL,
			APIKey:         env.Cfg.QDRANT_API_KEY,
			Timeout:        15 * time.Second,
			CollectionName: collectionName,
		})
		require.NoError(t, err)
		require.NoError(t, client.EnsureEmbeddingsCollection(ctx))

		queryVector := make([]float32, 1536)
		for i := range queryVector {
			queryVector[i] = float32(i%256) / 255.0
		}

		results, err := client.SearchSimilar(ctx, queryVector, "test-chatbot-id", 5)
		assert.NoError(t, err)
		assert.NotNil(t, results)
	})
}

func TestRealQdrant_ConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Qdrant integration test in short mode")
	}

	env := integration.SetupRealServices(t)
	t.Cleanup(env.Cleanup)

	ctx := t.Context()
	collectionName := "concurrent-test-" + uuid.New().String()[:8]

	testClient, err := rag.NewQdrantClient(&rag.QdrantConfig{
		URL:            env.Cfg.QDRANT_URL,
		APIKey:         env.Cfg.QDRANT_API_KEY,
		Timeout:        15 * time.Second,
		CollectionName: collectionName,
	})
	require.NoError(t, err)
	require.NoError(t, testClient.EnsureEmbeddingsCollection(ctx))

	t.Run("concurrent upserts", func(t *testing.T) {
		done := make(chan error, 100)

		for i := 0; i < 100; i++ {
			go func(id int) {
				pointID := uuid.New().String()
				payload := rag.EmbeddingPayload{
					ChatbotID:    "concurrent-test",
					SourceID:     "concurrent-source",
					ChunkIndex:   id,
					OriginalText: "test content",
					SourceType:   "text",
				}

				vector := make([]float32, 1536)
				for j := range vector {
					vector[j] = float32((id*100+j)%256) / 255.0
				}

				err := testClient.UpsertEmbedding(ctx, pointID, vector, payload)
				done <- err
			}(i)
		}

		errors := 0
		for i := 0; i < 100; i++ {
			if err := <-done; err != nil {
				errors++
			}
		}

		assert.Equal(t, 0, errors, "All upserts should succeed")
	})

	t.Run("concurrent searches", func(t *testing.T) {
		queryVector := make([]float32, 1536)
		for i := range queryVector {
			queryVector[i] = float32(i%256) / 255.0
		}

		done := make(chan bool, 100)
		errors := 0

		for i := 0; i < 100; i++ {
			go func() {
				_, err := testClient.SearchSimilar(ctx, queryVector, "concurrent-test", 5)
				if err != nil {
					done <- false
					return
				}
				done <- true
			}()
		}

		successCount := 0
		for i := 0; i < 100; i++ {
			if <-done {
				successCount++
			} else {
				errors++
			}
		}

		assert.Equal(t, 100, successCount, "All 100 searches should succeed")
		assert.Equal(t, 0, errors, "No searches should fail")
	})
}
