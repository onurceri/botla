package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(usage); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
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
		return models.Usage{}, err
	}

	filesCount, err := db.GetFileCountByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	urlsCount, err := db.GetURLCountByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	tokensUsed, err := db.GetMonthlyTokenUsage(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	storageUsedMB, err := db.GetStorageUsedMBByUserID(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	usedIngestions, usedEmbedTokens, err := db.GetMonthlyIngestionUsage(ctx, h.DB, userID, time.Now())
	if err != nil {
		return models.Usage{}, err
	}
	maxFilesBot, err := db.GetMaxFileCountInAnyBot(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	maxURLsBot, err := db.GetMaxURLCountInAnyBot(ctx, h.DB, userID)
	if err != nil {
		return models.Usage{}, err
	}
	refreshCount, err := db.GetMonthlyRefreshCount(ctx, h.DB, userID, time.Now())
	if err != nil {
		return models.Usage{}, err
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
