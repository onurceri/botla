package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func TestGetSourceChunks_Success(t *testing.T) {
	// 1. Setup DB
	db := testdb.OpenParallelTestDB(t)

	// Create user, plan, chatbot, source
	userID := uuid.New().String()
	chatbotID := uuid.New().String()
	sourceID := uuid.New().String()

	// Insert User (and Plan/Lang via migrations or seeding if needed, but lets try raw insert)
	// Note: Migrations usually seed plans/languages.

	_, err := db.Exec(`INSERT INTO users (id, email, password_hash, plan_id) VALUES ($1, 'test@example.com', 'hash', (SELECT id FROM plans LIMIT 1))`, userID)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}

	_, err = db.Exec(`INSERT INTO chatbots (id, user_id, name) VALUES ($1, $2, 'Test Bot')`, chatbotID, userID)
	if err != nil {
		t.Fatalf("insert chatbot: %v", err)
	}

	_, err = db.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url) VALUES ($1, $2, 'web', 'http://example.com')`, sourceID, chatbotID)
	if err != nil {
		t.Fatalf("insert source: %v", err)
	}

	// 2. Mock Qdrant
	qSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/collections/embeddings/points/scroll" {
			w.Header().Set("Content-Type", "application/json")
			items := []rag.SearchResult{
				{
					ID:    "chunk-1",
					Score: 0.0,
					Payload: rag.EmbeddingPayload{
						SourceID:     sourceID,
						OriginalText: "This is chunk 1 content.",
						ChunkIndex:   0,
					},
				},
			}
			resp := map[string]any{
				"status": "ok",
				"result": map[string]any{
					"points":           items,
					"next_page_offset": nil,
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer qSrv.Close()
	t.Setenv("QDRANT_URL", qSrv.URL)
	qc, _ := rag.NewQdrantClientFromEnv()

	// 3. Init Handler
	sh := &SourcesHandlers{DB: db, QdrantClient: qc}

	// 4. Request
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/sources/%s/chunks?chatbot_id=%s", sourceID, chatbotID), nil)
	req.SetPathValue("id", sourceID)

	// Inject UserID into context (simulating AuthMiddleware)
	ctx := context.WithValue(req.Context(), middleware.ContextKeyUserID, userID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Call Handler directly
	sh.GetSourceChunks(rr, req)

	// 5. Assert
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v. Body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}

	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	if chunks, ok := resp["chunks"].([]any); !ok || len(chunks) != 1 {
		t.Errorf("expected 1 chunk, got %v", resp)
	}
}
