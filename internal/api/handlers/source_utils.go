package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
)

// checkIngestionQuota validates monthly ingestion quota
func (h *SourcesHandlers) checkIngestionQuota(r *http.Request, userID string, plan *models.Plan) error {
	usedSources, _, _ := db.GetMonthlyIngestionUsage(r.Context(), h.DB, userID, time.Now())
	maxIngest := plan.Config.MaxMonthlyIngestions
	if maxIngest <= 0 {
		maxIngest = 50
	}
	if usedSources >= maxIngest {
		return &quotaError{"Monthly ingestion limit exceeded"}
	}
	return nil
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

// parseChatbotIDFromPath extracts chatbot ID from /api/v1/chatbots/:id/sources
func parseChatbotIDFromPath(p string) (string, bool) {
	const prefix = "/api/v1/chatbots/"
	if !strings.HasPrefix(p, prefix) {
		return "", false
	}
	rest := strings.TrimPrefix(p, prefix)
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[1] != "sources" || strings.TrimSpace(parts[0]) == "" {
		return "", false
	}
	return parts[0], true
}

// parseSourceIDFromPath extracts source ID from /api/v1/sources/:id
func parseSourceIDFromPath(p string) (string, bool) {
	const prefix = "/api/v1/sources/"
	if !strings.HasPrefix(p, prefix) {
		return "", false
	}
	sourceID := strings.TrimPrefix(p, prefix)
	// Ensure no trailing paths like /refresh
	if strings.Contains(sourceID, "/") || sourceID == "" {
		return "", false
	}
	return sourceID, true
}

// parseRefreshSourceIDFromPath extracts source ID from /api/v1/sources/:id/refresh
func parseRefreshSourceIDFromPath(p string) (string, bool) {
	const prefix = "/api/v1/sources/"
	const suffix = "/refresh"
	if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, suffix) {
		return "", false
	}
	sourceID := strings.TrimSuffix(strings.TrimPrefix(p, prefix), suffix)
	if sourceID == "" {
		return "", false
	}
	return sourceID, true
}

// isPDFContentType checks if content type or filename indicates PDF
func isPDFContentType(ct, name string) bool {
	if ct == "application/pdf" {
		return true
	}
	return strings.HasSuffix(name, ".pdf")
}

// parseSitemapDiscoverPath extracts chatbot ID from /api/v1/chatbots/:id/sitemap/discover
func parseSitemapDiscoverPath(p string) (string, bool) {
	const prefix = "/api/v1/chatbots/"
	const suffix = "/sitemap/discover"
	if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, suffix) {
		return "", false
	}
	chatbotID := strings.TrimSuffix(strings.TrimPrefix(p, prefix), suffix)
	if chatbotID == "" || strings.Contains(chatbotID, "/") {
		return "", false
	}
	return chatbotID, true
}

// parseBulkSourcesPath extracts chatbot ID from /api/v1/chatbots/:id/sources/bulk
func parseBulkSourcesPath(p string) (string, bool) {
	const prefix = "/api/v1/chatbots/"
	const suffix = "/sources/bulk"
	if !strings.HasPrefix(p, prefix) || !strings.HasSuffix(p, suffix) {
		return "", false
	}
	chatbotID := strings.TrimSuffix(strings.TrimPrefix(p, prefix), suffix)
	if chatbotID == "" || strings.Contains(chatbotID, "/") {
		return "", false
	}
	return chatbotID, true
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
