package handlers

import (
	"database/sql"

	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/storage"
)

// SourcesHandlers handles all source-related HTTP endpoints
type SourcesHandlers struct {
	DB           *sql.DB
	Queue        *processing.SourceQueue
	Storage      storage.StorageService
	QdrantClient rag.VectorClient
	Log          *logger.Logger
}
