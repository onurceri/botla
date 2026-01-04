package services

import (
	"context"
	"database/sql"

	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/logger"
)

type RAGService struct {
	DB  *sql.DB
	RAG rag.RAGSubsystem
	Log *logger.Logger
}

func NewRAGService(db *sql.DB, ragSubsystem rag.RAGSubsystem, log *logger.Logger) *RAGService {
	return &RAGService{
		DB:  db,
		RAG: ragSubsystem,
		Log: log,
	}
}

func (s *RAGService) DeleteBotVectors(ctx context.Context, chatbotID string) error {
	if s.RAG == nil {
		return nil
	}

	rows, err := s.DB.QueryContext(ctx, `SELECT id FROM data_sources WHERE chatbot_id = $1 AND deleted_at IS NULL`, chatbotID)
	if err != nil {
		return pkgerrors.Wrapf(err, "querying source IDs")
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var sourceID string
		if err := rows.Scan(&sourceID); err != nil {
			continue
		}
		if err := s.RAG.DeleteBySource(ctx, sourceID); err != nil {
			if s.Log != nil {
				s.Log.Warn("delete_source_vectors_failed", map[string]any{"source_id": sourceID, "error": err.Error()})
			}
		}
	}

	if err := rows.Err(); err != nil {
		return pkgerrors.Wrapf(err, "iterating rows")
	}
	return nil
}

func (s *RAGService) DeleteSourceVectors(ctx context.Context, sourceID string) error {
	if s.RAG == nil {
		return nil
	}
	if err := s.RAG.DeleteBySource(ctx, sourceID); err != nil {
		return pkgerrors.Wrapf(err, "deleting source vectors")
	}
	return nil
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
