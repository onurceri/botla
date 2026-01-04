package handlers

import (
	"database/sql"

	"github.com/onurceri/botla-app/internal/processing"
	"github.com/onurceri/botla-app/internal/rag"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/storage"
	"github.com/onurceri/botla-app/pkg/urlutil"
)

// SourcesHandlers handles all source-related HTTP endpoints
type SourcesHandlers struct {
	DB               *sql.DB
	Queue            *processing.SourceQueue
	Storage          storage.StorageService
	QdrantClient     rag.VectorClient
	Log              *logger.Logger
	WorkspaceService *services.WorkspaceService
	OrgService       *services.OrganizationService
	SSRFValidator    *urlutil.SSRFValidator
	PlanRepo         repository.PlanRepository
	SourceRepo       repository.SourceRepository
	UsageRepo        repository.UsageRepository
	ChatbotRepo      repository.ChatbotRepository
}
