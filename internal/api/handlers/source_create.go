package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
	"github.com/onurceri/botla-co/pkg/urlutil"
)

var ssrfValidator = urlutil.NewSSRFValidator()

// createSource handles POST request to create a new source
func (h *SourcesHandlers) createSource(w http.ResponseWriter, r *http.Request, bot *models.Chatbot, userID string) {
	plan, err := db.GetPlanByUserID(r.Context(), h.DB, userID)
	if err != nil || plan == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Check monthly ingestion quota
	if err = h.checkIngestionQuota(r, userID, plan); err != nil {
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
		h.handlePDFUpload(w, r, bot, plan)
	case "url":
		h.handleURLSource(w, r, bot.ID, plan)
	case "text":
		h.handleTextSource(w, r, bot, userID, plan)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

// handlePDFUpload handles PDF file upload
func (h *SourcesHandlers) handlePDFUpload(w http.ResponseWriter, r *http.Request, bot *models.Chatbot, plan *models.Plan) {
	// Check file count limit
	cnt, err := db.CountSourcesByType(r.Context(), h.DB, bot.ID, "pdf")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	limit := plan.Config.Files.MaxFilesPerBot
	if limit <= 0 {
		limit = 5 // Safe fallback
	}
	if cnt >= limit {
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrPdfLimitReached)
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
		api.WriteErrorCode(w, http.StatusRequestEntityTooLarge, api.ErrFileTooLarge)
		return
	}

	// Check total storage quota
	userID, _ := middleware.UserIDFromContext(r.Context())
	if err = h.checkStorageQuota(r, userID, int(header.Size), plan); err != nil {
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
		h.logError("storage_missing", map[string]any{"chatbot_id": bot.ID, "path": r.URL.Path})
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Read file into memory to compute hash and then upload
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(file); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hval := computeHash(buf.Bytes())

	// Check for duplicate content
	exists, err := db.SourceExistsByHash(r.Context(), h.DB, bot.ID, hval)
	if err != nil {
		h.logError("hash_check_failed", map[string]any{"error": err.Error(), "chatbot_id": bot.ID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		api.WriteErrorCode(w, http.StatusConflict, api.ErrDuplicateContent)
		return
	}

	key := generateSourceStorageKey(bot, header.Filename)
	uploadedKey, err := h.Storage.UploadFile(r.Context(), key, bytes.NewReader(buf.Bytes()))
	if err != nil {
		h.logError("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": bot.ID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ds := models.DataSource{
		ChatbotID:        bot.ID,
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
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrURLLimitReached)
		return
	}

	rawURL := strings.TrimSpace(r.FormValue("source_url"))
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// NEW: SSRF validation
	if err = ssrfValidator.ValidateURL(rawURL); err != nil {
		h.logWarn("ssrf_blocked", map[string]any{
			"url":    rawURL,
			"reason": err.Error(),
		})
		api.WriteErrorCode(w, http.StatusForbidden, api.ErrBlockedURL)
		return
	}

	// Normalize URL to prevent duplicates with trailing slash variations
	url, err := urlutil.NormalizeURL(rawURL)
	if err != nil || url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check cooldown after delete
	lastDel, _ := db.GetLastDeletedAtForURL(r.Context(), h.DB, chatbotID, url)
	if lastDel.Valid {
		if remaining, ok := h.checkCooldown(r, &lastDel.Time, plan); !ok {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", remaining.Seconds()))
			api.WriteErrorCode(w, http.StatusTooManyRequests, api.ErrReaddCooldownActive)
			return
		}
	}

	if exists, _ := db.SourceExists(r.Context(), h.DB, chatbotID, url); exists {
		api.WriteErrorCode(w, http.StatusConflict, api.ErrDuplicateURL)
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
func (h *SourcesHandlers) handleTextSource(w http.ResponseWriter, r *http.Request, bot *models.Chatbot, userID string, plan *models.Plan) {
	text := strings.TrimSpace(r.FormValue("text"))
	if text == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check text length limit
	limit := plan.Config.Files.MaxTextLength
	if limit <= 0 {
		limit = 400000 // Safe fallback
	}
	if len(text) > limit {
		api.WriteErrorCode(w, http.StatusRequestEntityTooLarge, api.ErrTextTooLong)
		return
	}

	if h.Storage == nil {
		h.logError("storage_missing", map[string]any{"chatbot_id": bot.ID, "path": r.URL.Path})
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	// Check total storage quota for inline text
	if err := h.checkStorageQuota(r, userID, len(text), plan); err != nil {
		http.Error(w, err.Error(), http.StatusPaymentRequired)
		return
	}

	hval := computeHash([]byte(text))

	// Check for duplicate content
	exists, err := db.SourceExistsByHash(r.Context(), h.DB, bot.ID, hval)
	if err != nil {
		h.logError("hash_check_failed", map[string]any{"error": err.Error(), "chatbot_id": bot.ID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if exists {
		api.WriteErrorCode(w, http.StatusConflict, api.ErrDuplicateContent)
		return
	}

	key := generateSourceStorageKey(bot, "inline.txt")
	uploadedKey, err := h.Storage.UploadFile(r.Context(), key, bytes.NewBufferString(text))
	if err != nil {
		h.logError("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": bot.ID})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	of := "inline.txt"
	ds := models.DataSource{
		ChatbotID:        bot.ID,
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
	newID, err := h.persistAndEnqueueInternal(r, ds)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	api.WriteJSON(w, http.StatusCreated, map[string]string{"id": newID})
}

// generateSourceStorageKey creates a hierarchical R2 key for source files.
// Uses org/ws/bot path structure when IDs are available.
func generateSourceStorageKey(bot *models.Chatbot, filename string) string {
	if bot.OrganizationID != nil && bot.WorkspaceID != nil {
		return storage.GenerateSourceKey(*bot.OrganizationID, *bot.WorkspaceID, bot.ID, filename)
	}
	// Fallback for legacy bots without org/ws (shouldn't happen with current system)
	return storage.GenerateKey("sources", filename)
}
