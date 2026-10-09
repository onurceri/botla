package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/api"
	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/internal/services"
)

const testHandoffRequestID = "22222222-2222-2222-2222-222222222222"

func newContactHandler(handoffRepo *repository.MockHandoffRepo) *HandoffHandlers {
	return &HandoffHandlers{
		ChatbotRepo:    botRepoWithVisibility(true),
		HandoffRepo:    handoffRepo,
		HandoffService: services.NewHandoffService(handoffRepo, nil, nil, nil),
	}
}

func contactPath(botID, requestID string) string {
	return "/api/v1/public/chatbots/" + botID + "/handoff/" + requestID + "/contact"
}

func TestPublicSubmitEmail_PersistsEmail(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	repo.GetHandoffRequestByIDFunc = func(_ context.Context, id string) (*models.HandoffRequest, error) {
		return &models.HandoffRequest{ID: id, ChatbotID: testIsPublicBotID, Status: models.HandoffStatusPending}, nil
	}
	h := newContactHandler(repo)

	body := []byte(`{"email":" visitor@example.com "}`)
	req := httptest.NewRequest(http.MethodPost, contactPath(testIsPublicBotID, testHandoffRequestID), bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.PublicSubmitEmail(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(repo.Calls.SetHandoffContactEmail) != 1 {
		t.Fatalf("expected email to be persisted once, got %d calls", len(repo.Calls.SetHandoffContactEmail))
	}
	call := repo.Calls.SetHandoffContactEmail[0]
	if call.ID != testHandoffRequestID || call.ChatbotID != testIsPublicBotID || call.Email != "visitor@example.com" {
		t.Fatalf("unexpected persisted values: %+v", call)
	}
	// The old implementation reset the status instead of storing the email.
	if len(repo.Calls.UpdateHandoffRequestStatus) != 0 {
		t.Fatalf("status must not be touched, got %+v", repo.Calls.UpdateHandoffRequestStatus)
	}
}

func TestPublicSubmitEmail_Errors(t *testing.T) {
	pending := func(chatbotID, status string) func(context.Context, string) (*models.HandoffRequest, error) {
		return func(_ context.Context, id string) (*models.HandoffRequest, error) {
			return &models.HandoffRequest{ID: id, ChatbotID: chatbotID, Status: status}, nil
		}
	}
	tests := []struct {
		name      string
		path      string
		body      string
		getFn     func(context.Context, string) (*models.HandoffRequest, error)
		wantCode  int
		wantError string
	}{
		{name: "invalid email", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{"email":"not-an-email"}`, wantCode: http.StatusBadRequest, wantError: api.ErrInvalidEmailFormat},
		{name: "display name email", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{"email":"Bob <bob@example.com>"}`, wantCode: http.StatusBadRequest, wantError: api.ErrInvalidEmailFormat},
		{name: "bad json", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{`, wantCode: http.StatusBadRequest, wantError: api.ErrInvalidRequestBody},
		{name: "invalid request id", path: contactPath(testIsPublicBotID, "not-a-uuid"), body: `{"email":"a@example.com"}`, wantCode: http.StatusBadRequest, wantError: api.ErrCodeBadRequest},
		{name: "unknown request", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{"email":"a@example.com"}`, getFn: func(context.Context, string) (*models.HandoffRequest, error) { return nil, nil }, wantCode: http.StatusNotFound, wantError: api.ErrHandoffNotFound},
		{name: "request of another bot", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{"email":"a@example.com"}`, getFn: pending("33333333-3333-3333-3333-333333333333", models.HandoffStatusPending), wantCode: http.StatusNotFound, wantError: api.ErrHandoffNotFound},
		{name: "resolved request", path: contactPath(testIsPublicBotID, testHandoffRequestID), body: `{"email":"a@example.com"}`, getFn: pending(testIsPublicBotID, models.HandoffStatusResolved), wantCode: http.StatusConflict, wantError: api.ErrHandoffClosed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repository.NewMockHandoffRepo()
			repo.GetHandoffRequestByIDFunc = tt.getFn
			h := newContactHandler(repo)

			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewReader([]byte(tt.body)))
			rr := httptest.NewRecorder()
			h.PublicSubmitEmail(rr, req)

			if rr.Code != tt.wantCode {
				t.Fatalf("expected %d, got %d: %s", tt.wantCode, rr.Code, rr.Body.String())
			}
			var resp api.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if resp.Code != tt.wantError {
				t.Fatalf("expected code %s, got %s", tt.wantError, resp.Code)
			}
			if len(repo.Calls.SetHandoffContactEmail) != 0 {
				t.Fatalf("email must not be persisted on error")
			}
		})
	}
}
