package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	pkgerrors "github.com/onurceri/botla-co/pkg/errors"
	"github.com/onurceri/botla-co/pkg/middleware"
)

// UsageHandlers handles usage-related endpoints
type UsageHandlers struct {
	DB *sql.DB
}

// GetUsage handles GET /me/usage endpoint
func (h *UsageHandlers) GetUsage(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Verify user exists
	u, err := db.GetUserByID(r.Context(), h.DB, uid)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if u == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Check for workspace context from header (same as chatbot list)
	wsID, _ := middleware.WorkspaceIDFromContext(r.Context())

	usage, err := h.getUserUsage(r.Context(), u.ID, wsID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	api.WriteJSON(w, http.StatusOK, usage)
}

// getUserUsage retrieves all usage statistics for a user
func (h *UsageHandlers) getUserUsage(ctx context.Context, userID string, workspaceID string) (models.Usage, error) {
	var chatbotsCount int
	var err error

	// Count chatbots based on workspace context (to match dashboard)
	if workspaceID != "" {
		chatbotsCount, err = db.CountChatbotsByWorkspace(ctx, h.DB, workspaceID)
	} else {
		chatbotsCount, err = db.CountChatbotsByUserID(ctx, h.DB, userID)
	}
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "count chatbots")
	}

	filesCount, err := db.GetFileCountByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get file count")
	}
	urlsCount, err := db.GetURLCountByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get url count")
	}
	tokensUsed, err := db.GetMonthlyTokenUsage(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get monthly token usage")
	}
	storageUsedMB, err := db.GetStorageUsedMBByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get storage used")
	}
	usedIngestions, usedEmbedTokens, err := db.GetMonthlyIngestionUsage(ctx, h.DB, userID, time.Now())
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get monthly ingestion usage")
	}
	maxFilesBot, err := db.GetMaxFileCountInAnyBot(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get max file count")
	}
	maxURLsBot, err := db.GetMaxURLCountInAnyBot(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get max url count")
	}
	refreshCount, err := db.GetMonthlyRefreshCount(ctx, h.DB, userID, time.Now())
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get monthly refresh count")
	}

	return models.Usage{
		ChatbotsCount:            chatbotsCount,
		FilesCount:               filesCount,
		MaxFilesCountInOneBot:    maxFilesBot,
		StorageUsedMB:            storageUsedMB,
		URLsCount:                urlsCount,
		MaxURLsCountInOneBot:     maxURLsBot,
		TokensUsed:               tokensUsed,
		IngestionsUsed:           usedIngestions,
		IngestionEmbeddingTokens: usedEmbedTokens,
		RefreshCount:             refreshCount,
	}, nil
}
