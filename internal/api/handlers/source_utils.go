package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

// checkIngestionQuota validates monthly ingestion quota
func (h *SourcesHandlers) checkIngestionQuota(r *http.Request, userID string, plan *models.Plan) error {
	available := h.getAvailableIngestionCount(r, userID, plan)
	if available <= 0 {
		return &quotaError{"Monthly ingestion limit exceeded"}
	}
	return nil
}

// getAvailableIngestionCount returns how many more ingestions the user can perform this month
func (h *SourcesHandlers) getAvailableIngestionCount(r *http.Request, userID string, plan *models.Plan) int {
	usedSources, _, _ := db.GetMonthlyIngestionUsage(r.Context(), h.DB, userID, time.Now())
	maxIngest := plan.Config.MaxMonthlyIngestions
	if maxIngest <= 0 {
		maxIngest = 50
	}
	if usedSources >= maxIngest {
		return 0
	}
	return maxIngest - usedSources
}

// persistAndEnqueueInternal saves the data source to database and enqueues for processing without writing response
func (h *SourcesHandlers) persistAndEnqueueInternal(r *http.Request, ds *models.DataSource) (string, error) {
	newID, err := db.CreateDataSource(r.Context(), h.DB, ds)
	if err != nil {
		h.logError("source_create_error", map[string]any{"error": err.Error(), "chatbot_id": ds.ChatbotID, "source_type": ds.SourceType})
		return "", fmt.Errorf("create data source: %w", err)
	}
	if h.Queue != nil {
		h.Queue.Enqueue(newID)
	}
	return newID, nil
}

// checkCooldown validates if enough time has passed since the last action
func (h *SourcesHandlers) checkCooldown(r *http.Request, lastActionTime *time.Time, plan *models.Plan) (time.Duration, bool) {
	cdMin := plan.Config.MinReAddCooldownMinutes
	if cdMin <= 0 || lastActionTime == nil {
		return 0, true
	}

	elapsed := time.Since(*lastActionTime)
	cooldown := time.Duration(cdMin) * time.Minute
	if elapsed < cooldown {
		return cooldown - elapsed, false
	}
	return 0, true
}

// checkStorageQuota validates total storage quota
func (h *SourcesHandlers) checkStorageQuota(r *http.Request, userID string, sizeBytes int, plan *models.Plan) error {
	limitMB := plan.Config.Files.TotalStorageMB
	if limitMB > 0 {
		usedMB, _ := db.GetStorageUsedMBByUserID(r.Context(), h.DB, userID)
		newMB := sizeBytes / (1 << 20)
		if usedMB+newMB > limitMB {
			return &quotaError{"Total storage limit exceeded"}
		}
	}
	return nil
}

// quotaError represents a quota limit error
type quotaError struct {
	msg string
}

func (e *quotaError) Error() string { return e.msg }

// computeHash returns SHA256 hash of data as hex string
func computeHash(data []byte) string {
	hsum := sha256.Sum256(data)
	return hex.EncodeToString(hsum[:])
}

// logError logs an error if logger is available
func (h *SourcesHandlers) logError(event string, data map[string]any) {
	if h.Log != nil {
		h.Log.Error(event, data)
	}
}

// logWarn logs a warning if logger is available
func (h *SourcesHandlers) logWarn(event string, data map[string]any) {
	if h.Log != nil {
		h.Log.Warn(event, data)
	}
}

// isPDFContentType checks if content type or filename indicates PDF
func isPDFContentType(ct, name string) bool {
	if ct == "application/pdf" {
		return true
	}
	return strings.HasSuffix(name, ".pdf")
}
