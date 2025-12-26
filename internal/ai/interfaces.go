package ai

import "context"

// VectorStore abstracts vector database operations
type VectorStore interface {
	// EnsureCollection ensures the vector collection exists
	EnsureCollection(ctx context.Context) error

	// Upsert inserts or updates a vector with associated metadata
	Upsert(ctx context.Context, id interface{}, vector []float32, payload VectorPayload) error

	// Search performs similarity search with filters
	Search(ctx context.Context, vector []float32, filter SearchFilter, limit int) ([]SearchResult, error)

	// Delete removes vectors matching the filter
	Delete(ctx context.Context, filter DeleteFilter) error

	// Scroll retrieves vectors in pages with optional offset for pagination
	// Returns results, next offset (nil if no more pages), and error
	Scroll(ctx context.Context, filter SearchFilter, limit int, offset interface{}) ([]SearchResult, interface{}, error)
}

// Embedder abstracts text embedding generation
type Embedder interface {
	// Embed generates an embedding for a single text
	Embed(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch generates embeddings for multiple texts
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimension returns the dimensionality of embeddings
	Dimension() int
}
