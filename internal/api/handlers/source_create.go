package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

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
		h.logError("storage_missing", map[string]any{"chatbot_id": chatbotID, "path": r.URL.Path})
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
