package integration

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockIngestionFlow(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	mockLLM := &rag.MockFullClient{}
	mockVC := &rag.MockVectorClient{}
	mockVC.On("EnsureEmbeddingsCollection", mock.Anything).Return(nil)

	// Setup Mux with mocks
	mux, q, rl, wp, _, _ := fixtures.NewTestMux(te.Cfg, te.DB, te.VectorStore, mockLLM, mockVC)
	if q != nil {
		defer q.Stop()
	}
	if rl != nil {
		defer rl.Close()
	}
	if wp != nil {
		defer wp.Shutdown(1 * time.Second)
	}
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// 1. Setup Mock Expectations

	// Metadata extraction mock
	mockLLM.On("CreateCompletion", mock.Anything, mock.Anything).Return(&models.CompletionResult{
		Content: `{"capability_summary": "Mock capability", "suggested_questions": ["What is mock?"]}`,
	}, nil)

	// Embedding mock
	mockLLM.On("CreateEmbeddingsBatch", mock.Anything, mock.Anything).Return([][]float32{{0.1, 0.2}}, nil)

	// Vector DB mock
	mockVC.On("UpsertEmbedding", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	token := authToken(t, ts.URL, "mockingest@example.com")

	// 2. Create Chatbot
	create := map[string]any{"name": "Mock Ingest Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots", strings.NewReader(string(cbj)))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// 3. Trigger Ingestion (Text)
	var b strings.Builder
	mw := multipart.NewWriter(&b)
	mw.WriteField("source_type", "text")
	mw.WriteField("text", "This is some mock text to ingest.")
	mw.Close()

	reqS, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/chatbots/"+bot.ID+"/sources", strings.NewReader(b.String()))
	reqS.Header.Set("Authorization", "Bearer "+token)
	reqS.Header.Set("Content-Type", mw.FormDataContentType())
	resS, _ := testHTTPClient().Do(reqS)
	assert.Equal(t, http.StatusCreated, resS.StatusCode)

	var sidResp map[string]string
	json.NewDecoder(resS.Body).Decode(&sidResp)
	resS.Body.Close()
	sourceID := sidResp["id"]

	// 4. Wait for processing to complete
	completed := false
	for i := 0; i < 20; i++ {
		reqG, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/sources/"+sourceID, nil)
		reqG.Header.Set("Authorization", "Bearer "+token)
		resG, _ := testHTTPClient().Do(reqG)
		if resG.StatusCode == http.StatusOK {
			var st map[string]any
			json.NewDecoder(resG.Body).Decode(&st)
			resG.Body.Close()
			if st["status"] == "completed" {
				completed = true
				break
			}
		} else {
			resG.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	assert.True(t, completed, "Source processing should complete")
	mockLLM.AssertExpectations(t)
	mockVC.AssertExpectations(t)
}
