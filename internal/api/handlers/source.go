package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/processing"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/onurceri/botla-co/pkg/middleware"
	"github.com/onurceri/botla-co/pkg/storage"
)

type SourcesHandlers struct {
	DB      *sql.DB
	Queue   *processing.SourceQueue
	Storage storage.StorageService
	Log     *logger.Logger
}

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
	var err error
	c, err := db.GetChatbotByID(r.Context(), h.DB, chatbotID)
	if err != nil {
		if h.Log != nil {
			h.Log.Error("chatbot_fetch_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
		}
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
		items, err := db.ListSourcesByChatbotID(r.Context(), h.DB, chatbotID)
		if err != nil {
			if h.Log != nil {
				h.Log.Error("sources_list_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "path": r.URL.Path})
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(items); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	case http.MethodPost:
		if err = r.ParseMultipartForm(52 << 20); err != nil { // ~52MB
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sourceType := strings.TrimSpace(r.FormValue("source_type"))
		if sourceType == "" {
			sourceType = "pdf"
		}
		var ds models.DataSource
		ds.ChatbotID = chatbotID
		ds.SourceType = sourceType
		ds.Status = "pending"

		switch sourceType {
		case "pdf":
			file, header, err := r.FormFile("file")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer func() { _ = file.Close() }()
			if header.Size > 50<<20 { // 50MB limit
				w.WriteHeader(http.StatusRequestEntityTooLarge)
				return
			}
			ct := header.Header.Get("Content-Type")
			name := strings.ToLower(header.Filename)
			if !isPDFContentType(ct, name) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if h.Storage == nil {
				if h.Log != nil {
					h.Log.Error("storage_missing", map[string]any{"chatbot_id": chatbotID, "path": r.URL.Path})
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			key := storage.GenerateKey("sources", header.Filename)
			uploadedKey, err := h.Storage.UploadFile(r.Context(), key, file)
			if err != nil {
				if h.Log != nil {
					h.Log.Error("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": chatbotID})
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ds.FilePath = &uploadedKey
			orig := header.Filename
			ds.OriginalFilename = &orig
		case "url":
			url := strings.TrimSpace(r.FormValue("source_url"))
			if url == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			ds.SourceURL = &url
		case "text":
			text := strings.TrimSpace(r.FormValue("text"))
			if text == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if h.Storage == nil {
				if h.Log != nil {
					h.Log.Error("storage_missing", map[string]any{"chatbot_id": chatbotID, "path": r.URL.Path})
				}
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			key := storage.GenerateKey("sources", "inline.txt")
			uploadedKey, err := h.Storage.UploadFile(r.Context(), key, bytes.NewBufferString(text))
			if err != nil {
				if h.Log != nil {
					h.Log.Error("storage_upload_error", map[string]any{"error": err.Error(), "key": key, "chatbot_id": chatbotID})
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			ds.FilePath = &uploadedKey
			of := "inline.txt"
			ds.OriginalFilename = &of
		default:
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newID, err := db.CreateDataSource(r.Context(), h.DB, &ds)
		if err != nil {
			if h.Log != nil {
				h.Log.Error("source_create_error", map[string]any{"error": err.Error(), "chatbot_id": chatbotID, "source_type": sourceType})
			}
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

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

func isPDFContentType(ct, name string) bool {
	if ct == "application/pdf" {
		return true
	}
	return strings.HasSuffix(name, ".pdf")
}

func (h *SourcesHandlers) GetSourceStatusOrDelete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	const prefix = "/api/v1/sources/"
	path := r.URL.Path
	if !strings.HasPrefix(path, prefix) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	sourceID := strings.TrimPrefix(path, prefix)
	if sourceID == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var err error
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(s); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	case http.MethodDelete:
		// Best-effort: delete associated vectors then remove source record
		if err = processing.DeleteSourceVectors(r.Context(), s.ID); err != nil {
			if h.Log != nil {
				h.Log.Warn("vector_delete_error", map[string]any{"source_id": s.ID, "error": err.Error()})
			}
		}
		// Also delete from storage if it's a file
		if s.FilePath != nil && h.Storage != nil {
			_ = h.Storage.DeleteFile(r.Context(), *s.FilePath)
		}

		if err = db.DeleteSource(r.Context(), h.DB, s.ID); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
