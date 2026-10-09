package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onurceri/botla-app/internal/models"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/middleware"
)

const otherBotID = "55555555-5555-5555-5555-555555555555"

func handoffItemRequest(method, body string) *http.Request {
	target := "/api/v1/chatbots/" + testIsPublicBotID + "/handoff-requests/" + testHandoffRequestID
	req := httptest.NewRequest(method, target, bytes.NewReader([]byte(body)))
	req.SetPathValue("id", testIsPublicBotID)
	return req.WithContext(context.WithValue(req.Context(), middleware.ContextKeyUserID, testHandoffOwnerID))
}

func TestGetHandoffRequestDetail_OtherChatbotIsNotFound(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	repo.GetHandoffRequestByIDFunc = func(_ context.Context, id string) (*models.HandoffRequest, error) {
		return &models.HandoffRequest{ID: id, ChatbotID: otherBotID, ConversationID: "c1"}, nil
	}
	h := newOwnedHandoffHandler(repo)

	rr := httptest.NewRecorder()
	h.GetHandoffRequestDetail(rr, handoffItemRequest(http.MethodGet, ""))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a request of another chatbot, got %d", rr.Code)
	}
	if len(repo.Calls.ListHandoffMessages) != 0 {
		t.Fatal("messages of another chatbot's conversation must not be loaded")
	}
}

func TestGetHandoffRequestDetail_OwnChatbot(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	repo.GetHandoffRequestByIDFunc = func(_ context.Context, id string) (*models.HandoffRequest, error) {
		return &models.HandoffRequest{ID: id, ChatbotID: testIsPublicBotID, ConversationID: "c1"}, nil
	}
	h := newOwnedHandoffHandler(repo)

	rr := httptest.NewRecorder()
	h.GetHandoffRequestDetail(rr, handoffItemRequest(http.MethodGet, ""))

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestUpdateHandoffRequest_ScopedToChatbot(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	h := newOwnedHandoffHandler(repo)

	rr := httptest.NewRecorder()
	h.UpdateHandoffRequest(rr, handoffItemRequest(http.MethodPatch, `{"status":"assigned"}`))

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rr.Code, rr.Body.String())
	}
	calls := repo.Calls.UpdateHandoffRequestStatus
	if len(calls) != 1 || calls[0].ChatbotID != testIsPublicBotID || calls[0].ID != testHandoffRequestID {
		t.Fatalf("update must be scoped to the path chatbot: %+v", calls)
	}
}

func TestUpdateHandoffRequest_OtherChatbotIsNotFound(t *testing.T) {
	repo := repository.NewMockHandoffRepo()
	// The scoped UPDATE matches no row when the request belongs to another chatbot.
	repo.UpdateHandoffRequestStatusFunc = func(context.Context, string, string, string, *string) error { return sql.ErrNoRows }
	h := newOwnedHandoffHandler(repo)

	rr := httptest.NewRecorder()
	h.UpdateHandoffRequest(rr, handoffItemRequest(http.MethodPatch, `{"status":"resolved"}`))

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
