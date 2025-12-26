package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
)

// RAGService provides operations related to RAG vector management
type RAGService struct {
	DB           *sql.DB
	VectorClient rag.VectorClient
	Log          *logger.Logger
}

// NewRAGService creates a new RAGService instance
func NewRAGService(db *sql.DB, vectorClient rag.VectorClient, log *logger.Logger) *RAGService {
	return &RAGService{
		DB:           db,
		VectorClient: vectorClient,
		Log:          log,
	}
}

// DeleteBotVectors deletes all vectors associated with a chatbot by deleting vectors for each source
func (s *RAGService) DeleteBotVectors(ctx context.Context, chatbotID string) error {
	if s.VectorClient == nil {
		return nil
	}

	// Get all source IDs for this chatbot
	rows, err := s.DB.QueryContext(ctx, `SELECT id FROM data_sources WHERE chatbot_id = $1 AND deleted_at IS NULL`, chatbotID)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sourceID string
		if err := rows.Scan(&sourceID); err != nil {
			continue
		}
		// Delete vectors for each source
		if err := s.VectorClient.DeleteBySourceID(ctx, sourceID); err != nil {
			if s.Log != nil {
				s.Log.Warn("delete_source_vectors_failed", map[string]any{"source_id": sourceID, "error": err.Error()})
			}
		}
	}

	return rows.Err()
}

// DeleteSourceVectors deletes all vectors associated with a specific source
func (s *RAGService) DeleteSourceVectors(ctx context.Context, sourceID string) error {
	if s.VectorClient == nil {
		return nil
	}
	return s.VectorClient.DeleteBySourceID(ctx, sourceID)
}

// Queue wraps SourceQueue to provide a simpler interface for handlers
type Queue struct {
	SourceQueue *processing.SourceQueue
}

// Enqueue adds a source ID to the processing queue
func (q *Queue) Enqueue(id string) error {
	if q == nil || q.SourceQueue == nil {
		return nil
	}
	q.SourceQueue.Enqueue(id)
	return nil
}
