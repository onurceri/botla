package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/onurceri/botla-co/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler(t *testing.T) (*SuggestionsHandlers, *sql.DB, func()) {
	dbConn, err := sql.Open("postgres", "postgres://botla:botla@localhost/botla_test?sslmode=disable")
	if err != nil {
		t.Skipf("skipping test: could not connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := dbConn.PingContext(ctx); err != nil {
		t.Skipf("skipping test: could not ping database: %v", err)
	}

	log := logger.New("info")

	h := &SuggestionsHandlers{
		DB:               dbConn,
		Log:              log,
		WorkspaceService: nil,
		OrgService:       nil,
	}

	teardown := func() {
		dbConn.Close()
	}

	return h, dbConn, teardown
}

func TestRegenerateSuggestions_Success(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, http.StatusAccepted, rec.Code)

	var response RegenerateResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotEmpty(t, response.JobID)

	job, err := db.GetSuggestionJob(ctx, dbConn, response.JobID)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, chatbotID, job.ChatbotID)
	assert.Equal(t, models.SuggestionJobStatusPending, job.Status)
}

func TestRegenerateSuggestions_MethodNotAllowed(t *testing.T) {
	h, _, teardown := setupTestHandler(t)
	defer teardown()

	result := testdb.CreateChatbot(t, nil)
	chatbotID := result.Chatbot.ID

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestGetSuggestionJobStatus_Success(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, dbConn, chatbotID)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, job.ID, response.JobID)
	assert.Equal(t, models.SuggestionJobStatusPending.String(), response.Status)
}

func TestGetSuggestionJobStatus_NotFound(t *testing.T) {
	h, _, teardown := setupTestHandler(t)
	defer teardown()

	chatbotID := uuid.NewString()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetSuggestionJobStatus_MethodNotAllowed(t *testing.T) {
	h, _, teardown := setupTestHandler(t)
	defer teardown()

	result := testdb.CreateChatbot(t, nil)
	chatbotID := result.Chatbot.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestGetSuggestionJobStatus_WithSuggestions(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, dbConn, chatbotID)
	assert.NoError(t, err)

	suggestions := []string{"Q1", "Q2", "Q3"}
	err = db.CompleteSuggestionJob(ctx, dbConn, job.ID, suggestions)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, models.SuggestionJobStatusCompleted.String(), response.Status)
	assert.Len(t, response.SuggestedQuestions, 1)
	assert.Len(t, response.SuggestedQuestions[0], 3)
}

func TestGetSuggestionJobStatus_WithError(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	job, err := db.CreateSuggestionJob(ctx, dbConn, chatbotID)
	assert.NoError(t, err)

	errMsg := "test error"
	err = db.FailSuggestionJob(ctx, dbConn, job.ID, errMsg)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response SuggestionJobStatusResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, models.SuggestionJobStatusFailed.String(), response.Status)
	assert.NotNil(t, response.ErrorMessage)
	assert.Equal(t, errMsg, *response.ErrorMessage)
}

func TestRegenerateSuggestions_InvalidChatbotID(t *testing.T) {
	h, _, teardown := setupTestHandler(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/invalid-uuid/suggestions/regenerate", nil)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetSuggestionJobStatus_InvalidChatbotID(t *testing.T) {
	h, _, teardown := setupTestHandler(t)
	defer teardown()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/invalid-uuid/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegenerateSuggestions_ReturnsJSON(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chatbots/"+chatbotID+"/suggestions/regenerate", nil)
	rec := httptest.NewRecorder()

	h.RegenerateSuggestions(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, ok := response["job_id"]
	assert.True(t, ok, "response should contain job_id")
}

func TestGetSuggestionJobStatus_ReturnsJSON(t *testing.T) {
	h, dbConn, teardown := setupTestHandler(t)
	defer teardown()

	ctx := context.Background()

	result := testdb.CreateChatbot(t, dbConn)
	chatbotID := result.Chatbot.ID

	_, err := db.CreateSuggestionJob(ctx, dbConn, chatbotID)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/chatbots/"+chatbotID+"/suggestions/status", nil)
	rec := httptest.NewRecorder()

	h.GetSuggestionJobStatus(rec, req)

	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)

	_, ok := response["job_id"]
	assert.True(t, ok, "response should contain job_id")
	_, ok = response["status"]
	assert.True(t, ok, "response should contain status")
}
