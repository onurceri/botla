package handlers

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

// SourcesHandlers handles all source-related HTTP endpoints
type SourcesHandlers struct {
	DB      *sql.DB
	Queue   *processing.SourceQueue
	Storage storage.StorageService
	Log     *logger.Logger
}

// ChatbotSources routes GET/POST requests for chatbot sources
func (h *SourcesHandlers) ChatbotSources(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	chatbotID, ok := parseChatbotIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if chatbotID == "new" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("chatbot_fetch_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.listSources(w, r, chatbotID)
	case http.MethodPost:
		h.createSource(w, r, chatbotID, userID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listSources handles GET request to list all sources for a chatbot
func (h *SourcesHandlers) listSources(w http.ResponseWriter, r *http.Request, chatbotID string) {
	items, err := db.ListSourcesByChatbotID(r.Context(), h.DB, chatbotID)
	if err != nil {
		h.logError("sources_list_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(items); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// createSource handles POST request to create a new source
func (h *SourcesHandlers) createSource(w http.ResponseWriter, r *http.Request, chatbotID, userID string) {
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check monthly ingestion quota
	if err := h.checkIngestionQuota(r, userID, plan); err != nil {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	}

	if err = r.ParseMultipartForm(52 << 20); err != nil { // ~52MB
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sourceType := strings.TrimSpace(r.FormValue("source_type"))
	if sourceType == "" {
		sourceType = "pdf"
	}

	switch sourceType {
	case "pdf":
		h.handlePDFUpload(w, r, chatbotID, plan)
	case "url":
		h.handleURLSource(w, r, chatbotID, plan)
	case "text":
		h.handleTextSource(w, r, chatbotID, userID, plan)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

// handlePDFUpload handles PDF file upload
func (h *SourcesHandlers) handlePDFUpload(w http.ResponseWriter, r *http.Request, chatbotID string, plan *models.Plan) {
	// Check file count limit
	cnt, err := db.CountSourcesByType(r.Context(), h.DB, chatbotID, "pdf")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	limit := plan.Config.Files.MaxFilesPerBot
	if limit <= 0 {
		limit = 5 // Safe fallback
	}
	if cnt >= limit {
		http.Error(w, "Limit reached: Max PDF files per chatbot", http.StatusForbidden)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer func() { _ = file.Close() }()

	// Check file size limit
	maxSizeMB := plan.Config.Files.MaxSizeMB
	if maxSizeMB <= 0 {
		maxSizeMB = 10 // Safe fallback
	}
	if header.Size > int64(maxSizeMB)<<20 {
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}

	// Check total storage quota
	userID, _ := middleware.UserIDFromContext(r.Context())
	if err := h.checkStorageQuota(r, userID, int(header.Size), plan); err != nil {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	}

	ct := header.Header.Get("Content-Type")
	name := strings.ToLower(header.Filename)
	if !isPDFContentType(ct, name) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.Storage == nil {
		h.logError("storage_missing", map[string]any{"chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Read file into memory to compute hash and then upload
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hval := computeHash(buf.Bytes())

	key := storage.GenerateKey("sources", header.Filename)
	uploadedKey, err := h.Storage.UploadFile(r.Context(), key, bytes.NewReader(buf.Bytes()))
	if err != nil {
		h.logError("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": chatbotID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ds := models.DataSource{
		ChatbotID:        chatbotID,
		SourceType:       "pdf",
		Status:           "pending",
		Hash:             &hval,
		FilePath:         &uploadedKey,
		OriginalFilename: &header.Filename,
		SizeBytes:        header.Size,
	}
	h.persistAndEnqueue(w, r, &ds)
}

// handleURLSource handles URL source creation
func (h *SourcesHandlers) handleURLSource(w http.ResponseWriter, r *http.Request, chatbotID string, plan *models.Plan) {
	// Check URL count limit
	cnt, err := db.CountSourcesByType(r.Context(), h.DB, chatbotID, "url")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	limit := plan.Config.Scraping.MaxURLsPerBot
	if limit <= 0 {
		limit = 5 // Safe fallback
	}
	if cnt >= limit {
		http.Error(w, "Limit reached: Max URLs per chatbot", http.StatusForbidden)
		return
	}

	url := strings.TrimSpace(r.FormValue("source_url"))
	if url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check cooldown after delete
	cdMin := plan.Config.MinReAddCooldownMinutes
	if cdMin > 0 {
		lastDel, _ := db.GetLastDeletedAtForURL(r.Context(), h.DB, chatbotID, url)
		if lastDel.Valid {
			if time.Since(lastDel.Time) < time.Duration(cdMin)*time.Minute {
				http.Error(w, "Re-add cooldown active", http.StatusTooManyRequests)
				return
			}
		}
	}

	if exists, _ := db.SourceExists(r.Context(), h.DB, chatbotID, url); exists {
		http.Error(w, "Duplicate URL", http.StatusConflict)
		return
	}

	ds := models.DataSource{
		ChatbotID:  chatbotID,
		SourceType: "url",
		Status:     "pending",
		SourceURL:  &url,
	}
	h.persistAndEnqueue(w, r, &ds)
}

// handleTextSource handles inline text source creation
func (h *SourcesHandlers) handleTextSource(w http.ResponseWriter, r *http.Request, chatbotID, userID string, plan *models.Plan) {
	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if h.Storage == nil {
		h.logError("storage_missing", map[string]any{"chatbot_id": chatbotID, "path": r.URL.Path})
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Check total storage quota for inline text
	if err := h.checkStorageQuota(r, userID, len(text), plan); err != nil {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	}

	hval := computeHash([]byte(text))
	key := storage.GenerateKey("sources", "inline.txt")
	uploadedKey, err := h.Storage.UploadFile(r.Context(), key, bytes.NewBufferString(text))
	if err != nil {
		h.logError("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": chatbotID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	of := "inline.txt"
	ds := models.DataSource{
		ChatbotID:        chatbotID,
		SourceType:       "text",
		Status:           "pending",
		Hash:             &hval,
		FilePath:         &uploadedKey,
		OriginalFilename: &of,
		SizeBytes:        int64(len(text)),
	}
	h.persistAndEnqueue(w, r, &ds)
}

// persistAndEnqueue saves the data source to database and enqueues for processing
func (h *SourcesHandlers) persistAndEnqueue(w http.ResponseWriter, r *http.Request, ds *models.DataSource) {
	newID, err := db.CreateDataSource(r.Context(), h.DB, ds)
	if err != nil {
		h.logError("source_create_error", map[string]any{"error": err.Error(), "chatbot_id": ds.ChatbotID, "source_type": ds.SourceType})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if h.Queue != nil {
		h.Queue.Enqueue(newID)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(map[string]string{"id": newID}); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// GetSourceStatusOrDelete handles GET/DELETE for individual sources
func (h *SourcesHandlers) GetSourceStatusOrDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sourceID, ok := parseSourceIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	c, err := db.GetChatbotByID(r.Context(), h.DB, s.ChatbotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getSourceStatus(w, r, s)
	case http.MethodDelete:
		h.deleteSource(w, r, s)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// getSourceStatus returns source status with ETag support
func (h *SourcesHandlers) getSourceStatus(w http.ResponseWriter, r *http.Request, s *models.DataSource) {
	// Compute ETag from status + processed_at + chunk_count
	etag := s.Status
	if s.ProcessedAt != nil {
		etag += "-" + s.ProcessedAt.UTC().Format(time.RFC3339Nano)
	}
	etag += "-" + strconv.Itoa(s.ChunkCount)

	inm := r.Header.Get("If-None-Match")
	if inm != "" && inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "private, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(s); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// deleteSource handles source deletion
func (h *SourcesHandlers) deleteSource(w http.ResponseWriter, r *http.Request, s *models.DataSource) {
	// Best-effort: delete associated vectors then remove source record
	if err := processing.DeleteSourceVectors(r.Context(), s.ID); err != nil {
		h.logWarn("vector_delete_error", map[string]any{"source_id": s.ID, "error": err.Error()})
	}

	// Also delete from storage if it's a file
	if s.FilePath != nil && h.Storage != nil {
		_ = h.Storage.DeleteFile(r.Context(), *s.FilePath)
	}

	if err := db.SoftDeleteSource(r.Context(), h.DB, s.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RefreshSource handles POST /api/v1/sources/:id/refresh
func (h *SourcesHandlers) RefreshSource(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	sourceID, ok := parseRefreshSourceIDFromPath(r.URL.Path)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s, err := db.GetSourceByID(r.Context(), h.DB, sourceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if s == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Verify ownership
	c, err := db.GetChatbotByID(r.Context(), h.DB, s.ChatbotID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if c == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if c.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Only URL sources can be refreshed
	if s.SourceType != "url" {
		http.Error(w, "Only URL sources can be refreshed", http.StatusBadRequest)
		return
	}

	// Check if source is already processing
	if s.Status == "pending" || s.Status == "processing" {
		http.Error(w, "Source is already being processed", http.StatusConflict)
		return
	}

	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check if refresh is enabled for this plan
	if !plan.Config.Refresh.Enabled {
		http.Error(w, "Refresh feature is not available on your plan", http.StatusForbidden)
		return
	}

	// Check monthly refresh quota
	usedRefreshes, _ := db.GetMonthlyRefreshCount(r.Context(), h.DB, userID, time.Now())
	if plan.Config.Refresh.MaxMonthly > 0 && usedRefreshes >= plan.Config.Refresh.MaxMonthly {
		http.Error(w, "Monthly refresh limit exceeded", http.StatusPaymentRequired)
		return
	}

	// Check cooldown
	cooldownMin := plan.Config.MinReAddCooldownMinutes
	if cooldownMin > 0 && s.LastRefreshedAt != nil {
		elapsed := time.Since(*s.LastRefreshedAt)
		if elapsed < time.Duration(cooldownMin)*time.Minute {
			remaining := time.Duration(cooldownMin)*time.Minute - elapsed
			w.Header().Set("Retry-After", strconv.Itoa(int(remaining.Seconds())))
			http.Error(w, "Refresh cooldown active", http.StatusTooManyRequests)
			return
		}
	}

	// Update source for refresh
	if err := db.UpdateSourceForRefresh(r.Context(), h.DB, sourceID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Increment refresh count
	_ = db.IncrementRefreshCount(r.Context(), h.DB, userID, time.Now())

	// Enqueue for processing
	if h.Queue != nil {
		h.Queue.Enqueue(sourceID)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(map[string]string{"id": sourceID})
}

// --- Helper functions ---

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
