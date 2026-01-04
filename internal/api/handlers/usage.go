package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	pkgerrors "github.com/onurceri/botla-app/pkg/errors"
	"github.com/onurceri/botla-app/pkg/middleware"
)

// UsageHandlers handles usage-related endpoints
type UsageHandlers struct {
	UserRepo    repository.UserRepository
	ChatbotRepo repository.ChatbotRepository
	UsageRepo   repository.UsageRepository
}

// GetUsage handles GET /me/usage endpoint
func (h *UsageHandlers) GetUsage(w http.ResponseWriter, r *http.Request) {
	uid, ok := middleware.UserIDFromContext(r.Context())
	if !ok || uid == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Verify user exists
	u, err := h.UserRepo.GetByID(r.Context(), uid)
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
		chatbotsCount, err = h.UsageRepo.CountChatbotsByWorkspace(ctx, workspaceID)
	} else {
		chatbotsCount, err = h.UsageRepo.CountChatbotsByUserID(ctx, userID)
	}
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "count chatbots")
	}

	filesCount, err := h.UsageRepo.GetFileCountByUserID(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get file count")
	}
	urlsCount, err := h.UsageRepo.GetURLCountByUserID(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get url count")
	}
	tokensUsed, err := h.UsageRepo.GetMonthlyTokenUsage(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get monthly token usage")
	}
	storageUsedMB, err := h.UsageRepo.GetStorageUsedMBByUserID(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get storage used")
	}
	usedIngestions, usedEmbedTokens, err := h.UsageRepo.GetMonthlyIngestionUsage(ctx, userID, time.Now())
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get monthly ingestion usage")
	}
	maxFilesBot, err := h.UsageRepo.GetMaxFileCountInAnyBot(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get max file count")
	}
	maxURLsBot, err := h.UsageRepo.GetMaxURLCountInAnyBot(ctx, userID)
	if err != nil {
		return models.Usage{}, pkgerrors.Wrapf(err, "get max url count")
	}
	refreshCount, err := h.UsageRepo.GetMonthlyRefreshCount(ctx, userID, time.Now())
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
