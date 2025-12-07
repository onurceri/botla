package models

// Chunk represents a text chunk with its token count, used for text processing.
type Chunk struct {
	Text       string
	TokenCount int
}

// ChunkMetadata contains metadata about a retrieved chunk from vector search.
type ChunkMetadata struct {
	SourceID   string
	SourceType string
	ChunkIndex int
	Score      float64
}
